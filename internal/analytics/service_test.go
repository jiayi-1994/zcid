package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjy/zcid/pkg/response"
)

type fakeAnalyticsRepo struct {
	since time.Time
	err   error
	resp  *Response
}

func (f *fakeAnalyticsRepo) Get(ctx context.Context, projectID string, since time.Time) (*Response, error) {
	f.since = since
	if f.err != nil {
		return nil, f.err
	}
	if f.resp != nil {
		return f.resp, nil
	}
	return &Response{}, nil
}

func TestServiceGetAppliesDefaultRangeAndNormalizesSlices(t *testing.T) {
	repo := &fakeAnalyticsRepo{}
	start := time.Now()
	got, err := NewService(repo).Get(context.Background(), "p1", "")
	require.NoError(t, err)

	assert.Equal(t, "7d", got.Range)
	assert.NotNil(t, got.DailyStats)
	assert.NotNil(t, got.TopFailingSteps)
	assert.NotNil(t, got.TopPipelines)
	assert.WithinDuration(t, start.AddDate(0, 0, -7), repo.since, 2*time.Second)
}

func TestServiceGetAcceptsSupportedRanges(t *testing.T) {
	tests := []struct {
		rangeValue string
		days       int
	}{
		{rangeValue: "7d", days: 7},
		{rangeValue: "30d", days: 30},
		{rangeValue: "90d", days: 90},
	}
	for _, tt := range tests {
		t.Run(tt.rangeValue, func(t *testing.T) {
			repo := &fakeAnalyticsRepo{}
			start := time.Now()
			got, err := NewService(repo).Get(context.Background(), "p1", tt.rangeValue)
			require.NoError(t, err)

			assert.Equal(t, tt.rangeValue, got.Range)
			assert.WithinDuration(t, start.AddDate(0, 0, -tt.days), repo.since, 2*time.Second)
		})
	}
}

func TestServiceGetRejectsInvalidRange(t *testing.T) {
	_, err := NewService(&fakeAnalyticsRepo{}).Get(context.Background(), "p1", "14d")
	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeValidation, bizErr.Code)
}

func TestServiceGetWrapsRepoErrors(t *testing.T) {
	_, err := NewService(&fakeAnalyticsRepo{err: errors.New("db down")}).Get(context.Background(), "p1", "7d")
	require.Error(t, err)
	var bizErr *response.BizError
	require.True(t, errors.As(err, &bizErr))
	assert.Equal(t, response.CodeInternalServerError, bizErr.Code)
}
