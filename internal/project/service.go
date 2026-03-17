package project

import (
	"context"
	"errors"
	"strings"

	"github.com/xjy/zcid/pkg/response"
)

type Repository interface {
	Create(ctx context.Context, p *Project) error
	FindByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, page, pageSize int) ([]*Project, int64, error)
	ListByIDs(ctx context.Context, ids []string, page, pageSize int) ([]*Project, int64, error)
	Update(ctx context.Context, id string, updates map[string]any) error
	SoftDelete(ctx context.Context, id string) error
	DeleteProjectCascade(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *ProjectMember) error
	RemoveMembersByProject(ctx context.Context, projectID string) error
	GetUserProjectIDs(ctx context.Context, userID string) ([]string, error)
	IsProjectMember(ctx context.Context, projectID, userID string) (bool, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (ProjectRole, error)
	ListMembers(ctx context.Context, projectID string) ([]MemberWithUsername, error)
	RemoveMember(ctx context.Context, projectID, userID string) error
	UpdateMemberRole(ctx context.Context, projectID, userID string, role ProjectRole) error
	SoftDeleteEnvironmentsByProject(ctx context.Context, projectID string) error
	SoftDeleteServicesByProject(ctx context.Context, projectID string) error
	SoftDeleteVariablesByProject(ctx context.Context, projectID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, name, description, ownerID string) (*Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, response.NewBizError(response.CodeValidation, "项目名称不能为空", "")
	}

	p := &Project{
		Name:        name,
		Description: strings.TrimSpace(description),
		OwnerID:     ownerID,
		Status:      ProjectStatusActive,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, ErrProjectNameTaken) {
			return nil, response.NewBizError(response.CodeConflict, "项目名称已存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	member := &ProjectMember{
		ProjectID: p.ID,
		UserID:    ownerID,
		Role:      RoleProjectAdmin,
	}
	if err := s.repo.AddMember(ctx, member); err != nil && !errors.Is(err, ErrMemberExists) {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return p, nil
}

func (s *Service) GetProject(ctx context.Context, id string) (*Project, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "项目不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return p, nil
}

func (s *Service) ListProjects(ctx context.Context, userID, userRole string, page, pageSize int) ([]*Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if userRole == "admin" {
		projects, total, err := s.repo.List(ctx, page, pageSize)
		if err != nil {
			return nil, 0, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
		return projects, total, nil
	}

	projectIDs, err := s.repo.GetUserProjectIDs(ctx, userID)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	projects, total, err := s.repo.ListByIDs(ctx, projectIDs, page, pageSize)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return projects, total, nil
}

func (s *Service) UpdateProject(ctx context.Context, id string, name, description *string) (*Project, error) {
	updates := map[string]any{}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "项目名称不能为空", "")
		}
		updates["name"] = trimmed
	}

	if description != nil {
		updates["description"] = strings.TrimSpace(*description)
	}

	if len(updates) == 0 {
		return nil, response.NewBizError(response.CodeValidation, "至少需要更新一个字段", "")
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			return nil, response.NewBizError(response.CodeNotFound, "项目不存在", "")
		case errors.Is(err, ErrProjectNameTaken):
			return nil, response.NewBizError(response.CodeConflict, "项目名称已存在", "")
		default:
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	return s.repo.FindByID(ctx, id)
}

func (s *Service) AddMember(ctx context.Context, projectID, userID, role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return response.NewBizError(response.CodeValidation, "角色不能为空", "")
	}
	if role != string(RoleProjectAdmin) && role != string(RoleMember) {
		return response.NewBizError(response.CodeValidation, "无效的角色类型", "")
	}

	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      ProjectRole(role),
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		if errors.Is(err, ErrMemberExists) {
			return response.NewBizError(response.CodeConflict, "用户已是该项目成员", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, projectID, userID string) error {
	if err := s.repo.RemoveMember(ctx, projectID, userID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return response.NewBizError(response.CodeNotFound, "项目成员不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, projectID, userID, role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return response.NewBizError(response.CodeValidation, "角色不能为空", "")
	}
	if role != string(RoleProjectAdmin) && role != string(RoleMember) {
		return response.NewBizError(response.CodeValidation, "无效的角色类型", "")
	}

	if err := s.repo.UpdateMemberRole(ctx, projectID, userID, ProjectRole(role)); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return response.NewBizError(response.CodeNotFound, "项目成员不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, projectID string) ([]MemberWithUsername, error) {
	members, err := s.repo.ListMembers(ctx, projectID)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return members, nil
}

func (s *Service) DeleteProject(ctx context.Context, id string) error {
	if err := s.repo.DeleteProjectCascade(ctx, id); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return response.NewBizError(response.CodeNotFound, "项目不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return nil
}
