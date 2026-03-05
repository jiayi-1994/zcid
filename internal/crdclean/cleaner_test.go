package crdclean

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCRDCleaner_CleanExpiredRuns(t *testing.T) {
	ctx := context.Background()
	called := false
	mock := &mockK8sDelete{
		deleteFn: func(ctx context.Context, olderThan time.Duration) error {
			called = true
			assert.Equal(t, 7*24*time.Hour, olderThan)
			return nil
		},
	}
	cleaner := NewCRDCleaner(mock, 7)
	cleaner.CleanExpiredRuns(ctx)
	assert.True(t, called)
}

func TestCRDCleaner_StartScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cleaner := NewCRDCleaner(&MockK8sClient{}, 30)
	go cleaner.StartScheduler(ctx, time.Hour)
	time.Sleep(20 * time.Millisecond)
}

type mockK8sDelete struct {
	deleteFn func(ctx context.Context, olderThan time.Duration) error
}

func (m *mockK8sDelete) DeleteExpiredPipelineRuns(ctx context.Context, olderThan time.Duration) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, olderThan)
	}
	return nil
}
