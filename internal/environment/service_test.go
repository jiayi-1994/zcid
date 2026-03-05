package environment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, e *Environment) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockRepo) FindByID(ctx context.Context, id, projectID string) (*Environment, error) {
	args := m.Called(ctx, id, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Environment), args.Error(1)
}

func (m *mockRepo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*Environment, int64, error) {
	args := m.Called(ctx, projectID, page, pageSize)
	return args.Get(0).([]*Environment), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	args := m.Called(ctx, id, projectID, updates)
	return args.Error(0)
}

func (m *mockRepo) SoftDelete(ctx context.Context, id, projectID string) error {
	args := m.Called(ctx, id, projectID)
	return args.Error(0)
}

func TestCreate_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*environment.Environment")).
		Return(nil).
		Run(func(args mock.Arguments) {
			e := args.Get(1).(*Environment)
			e.ID = "env-123"
		})

	env, err := svc.Create(ctx, "proj-1", "dev", "dev-ns", "Development")

	assert.NoError(t, err)
	assert.Equal(t, "dev", env.Name)
	assert.Equal(t, "dev-ns", env.Namespace)
	assert.Equal(t, "proj-1", env.ProjectID)
	repo.AssertExpectations(t)
}

func TestCreate_EmptyName(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "proj-1", "  ", "ns", "")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
}

func TestCreate_EmptyNamespace(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "proj-1", "dev", "", "")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeValidation, bizErr.Code)
}

func TestCreate_NameConflict(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.Anything).Return(ErrNameTaken)

	_, err := svc.Create(ctx, "proj-1", "existing", "ns", "")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeConflict, bizErr.Code)
}

func TestCreate_NamespaceTaken(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.Anything).Return(ErrNamespaceTaken)

	_, err := svc.Create(ctx, "proj-1", "dev", "taken-ns", "")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeConflict, bizErr.Code)
}

func TestGet_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("FindByID", ctx, "nonexistent", "proj-1").Return(nil, ErrNotFound)

	_, err := svc.Get(ctx, "nonexistent", "proj-1")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeNotFound, bizErr.Code)
}

func TestList_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	expected := []*Environment{{ID: "e1", Name: "dev"}, {ID: "e2", Name: "staging"}}
	repo.On("ListByProject", ctx, "proj-1", 1, 20).Return(expected, int64(2), nil)

	envs, total, err := svc.List(ctx, "proj-1", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, envs, 2)
}

func TestDelete_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("SoftDelete", ctx, "e1", "proj-1").Return(nil)

	err := svc.Delete(ctx, "e1", "proj-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("SoftDelete", ctx, "nonexistent", "proj-1").Return(ErrNotFound)

	err := svc.Delete(ctx, "nonexistent", "proj-1")

	assert.Error(t, err)
	bizErr, ok := err.(*response.BizError)
	assert.True(t, ok)
	assert.Equal(t, response.CodeNotFound, bizErr.Code)
}
