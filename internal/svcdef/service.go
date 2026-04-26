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
	ListLinkedPipelines(ctx context.Context, svc *ServiceDef) ([]VitalsPipeline, error)
	ListRecentRuns(ctx context.Context, projectID string, pipelineIDs []string, limit int) ([]VitalsRun, error)
	ListLatestDeployments(ctx context.Context, projectID string, environmentIDs []string, limit int) ([]VitalsDeployment, error)
	ListFailedSteps(ctx context.Context, projectID string, runIDs []string, limit int) ([]VitalsStepWarning, error)
	ListLatestSignals(ctx context.Context, projectID string, targets []VitalsSignalTarget, limit int) ([]VitalsSignal, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name           string
	Description    string
	RepoURL        string
	ServiceType    string
	Language       string
	Owner          string
	Tags           []string
	PipelineIDs    []string
	EnvironmentIDs []string
}

type UpdateInput struct {
	Name               *string
	Description        *string
	RepoURL            *string
	ServiceType        *string
	Language           *string
	Owner              *string
	Tags               []string
	PipelineIDs        []string
	EnvironmentIDs     []string
	UpdateTags         bool
	UpdatePipelines    bool
	UpdateEnvironments bool
}

func (s *Service) Create(ctx context.Context, projectID, name, description, repoURL string) (*ServiceDef, error) {
	return s.CreateWithInput(ctx, projectID, CreateInput{Name: name, Description: description, RepoURL: repoURL})
}

func (s *Service) CreateWithInput(ctx context.Context, projectID string, input CreateInput) (*ServiceDef, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, response.NewBizError(response.CodeValidation, "服务名称不能为空", "")
	}

	svc := &ServiceDef{
		ProjectID:      projectID,
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		RepoURL:        strings.TrimSpace(input.RepoURL),
		ServiceType:    strings.TrimSpace(input.ServiceType),
		Language:       strings.TrimSpace(input.Language),
		Owner:          strings.TrimSpace(input.Owner),
		Tags:           normalizeList(input.Tags),
		PipelineIDs:    normalizeList(input.PipelineIDs),
		EnvironmentIDs: normalizeList(input.EnvironmentIDs),
		Status:         StatusActive,
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
	return s.UpdateWithInput(ctx, id, projectID, UpdateInput{Name: name, Description: description, RepoURL: repoURL})
}

func (s *Service) UpdateWithInput(ctx context.Context, id, projectID string, input UpdateInput) (*ServiceDef, error) {
	updates := map[string]any{}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "服务名称不能为空", "")
		}
		updates["name"] = trimmed
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}
	if input.RepoURL != nil {
		updates["repo_url"] = strings.TrimSpace(*input.RepoURL)
	}
	if input.ServiceType != nil {
		updates["service_type"] = strings.TrimSpace(*input.ServiceType)
	}
	if input.Language != nil {
		updates["language"] = strings.TrimSpace(*input.Language)
	}
	if input.Owner != nil {
		updates["owner"] = strings.TrimSpace(*input.Owner)
	}
	if input.UpdateTags {
		updates["tags"] = normalizeList(input.Tags)
	}
	if input.UpdatePipelines {
		updates["pipeline_ids"] = normalizeList(input.PipelineIDs)
	}
	if input.UpdateEnvironments {
		updates["environment_ids"] = normalizeList(input.EnvironmentIDs)
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

func normalizeList(values []string) StringList {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item != "" {
			trimmed = append(trimmed, item)
		}
	}
	return NewStringList(trimmed)
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
