package project

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, p *Project) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Project), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, page, pageSize int) ([]*Project, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*Project), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) ListByIDs(ctx context.Context, ids []string, page, pageSize int) ([]*Project, int64, error) {
	args := m.Called(ctx, ids, page, pageSize)
	return args.Get(0).([]*Project), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) Update(ctx context.Context, id string, updates map[string]any) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *mockRepo) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepo) AddMember(ctx context.Context, member *ProjectMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockRepo) RemoveMembersByProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *mockRepo) GetUserProjectIDs(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockRepo) IsProjectMember(ctx context.Context, projectID, userID string) (bool, error) {
	args := m.Called(ctx, projectID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepo) GetMemberRole(ctx context.Context, projectID, userID string) (ProjectRole, error) {
	args := m.Called(ctx, projectID, userID)
	return args.Get(0).(ProjectRole), args.Error(1)
}

func (m *mockRepo) ListMembers(ctx context.Context, projectID string) ([]MemberWithUsername, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]MemberWithUsername), args.Error(1)
}

func (m *mockRepo) RemoveMember(ctx context.Context, projectID, userID string) error {
	args := m.Called(ctx, projectID, userID)
	return args.Error(0)
}

func (m *mockRepo) UpdateMemberRole(ctx context.Context, projectID, userID string, role ProjectRole) error {
	args := m.Called(ctx, projectID, userID, role)
	return args.Error(0)
}

func (m *mockRepo) SoftDeleteEnvironmentsByProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *mockRepo) SoftDeleteServicesByProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *mockRepo) SoftDeleteVariablesByProject(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func TestCreateProject_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*project.Project")).
		Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(1).(*Project)
			p.ID = "proj-123"
		})
	repo.On("AddMember", ctx, mock.AnythingOfType("*project.ProjectMember")).Return(nil)

	p, err := svc.CreateProject(ctx, "my-project", "desc", "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "my-project", p.Name)
	assert.Equal(t, "desc", p.Description)
	assert.Equal(t, "user-1", p.OwnerID)
	repo.AssertExpectations(t)
}

func TestCreateProject_EmptyName(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.CreateProject(context.Background(), "  ", "", "user-1")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
}

func TestCreateProject_NameConflict(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.Anything).Return(ErrProjectNameTaken)

	_, err := svc.CreateProject(ctx, "existing-project", "", "user-1")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeConflict, bizErr.Code)
}

func TestListProjects_AdminSeesAll(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	expected := []*Project{{ID: "p1", Name: "proj1"}, {ID: "p2", Name: "proj2"}}
	repo.On("List", ctx, 1, 20).Return(expected, int64(2), nil)

	projects, total, err := svc.ListProjects(ctx, "admin-user", "admin", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, projects, 2)
}

func TestListProjects_MemberSeesOwnProjects(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("GetUserProjectIDs", ctx, "member-user").Return([]string{"p1"}, nil)
	expected := []*Project{{ID: "p1", Name: "proj1"}}
	repo.On("ListByIDs", ctx, []string{"p1"}, 1, 20).Return(expected, int64(1), nil)

	projects, total, err := svc.ListProjects(ctx, "member-user", "member", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, projects, 1)
}

func TestGetProject_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("FindByID", ctx, "nonexistent").Return(nil, ErrProjectNotFound)

	_, err := svc.GetProject(ctx, "nonexistent")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeNotFound, bizErr.Code)
}

func TestDeleteProject_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("SoftDelete", ctx, "p1").Return(nil)
	repo.On("SoftDeleteEnvironmentsByProject", ctx, "p1").Return(nil)
	repo.On("SoftDeleteServicesByProject", ctx, "p1").Return(nil)
	repo.On("SoftDeleteVariablesByProject", ctx, "p1").Return(nil)
	repo.On("RemoveMembersByProject", ctx, "p1").Return(nil)

	err := svc.DeleteProject(ctx, "p1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteProject_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("SoftDelete", ctx, "nonexistent").Return(ErrProjectNotFound)

	err := svc.DeleteProject(ctx, "nonexistent")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeNotFound, bizErr.Code)
}

func TestUpdateProject_NameConflict(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	name := "existing-name"
	repo.On("Update", ctx, "p1", mock.Anything).Return(ErrProjectNameTaken)

	_, err := svc.UpdateProject(ctx, "p1", &name, nil)

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeConflict, bizErr.Code)
}

func TestAddMember_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("AddMember", ctx, mock.AnythingOfType("*project.ProjectMember")).Return(nil)

	err := svc.AddMember(ctx, "proj-1", "user-2", "member")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAddMember_Duplicate(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("AddMember", ctx, mock.AnythingOfType("*project.ProjectMember")).Return(ErrMemberExists)

	err := svc.AddMember(ctx, "proj-1", "user-2", "member")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeConflict, bizErr.Code)
}

func TestAddMember_InvalidRole(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	err := svc.AddMember(ctx, "proj-1", "user-2", "invalid_role")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
}

func TestRemoveMember_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("RemoveMember", ctx, "proj-1", "user-2").Return(nil)

	err := svc.RemoveMember(ctx, "proj-1", "user-2")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRemoveMember_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("RemoveMember", ctx, "proj-1", "nonexistent").Return(ErrMemberNotFound)

	err := svc.RemoveMember(ctx, "proj-1", "nonexistent")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeNotFound, bizErr.Code)
}

func TestUpdateMemberRole_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("UpdateMemberRole", ctx, "proj-1", "user-2", RoleProjectAdmin).Return(nil)

	err := svc.UpdateMemberRole(ctx, "proj-1", "user-2", "project_admin")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestListMembers_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	expected := []MemberWithUsername{
		{UserID: "user-1", Username: "alice", Role: RoleProjectAdmin, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{UserID: "user-2", Username: "bob", Role: RoleMember, CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	}
	repo.On("ListMembers", ctx, "proj-1").Return(expected, nil)

	members, err := svc.ListMembers(ctx, "proj-1")

	assert.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "alice", members[0].Username)
}
