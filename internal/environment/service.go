package environment

import (
	"context"
	"errors"
	"strings"

	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/response"
)

type Repository interface {
	Create(ctx context.Context, e *Environment) error
	FindByID(ctx context.Context, id, projectID string) (*Environment, error)
	ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Environment, int64, error)
	Update(ctx context.Context, id, projectID string, updates map[string]any) error
	SoftDelete(ctx context.Context, id, projectID string) error
}

type Service struct {
	repo      Repository
	signalSvc *signal.Service
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetSignalService(signalSvc *signal.Service) {
	s.signalSvc = signalSvc
}

func (s *Service) Create(ctx context.Context, projectID, name, namespace, description string) (*Environment, error) {
	name = strings.TrimSpace(name)
	namespace = strings.TrimSpace(namespace)
	if name == "" {
		return nil, response.NewBizError(response.CodeValidation, "环境名称不能为空", "")
	}
	if namespace == "" {
		return nil, response.NewBizError(response.CodeValidation, "Namespace 不能为空", "")
	}

	e := &Environment{
		ProjectID:   projectID,
		Name:        name,
		Namespace:   namespace,
		Description: strings.TrimSpace(description),
		Status:      StatusActive,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		switch {
		case errors.Is(err, ErrNameTaken):
			return nil, response.NewBizError(response.CodeConflict, "该项目下环境名称已存在", "")
		case errors.Is(err, ErrNamespaceTaken):
			return nil, response.NewBizError(response.CodeConflict, "Namespace 已被占用", "")
		default:
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}
	return e, nil
}

func (s *Service) Get(ctx context.Context, id, projectID string) (*Environment, error) {
	e, err := s.repo.FindByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "环境不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return e, nil
}

func (s *Service) List(ctx context.Context, projectID string, page, pageSize int) ([]*Environment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	envs, total, err := s.repo.ListByProject(ctx, projectID, page, pageSize)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return envs, total, nil
}

func (s *Service) Health(ctx context.Context, projectID, envID string) EnvironmentHealthResponse {
	if s.signalSvc == nil {
		return EnvironmentHealthResponse{Status: "unknown", Reason: "No health signal service configured"}
	}
	signals, err := s.signalSvc.ListLatestByTarget(ctx, projectID, signal.TargetEnvironment, envID, 1)
	if err != nil || len(signals) == 0 {
		return EnvironmentHealthResponse{Status: "unknown", Reason: "No environment health signals yet"}
	}
	item := signals[0]
	resp := EnvironmentHealthResponse{
		Status:       item.EffectiveStatus,
		Reason:       firstNonEmpty(item.Message, item.Reason, "Latest environment signal"),
		LastSignalAt: item.ObservedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Stale:        item.EffectiveStatus == string(signal.StatusStale),
	}
	return resp
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (s *Service) Update(ctx context.Context, id, projectID string, name, namespace, description *string) (*Environment, error) {
	updates := map[string]any{}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "环境名称不能为空", "")
		}
		updates["name"] = trimmed
	}
	if namespace != nil {
		trimmed := strings.TrimSpace(*namespace)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "Namespace 不能为空", "")
		}
		updates["namespace"] = trimmed
	}
	if description != nil {
		updates["description"] = strings.TrimSpace(*description)
	}

	if len(updates) == 0 {
		return nil, response.NewBizError(response.CodeValidation, "至少需要更新一个字段", "")
	}

	if err := s.repo.Update(ctx, id, projectID, updates); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, response.NewBizError(response.CodeNotFound, "环境不存在", "")
		case errors.Is(err, ErrNameTaken):
			return nil, response.NewBizError(response.CodeConflict, "该项目下环境名称已存在", "")
		case errors.Is(err, ErrNamespaceTaken):
			return nil, response.NewBizError(response.CodeConflict, "Namespace 已被占用", "")
		default:
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	return s.repo.FindByID(ctx, id, projectID)
}

func (s *Service) Delete(ctx context.Context, id, projectID string) error {
	if err := s.repo.SoftDelete(ctx, id, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotFound, "环境不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}
