package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogBufferAppendAndGetSince(t *testing.T) {
	b := NewLogBuffer(100)
	now := time.Now()
	b.Append(LogLine{StepID: "s1", Content: "line1", Level: "info", Timestamp: now})
	b.Append(LogLine{StepID: "s1", Content: "line2", Level: "info", Timestamp: now})
	b.Append(LogLine{StepID: "s1", Content: "line3", Level: "info", Timestamp: now})

	since := b.GetSince(0)
	require.Len(t, since, 3)
	assert.Equal(t, int64(1), since[0].Seq)
	assert.Equal(t, "line1", since[0].Content)
	assert.Equal(t, int64(3), since[2].Seq)
	assert.Equal(t, "line3", since[2].Content)

	since = b.GetSince(1)
	require.Len(t, since, 2)
	assert.Equal(t, int64(2), since[0].Seq)
	assert.Equal(t, int64(3), since[1].Seq)

	since = b.GetSince(3)
	assert.Len(t, since, 0)

	since = b.GetSince(100)
	assert.Len(t, since, 0)
}

func TestLogBufferMaxLines(t *testing.T) {
	b := NewLogBuffer(5)
	now := time.Now()
	for i := 0; i < 10; i++ {
		b.Append(LogLine{StepID: "s1", Content: "line", Level: "info", Timestamp: now})
	}
	lines := b.GetSince(0)
	assert.Len(t, lines, 5)
	assert.Equal(t, int64(6), lines[0].Seq)
	assert.Equal(t, int64(10), lines[4].Seq)
}

func TestGetSinceReplay(t *testing.T) {
	b := NewLogBuffer(100)
	now := time.Now()
	b.Append(LogLine{StepID: "s1", Content: "a", Level: "info", Timestamp: now})
	b.Append(LogLine{StepID: "s1", Content: "b", Level: "info", Timestamp: now})
	b.Append(LogLine{StepID: "s1", Content: "c", Level: "info", Timestamp: now})

	// Simulate reconnection: client had lastSeq=2, wants everything after
	replay := b.GetSince(2)
	require.Len(t, replay, 1)
	assert.Equal(t, int64(3), replay[0].Seq)
	assert.Equal(t, "c", replay[0].Content)
}
