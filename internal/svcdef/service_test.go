package svcdef

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

func (m *mockRepo) Create(ctx context.Context, s *ServiceDef) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *mockRepo) FindByID(ctx context.Context, id, projectID string) (*ServiceDef, error) {
	args := m.Called(ctx, id, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ServiceDef), args.Error(1)
}

func (m *mockRepo) ListByProject(ctx context.Context, projectID string, page, pageSize int) ([]*ServiceDef, int64, error) {
	args := m.Called(ctx, projectID, page, pageSize)
	return args.Get(0).([]*ServiceDef), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) Update(ctx context.Context, id, projectID string, updates map[string]any) error {
	args := m.Called(ctx, id, projectID, updates)
	return args.Error(0)
}

func (m *mockRepo) SoftDelete(ctx context.Context, id, projectID string) error {
	args := m.Called(ctx, id, projectID)
	return args.Error(0)
}

func (m *mockRepo) ListLinkedPipelines(ctx context.Context, svc *ServiceDef) ([]VitalsPipeline, error) {
	args := m.Called(ctx, svc)
	return args.Get(0).([]VitalsPipeline), args.Error(1)
}

func (m *mockRepo) ListRecentRuns(ctx context.Context, projectID string, pipelineIDs []string, limit int) ([]VitalsRun, error) {
	args := m.Called(ctx, projectID, pipelineIDs, limit)
	return args.Get(0).([]VitalsRun), args.Error(1)
}

func (m *mockRepo) ListLatestDeployments(ctx context.Context, projectID string, environmentIDs []string, limit int) ([]VitalsDeployment, error) {
	args := m.Called(ctx, projectID, environmentIDs, limit)
	return args.Get(0).([]VitalsDeployment), args.Error(1)
}

func (m *mockRepo) ListFailedSteps(ctx context.Context, projectID string, runIDs []string, limit int) ([]VitalsStepWarning, error) {
	args := m.Called(ctx, projectID, runIDs, limit)
	return args.Get(0).([]VitalsStepWarning), args.Error(1)
}

func (m *mockRepo) ListLatestSignals(ctx context.Context, projectID string, targets []VitalsSignalTarget, limit int) ([]VitalsSignal, error) {
	args := m.Called(ctx, projectID, targets, limit)
	return args.Get(0).([]VitalsSignal), args.Error(1)
}

func TestCreate_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*svcdef.ServiceDef")).
		Return(nil).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*ServiceDef)
			s.ID = "svc-123"
		})

	s, err := svc.Create(ctx, "proj-1", "api-gateway", "API Gateway service", "https://github.com/org/api")

	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", s.Name)
	assert.Equal(t, "proj-1", s.ProjectID)
	repo.AssertExpectations(t)
}

func TestCreateWithInput_Metadata(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*svcdef.ServiceDef")).
		Return(nil).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*ServiceDef)
			assert.Equal(t, "api", s.ServiceType)
			assert.Equal(t, "go", s.Language)
			assert.Equal(t, "platform", s.Owner)
			assert.Equal(t, StringList{"critical", "payments"}, s.Tags)
			assert.Equal(t, StringList{"pipe-1"}, s.PipelineIDs)
			assert.Equal(t, StringList{"env-1"}, s.EnvironmentIDs)
			s.ID = "svc-123"
		})

	created, err := svc.CreateWithInput(ctx, "proj-1", CreateInput{
		Name:           " api-gateway ",
		Description:    " API Gateway service ",
		RepoURL:        " https://github.com/org/api ",
		ServiceType:    " api ",
		Language:       " go ",
		Owner:          " platform ",
		Tags:           []string{"critical", "", "payments", "critical"},
		PipelineIDs:    []string{"pipe-1", "pipe-1"},
		EnvironmentIDs: []string{"env-1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", created.Name)
	repo.AssertExpectations(t)
}

func TestCreate_EmptyName(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "proj-1", "  ", "", "")

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

	_, err := svc.Create(ctx, "proj-1", "existing", "", "")

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

	expected := []*ServiceDef{{ID: "s1", Name: "frontend"}, {ID: "s2", Name: "backend"}}
	repo.On("ListByProject", ctx, "proj-1", 1, 20).Return(expected, int64(2), nil)

	svcs, total, err := svc.List(ctx, "proj-1", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, svcs, 2)
}

func TestDelete_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()

	repo.On("SoftDelete", ctx, "s1", "proj-1").Return(nil)

	err := svc.Delete(ctx, "s1", "proj-1")

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

func TestUpdateWithInput_Metadata(t *testing.T) {
	repo := new(mockRepo)
	svc := NewService(repo)
	ctx := context.Background()
	owner := " team-a "
	serviceType := "worker"

	repo.On("Update", ctx, "s1", "proj-1", mock.MatchedBy(func(updates map[string]any) bool {
		return updates["owner"] == "team-a" &&
			updates["service_type"] == "worker" &&
			assert.ObjectsAreEqual(StringList{"batch"}, updates["tags"])
	})).Return(nil)
	repo.On("FindByID", ctx, "s1", "proj-1").Return(&ServiceDef{ID: "s1", ProjectID: "proj-1", Name: "svc", Owner: "team-a", ServiceType: "worker", Tags: StringList{"batch"}}, nil)

	updated, err := svc.UpdateWithInput(ctx, "s1", "proj-1", UpdateInput{
		Owner:       &owner,
		ServiceType: &serviceType,
		Tags:        []string{"batch"},
		UpdateTags:  true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "team-a", updated.Owner)
	repo.AssertExpectations(t)
}
