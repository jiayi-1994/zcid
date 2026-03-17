package logarchive

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/xjy/zcid/internal/ws"
)

const (
	logPrefix   = "logs/"
	chunkSize   = 1024 * 1024 // 1MB
	contentType = "application/x-ndjson"
)

// Service archives and retrieves pipeline run logs from object storage.
type Service struct {
	storage StorageClient
	bucket  string
}

// NewService creates a LogArchiveService.
func NewService(storage StorageClient, bucket string) *Service {
	if bucket == "" {
		bucket = "zcid-logs"
	}
	return &Service{storage: storage, bucket: bucket}
}

// ArchiveRunLogs marshals logs to JSON Lines, splits into 1MB chunks, and uploads to storage.
func (s *Service) ArchiveRunLogs(ctx context.Context, runID string, logs []ws.LogLine) error {
	if len(logs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	var chunks [][]byte
	for _, line := range logs {
		buf.Reset()
		if err := enc.Encode(line); err != nil {
			return err
		}
		chunks = append(chunks, append([]byte(nil), buf.Bytes()...))
	}

	// Merge into 1MB chunks
	var current []byte
	chunkIdx := 0
	for _, b := range chunks {
		if len(current)+len(b) > chunkSize && len(current) > 0 {
			key := logPrefix + runID + "/chunk-" + strconv.Itoa(chunkIdx) + ".jsonl"
			if err := s.storage.PutObject(ctx, s.bucket, key, current, contentType); err != nil {
				return err
			}
			chunkIdx++
			current = nil
		}
		current = append(current, b...)
	}
	if len(current) > 0 {
		key := logPrefix + runID + "/chunk-" + strconv.Itoa(chunkIdx) + ".jsonl"
		if err := s.storage.PutObject(ctx, s.bucket, key, current, contentType); err != nil {
			return err
		}
	}
	return nil
}

// GetArchivedLogs returns a paginated slice of log entries and total count.
func (s *Service) GetArchivedLogs(ctx context.Context, runID string, page, pageSize int) ([]LogEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	prefix := logPrefix + runID + "/"
	keys, err := s.storage.ListObjects(ctx, s.bucket, prefix)
	if err != nil {
		return nil, 0, err
	}

	sort.Strings(keys)
	if len(keys) == 0 {
		return []LogEntry{}, 0, nil
	}

	var allEntries []LogEntry
	for _, key := range keys {
		data, err := s.storage.GetObject(ctx, s.bucket, key)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		scanner.Buffer(make([]byte, 0, chunkSize), chunkSize)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var ll ws.LogLine
			if err := json.Unmarshal([]byte(line), &ll); err != nil {
				continue
			}
			allEntries = append(allEntries, LogEntry{
				Seq:       ll.Seq,
				StepID:    ll.StepID,
				Content:   ll.Content,
				Level:     ll.Level,
				Timestamp: ll.Timestamp,
			})
		}
	}

	total := len(allEntries)
	offset := (page - 1) * pageSize
	if offset >= total {
		return []LogEntry{}, total, nil
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	return allEntries[offset:end], total, nil
}
