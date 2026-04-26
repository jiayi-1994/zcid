package signal

import (
	"context"
	"strings"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Record(ctx context.Context, input RecordInput) (*HealthSignal, error) {
	now := s.now()
	row := &HealthSignal{
		ProjectID:     strings.TrimSpace(input.ProjectID),
		TargetType:    input.TargetType,
		TargetID:      strings.TrimSpace(input.TargetID),
		Source:        strings.TrimSpace(input.Source),
		Status:        input.Status,
		Severity:      input.Severity,
		Reason:        strings.TrimSpace(input.Reason),
		Message:       strings.TrimSpace(input.Message),
		ObservedValue: NewJSONRaw(input.ObservedValue),
		ObservedAt:    input.ObservedAt,
		StaleAfter:    input.StaleAfter,
	}
	row.Normalize(now)

	if err := validateSignal(row); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) ListLatestByTarget(ctx context.Context, projectID string, targetType TargetType, targetID string, limit int) ([]HealthSignalResponse, error) {
	rows, err := s.repo.ListLatestByTarget(ctx, strings.TrimSpace(projectID), targetType, strings.TrimSpace(targetID), limit)
	if err != nil {
		return nil, err
	}
	now := s.now()
	items := make([]HealthSignalResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToResponse(row, now))
	}
	return items, nil
}

func validateSignal(row *HealthSignal) error {
	if row.ProjectID == "" || row.TargetID == "" || row.Source == "" {
		return response.NewBizError(response.CodeValidation, "invalid signal", "projectId, targetId, and source are required")
	}
	if !validTarget(row.TargetType) {
		return response.NewBizError(response.CodeValidation, "invalid signal target", string(row.TargetType))
	}
	if !validStatus(row.Status) {
		return response.NewBizError(response.CodeValidation, "invalid signal status", string(row.Status))
	}
	if !validSeverity(row.Severity) {
		return response.NewBizError(response.CodeValidation, "invalid signal severity", string(row.Severity))
	}
	return nil
}

func validTarget(value TargetType) bool {
	switch value {
	case TargetService, TargetEnvironment, TargetPipeline, TargetDeployment, TargetIntegration:
		return true
	default:
		return false
	}
}

func validStatus(value Status) bool {
	switch value {
	case StatusHealthy, StatusWarning, StatusDegraded, StatusUnknown:
		return true
	default:
		return false
	}
}

func validSeverity(value Severity) bool {
	switch value {
	case SeverityInfo, SeverityWarning, SeverityCritical:
		return true
	default:
		return false
	}
}
