package logarchive

import (
	"context"
	"sync"
)

// MockStorage implements StorageClient for tests.
type MockStorage struct {
	mu    sync.Mutex
	store map[string][]byte
}

// NewMockStorage creates a mock storage.
func NewMockStorage() *MockStorage {
	return &MockStorage{store: make(map[string][]byte)}
}

// PutObject stores data.
func (m *MockStorage) PutObject(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[bucket+"/"+key] = append([]byte(nil), data...)
	return nil
}

// GetObject retrieves data.
func (m *MockStorage) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b, ok := m.store[bucket+"/"+key]; ok {
		return append([]byte(nil), b...), nil
	}
	return nil, context.DeadlineExceeded // signal not found for tests
}

// ListObjects returns keys under prefix.
func (m *MockStorage) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fullPrefix := bucket + "/" + prefix
	var keys []string
	for k := range m.store {
		if len(k) >= len(fullPrefix) && k[:len(fullPrefix)] == fullPrefix {
			keys = append(keys, k[len(bucket)+1:])
		}
	}
	return keys, nil
}
