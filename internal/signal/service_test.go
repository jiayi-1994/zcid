package signal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, row *HealthSignal) error {
	args := m.Called(ctx, row)
	if row.ID == "" {
		row.ID = "signal-1"
	}
	return args.Error(0)
}

func (m *mockRepo) ListLatestByTarget(ctx context.Context, projectID string, targetType TargetType, targetID string, limit int) ([]HealthSignal, error) {
	args := m.Called(ctx, projectID, targetType, targetID, limit)
	return args.Get(0).([]HealthSignal), args.Error(1)
}

func TestRecord_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	fixedNow := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixedNow }
	ctx := context.Background()
	staleAfter := fixedNow.Add(5 * time.Minute)

	repo.On("Create", ctx, mock.AnythingOfType("*signal.HealthSignal")).
		Return(nil).
		Run(func(args mock.Arguments) {
			row := args.Get(1).(*HealthSignal)
			assert.Equal(t, "proj-1", row.ProjectID)
			assert.Equal(t, TargetEnvironment, row.TargetType)
			assert.Equal(t, "env-1", row.TargetID)
			assert.Equal(t, "argocd", row.Source)
			assert.Equal(t, StatusHealthy, row.Status)
			assert.Equal(t, SeverityInfo, row.Severity)
			assert.Equal(t, fixedNow, row.ObservedAt)
			assert.Equal(t, &staleAfter, row.StaleAfter)
		})

	row, err := svc.Record(ctx, RecordInput{
		ProjectID:     " proj-1 ",
		TargetType:    TargetEnvironment,
		TargetID:      " env-1 ",
		Source:        " argocd ",
		Status:        StatusHealthy,
		ObservedValue: map[string]any{"sync": "Synced"},
		StaleAfter:    &staleAfter,
	})

	assert.NoError(t, err)
	assert.Equal(t, "signal-1", row.ID)
	repo.AssertExpectations(t)
}

func TestRecord_InvalidTarget(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.Record(context.Background(), RecordInput{
		ProjectID:  "proj-1",
		TargetType: TargetType("tenant"),
		TargetID:   "env-1",
		Source:     "probe",
		Status:     StatusHealthy,
	})

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
	repo.AssertNotCalled(t, "Create")
}

func TestRecord_InvalidStatus(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.Record(context.Background(), RecordInput{
		ProjectID:  "proj-1",
		TargetType: TargetService,
		TargetID:   "svc-1",
		Source:     "probe",
		Status:     Status("green"),
	})

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
	repo.AssertNotCalled(t, "Create")
}

func TestListLatestByTarget_EffectiveStatusMarksStale(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	fixedNow := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixedNow }
	ctx := context.Background()
	staleAfter := fixedNow.Add(-time.Minute)

	repo.On("ListLatestByTarget", ctx, "proj-1", TargetEnvironment, "env-1", 10).
		Return([]HealthSignal{{
			ID:            "signal-1",
			ProjectID:     "proj-1",
			TargetType:    TargetEnvironment,
			TargetID:      "env-1",
			Source:        "argocd",
			Status:        StatusHealthy,
			Severity:      SeverityInfo,
			ObservedValue: RawObject(),
			ObservedAt:    fixedNow.Add(-10 * time.Minute),
			StaleAfter:    &staleAfter,
		}}, nil)

	items, err := svc.ListLatestByTarget(ctx, " proj-1 ", TargetEnvironment, " env-1 ", 10)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, string(StatusHealthy), items[0].Status)
	assert.Equal(t, string(StatusStale), items[0].EffectiveStatus)
	repo.AssertExpectations(t)
}
