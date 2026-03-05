package crdclean

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockArgo struct {
	sync   string
	health string
	err    error
}

func (m *mockArgo) GetAppStatus(ctx context.Context, appName string) (sync, health string, err error) {
	_ = ctx
	_ = appName
	return m.sync, m.health, m.err
}

func TestDriftDetector_CheckDrift_Synced(t *testing.T) {
	ctx := context.Background()
	detector := NewDriftDetector(&mockArgo{sync: "Synced", health: "Healthy"})
	report, err := detector.CheckDrift(ctx, "my-app")
	require.NoError(t, err)
	assert.False(t, report.Drifted)
	assert.Contains(t, report.Details, "Synced")
}

func TestDriftDetector_CheckDrift_Drifted(t *testing.T) {
	ctx := context.Background()
	detector := NewDriftDetector(&mockArgo{sync: "OutOfSync", health: "Degraded"})
	report, err := detector.CheckDrift(ctx, "my-app")
	require.NoError(t, err)
	assert.True(t, report.Drifted)
	assert.Contains(t, report.Details, "drifted")
}

func TestDriftDetector_CheckDrift_NilClient(t *testing.T) {
	ctx := context.Background()
	detector := NewDriftDetector(nil)
	report, err := detector.CheckDrift(ctx, "my-app")
	require.NoError(t, err)
	assert.False(t, report.Drifted)
}

func TestDriftDetector_CheckDrift_Error(t *testing.T) {
	ctx := context.Background()
	detector := NewDriftDetector(&mockArgo{err: errors.New("connection refused")})
	_, err := detector.CheckDrift(ctx, "my-app")
	require.Error(t, err)
}
