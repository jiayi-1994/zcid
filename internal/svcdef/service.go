package svcdef

import (
	"context"
	"errors"
	"strings"

	"github.com/xjy/zcid/pkg/response"
)

type Repository interface {
	Create(ctx context.Context, s *ServiceDef) error
	FindByID(ctx context.Context, id, projectID string) (*ServiceDef, error)
	ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*ServiceDef, int64, error)
	Update(ctx context.Context, id, projectID string, updates map[string]any) error
	SoftDelete(ctx context.Context, id, projectID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, projectID, name, description, repoURL string) (*ServiceDef, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, response.NewBizError(response.CodeValidation, "服务名称不能为空", "")
	}

	svc := &ServiceDef{
		ProjectID:   projectID,
		Name:        name,
		Description: strings.TrimSpace(description),
		RepoURL:     strings.TrimSpace(repoURL),
		Status:      StatusActive,
	}

	if err := s.repo.Create(ctx, svc); err != nil {
		if errors.Is(err, ErrNameTaken) {
			return nil, response.NewBizError(response.CodeConflict, "该项目下服务名称已存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return svc, nil
}

func (s *Service) Get(ctx context.Context, id, projectID string) (*ServiceDef, error) {
	svc, err := s.repo.FindByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "服务不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return svc, nil
}

func (s *Service) List(ctx context.Context, projectID string, page, pageSize int) ([]*ServiceDef, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	svcs, total, err := s.repo.ListByProject(ctx, projectID, page, pageSize)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return svcs, total, nil
}

func (s *Service) Update(ctx context.Context, id, projectID string, name, description, repoURL *string) (*ServiceDef, error) {
	updates := map[string]any{}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "服务名称不能为空", "")
		}
		updates["name"] = trimmed
	}
	if description != nil {
		updates["description"] = strings.TrimSpace(*description)
	}
	if repoURL != nil {
		updates["repo_url"] = strings.TrimSpace(*repoURL)
	}

	if len(updates) == 0 {
		return nil, response.NewBizError(response.CodeValidation, "至少需要更新一个字段", "")
	}

	if err := s.repo.Update(ctx, id, projectID, updates); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, response.NewBizError(response.CodeNotFound, "服务不存在", "")
		case errors.Is(err, ErrNameTaken):
			return nil, response.NewBizError(response.CodeConflict, "该项目下服务名称已存在", "")
		default:
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	return s.repo.FindByID(ctx, id, projectID)
}

func (s *Service) Delete(ctx context.Context, id, projectID string) error {
	if err := s.repo.SoftDelete(ctx, id, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotFound, "服务不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}
