package logarchive

import "time"

// LogEntry represents a single log line for API responses.
type LogEntry struct {
	Seq       int64     `json:"seq"`
	StepID    string    `json:"stepId"`
	Content   string    `json:"content"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}
