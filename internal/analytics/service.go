package analytics

import (
	"context"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Get(ctx context.Context, projectID string, rangeValue string) (*Response, error) {
	days := 7
	switch rangeValue {
	case "", "7d":
		rangeValue = "7d"
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		return nil, response.NewBizError(response.CodeValidation, "invalid range", "range must be one of 7d, 30d, 90d")
	}
	result, err := s.repo.Get(ctx, projectID, time.Now().AddDate(0, 0, -days))
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "failed to load analytics", err.Error())
	}
	result.Range = rangeValue
	if result.DailyStats == nil {
		result.DailyStats = []DailyStat{}
	}
	if result.TopFailingSteps == nil {
		result.TopFailingSteps = []TopFailingStep{}
	}
	if result.TopPipelines == nil {
		result.TopPipelines = []TopPipeline{}
	}
	return result, nil
}
