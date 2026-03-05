package ws

import "time"

// Message types
const (
	MsgTypeLog       = "log"
	MsgTypeStatus    = "status"
	MsgTypeHeartbeat = "heartbeat"
	MsgTypeError     = "error"
)

// WSMessage is the envelope for all WebSocket messages
type WSMessage struct {
	Type      string      `json:"type"`
	Seq       int64       `json:"seq"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// LogData holds log line content
type LogData struct {
	Line   string `json:"line"`
	StepID string `json:"stepId"`
	Level  string `json:"level"`
}

// StatusData holds pipeline run status update
type StatusData struct {
	RunID        string       `json:"runId"`
	Status       string       `json:"status"`
	StepStatuses []StepStatus `json:"stepStatuses"`
}

// StepStatus holds a single step's status
type StepStatus struct {
	StepID     string     `json:"stepId"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}
