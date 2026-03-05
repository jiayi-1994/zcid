package crdclean

import (
	"context"
	"time"
)

type K8sClient interface {
	DeleteExpiredPipelineRuns(ctx context.Context, olderThan time.Duration) error
}

type MockK8sClient struct{}

func (m *MockK8sClient) DeleteExpiredPipelineRuns(ctx context.Context, olderThan time.Duration) error {
	_ = ctx
	_ = olderThan
	// TODO: Implement real K8s/Tekton PipelineRun deletion for runs older than TTL
	return nil
}

type CRDCleaner struct {
	K8sClient K8sClient
	TTLDays   int
}

func NewCRDCleaner(k8s K8sClient, ttlDays int) *CRDCleaner {
	if ttlDays <= 0 {
		ttlDays = 30
	}
	return &CRDCleaner{K8sClient: k8s, TTLDays: ttlDays}
}

func (c *CRDCleaner) CleanExpiredRuns(ctx context.Context) {
	ttl := time.Duration(c.TTLDays) * 24 * time.Hour
	_ = c.K8sClient.DeleteExpiredPipelineRuns(ctx, ttl)
}

func (c *CRDCleaner) StartScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.CleanExpiredRuns(ctx)
		}
	}
}
