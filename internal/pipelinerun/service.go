package pipelinerun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/internal/variable"
	"github.com/xjy/zcid/pkg/response"
	"github.com/xjy/zcid/pkg/tekton"
)

// PipelineGetter retrieves pipeline by ID and project
type PipelineGetter interface {
	GetByIDAndProject(ctx context.Context, id, projectID string) (*pipeline.Pipeline, error)
}

type Service struct {
	repo            Repository
	pipelineRepo    PipelineGetter
	variableService *variable.Service
	translator      *tekton.Translator
	k8sClient       K8sClient
	secretInjector  SecretInjector
}

func NewService(repo Repository, pipelineRepo PipelineGetter, variableService *variable.Service, translator *tekton.Translator, k8sClient K8sClient, secretInjector SecretInjector) *Service {
	return &Service{
		repo:            repo,
		pipelineRepo:    pipelineRepo,
		variableService: variableService,
		translator:      translator,
		k8sClient:       k8sClient,
		secretInjector:  secretInjector,
	}
}

func (s *Service) TriggerRun(ctx context.Context, projectID, pipelineID, userID string, req TriggerRunRequest) (*PipelineRunResponse, error) {
	p, err := s.pipelineRepo.GetByIDAndProject(ctx, pipelineID, projectID)
	if err != nil {
		if errors.Is(err, pipeline.ErrNotFound) {
			return nil, response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询流水线失败", err.Error())
	}

	running, err := s.repo.CountRunning(ctx, pipelineID)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "检查运行中任务失败", err.Error())
	}

	switch p.ConcurrencyPolicy {
	case pipeline.ConcurrencyReject:
		if running > 0 {
			return nil, response.NewBizError(response.CodeRunConcurrency, "流水线已有运行中的任务，请稍后重试", "")
		}
	case pipeline.ConcurrencyCancelOld:
		if running > 0 {
			runningRuns, listErr := s.repo.ListRunning(ctx, pipelineID)
			if listErr == nil {
				for _, r := range runningRuns {
					_ = s.CancelRun(ctx, projectID, r.ID)
				}
			}
		}
	case pipeline.ConcurrencyQueue:
		// No check, allow queueing
	}

	runNumber, err := s.repo.GetNextRunNumber(ctx, pipelineID)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "获取运行编号失败", err.Error())
	}

	runID := uuid.New().String()
	namespace := "zcid-run"

	plainParams := make(map[string]string)
	secretParams := make(map[string]string)
	if s.variableService != nil {
		resolved, err := s.variableService.ResolveVariables(projectID, pipelineID)
		if err == nil {
			for _, v := range resolved {
				if v.VarType == variable.TypeSecret {
					secretParams[v.Key] = v.Value
				} else {
					plainParams[v.Key] = v.Value
				}
			}
		}
	}
	for k, v := range req.Params {
		plainParams[k] = v
	}
	if req.GitBranch != "" {
		plainParams["GIT_BRANCH"] = req.GitBranch
	}
	if req.GitCommit != "" {
		plainParams["GIT_COMMIT"] = req.GitCommit
	}

	params := plainParams
	var secretName string
	if s.secretInjector != nil && len(secretParams) > 0 {
		var injErr error
		secretName, injErr = s.secretInjector.InjectSecrets(ctx, namespace, runID, secretParams)
		if injErr != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "注入密钥失败", injErr.Error())
		}
		// Secret keys added as ValueFrom env refs via injectSecretEnvRefs after Translate
	}

	gitInfo := &tekton.GitInfo{}
	if req.GitBranch != "" {
		gitInfo.Branch = req.GitBranch
	}
	if req.GitCommit != "" {
		gitInfo.CommitSHA = req.GitCommit
	}

	configSnapshot, err := json.Marshal(p.Config)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "配置序列化失败", err.Error())
	}

	run := &PipelineRun{
		ID:             runID,
		PipelineID:     pipelineID,
		ProjectID:      projectID,
		RunNumber:      runNumber,
		Status:         StatusQueued,
		TriggerType:    string(p.TriggerType),
		TriggeredBy:    &userID,
		ConfigSnapshot: configSnapshot,
		Params:         params,
		Namespace:      &namespace,
	}
	if req.GitBranch != "" {
		run.GitBranch = &req.GitBranch
	}
	if req.GitCommit != "" {
		run.GitCommit = &req.GitCommit
	}

	if err := s.repo.Create(ctx, run); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "创建运行记录失败", err.Error())
	}

	pr, err := s.translator.Translate(pipelineID, run.ID, projectID, namespace, p.Config, params, gitInfo)
	if err != nil {
		return nil, response.NewBizError(response.CodePipelineCRDFailed, "转换流水线配置失败", err.Error())
	}

	// FR13: Inject secret env refs into all steps
	if secretName != "" && len(secretParams) > 0 {
		injectSecretEnvRefs(pr, secretName, secretParams)
	}

	safePrefix := func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n]
	}
	pr.ObjectMeta.Name = fmt.Sprintf("run-%s-%d", safePrefix(pipelineID, 8), runNumber)
	if len(pr.ObjectMeta.Name) > 63 {
		pr.ObjectMeta.Name = pr.ObjectMeta.Name[:63]
	}

	tektonName := pr.ObjectMeta.Name
	if err := s.repo.Update(ctx, run.ID, projectID, map[string]interface{}{"tekton_name": tektonName}); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "更新运行记录失败", err.Error())
	}

	if err := s.k8sClient.SubmitPipelineRun(ctx, namespace, pr); err != nil {
		_ = s.repo.UpdateStatus(ctx, run.ID, projectID, StatusFailed, ptr("提交到集群失败"))
		return nil, response.NewBizError(response.CodeRunSubmitFailed, "提交流水线运行失败", err.Error())
	}

	_ = secretName

	go s.syncRunStatus(run.ID, projectID, namespace, tektonName)

	return toResponse(run), nil
}

func (s *Service) syncRunStatus(runID, projectID, namespace, tektonName string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)
	var started bool

	for {
		select {
		case <-timeout:
			_ = s.repo.UpdateStatus(context.Background(), runID, projectID, StatusFailed, ptr("运行超时"))
			return
		case <-ticker.C:
			status, err := s.k8sClient.GetPipelineRunStatus(context.Background(), namespace, tektonName)
			if err != nil {
				continue
			}

			switch status {
			case "Running":
				if !started {
					started = true
					now := time.Now()
					_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
						"status":     StatusRunning,
						"started_at": now,
						"updated_at": now,
					})
				}
			case "Succeeded":
				now := time.Now()
				_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
					"status":      StatusSucceeded,
					"finished_at": now,
					"updated_at":  now,
				})
				return
			case "Failed":
				now := time.Now()
				_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
					"status":        StatusFailed,
					"finished_at":   now,
					"updated_at":    now,
					"error_message": "Pipeline execution failed",
				})
				return
			}
		}
	}
}

func (s *Service) CancelRun(ctx context.Context, projectID, runID string) error {
	run, err := s.repo.GetByIDAndProject(ctx, runID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeRunNotFound, "运行记录不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "查询运行记录失败", err.Error())
	}

	if run.Status != StatusPending && run.Status != StatusQueued && run.Status != StatusRunning {
		return response.NewBizError(response.CodeRunAlreadyDone, "运行已结束，无法取消", "")
	}

	if run.TektonName != nil && run.Namespace != nil && *run.TektonName != "" {
		if err := s.k8sClient.DeletePipelineRun(ctx, *run.Namespace, *run.TektonName); err != nil {
			return response.NewBizError(response.CodeRunCancelFailed, "取消运行失败", err.Error())
		}
	}

	now := time.Now()
	return s.repo.Update(ctx, runID, projectID, map[string]interface{}{
		"status":      StatusCancelled,
		"finished_at": now,
		"updated_at":  now,
	})
}

func (s *Service) GetRun(ctx context.Context, projectID, runID string) (*PipelineRunResponse, error) {
	run, err := s.repo.GetByIDAndProject(ctx, runID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeRunNotFound, "运行记录不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询运行记录失败", err.Error())
	}
	return toResponse(run), nil
}

func (s *Service) ListRuns(ctx context.Context, projectID, pipelineID string, page, pageSize int) (*PipelineRunListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	runs, total, err := s.repo.ListByPipeline(ctx, pipelineID, projectID, page, pageSize)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询运行列表失败", err.Error())
	}

	items := make([]PipelineRunSummary, len(runs))
	for i, r := range runs {
		items[i] = toSummary(r)
	}
	return &PipelineRunListResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) UpdateArtifacts(ctx context.Context, projectID, runID string, artifacts []Artifact) error {
	_, err := s.repo.GetByIDAndProject(ctx, runID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeRunNotFound, "运行记录不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "查询运行记录失败", err.Error())
	}

	return s.repo.UpdateArtifacts(ctx, runID, projectID, artifacts)
}

func (s *Service) GetArtifacts(ctx context.Context, projectID, runID string) ([]Artifact, error) {
	run, err := s.repo.GetByIDAndProject(ctx, runID, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeRunNotFound, "运行记录不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询运行记录失败", err.Error())
	}
	if run.Artifacts == nil {
		return []Artifact{}, nil
	}
	return run.Artifacts, nil
}

func toResponse(r *PipelineRun) *PipelineRunResponse {
	resp := &PipelineRunResponse{
		ID:           r.ID,
		PipelineID:   r.PipelineID,
		ProjectID:    r.ProjectID,
		RunNumber:    r.RunNumber,
		Status:       string(r.Status),
		TriggerType:  r.TriggerType,
		TriggeredBy:  r.TriggeredBy,
		GitBranch:    r.GitBranch,
		GitCommit:    r.GitCommit,
		GitAuthor:    r.GitAuthor,
		GitMessage:   r.GitMessage,
		Params:       r.Params,
		TektonName:   r.TektonName,
		Namespace:    r.Namespace,
		StartedAt:    r.StartedAt,
		FinishedAt:   r.FinishedAt,
		ErrorMessage: r.ErrorMessage,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	if r.Artifacts != nil {
		resp.Artifacts = r.Artifacts
	}
	return resp
}

func toSummary(r *PipelineRun) PipelineRunSummary {
	return PipelineRunSummary{
		ID:          r.ID,
		PipelineID:  r.PipelineID,
		RunNumber:   r.RunNumber,
		Status:      string(r.Status),
		TriggerType: r.TriggerType,
		TriggeredBy: r.TriggeredBy,
		GitBranch:   r.GitBranch,
		StartedAt:   r.StartedAt,
		FinishedAt:  r.FinishedAt,
		CreatedAt:   r.CreatedAt,
	}
}

func ptr(s string) *string {
	return &s
}

func injectSecretEnvRefs(pr *tekton.PipelineRun, secretName string, secretKeys map[string]string) {
	for k := range secretKeys {
		ref := &tekton.EnvVar{
			Name: k,
			ValueFrom: &tekton.EnvVarSource{
				SecretKeyRef: &tekton.SecretKeyRef{Name: secretName, Key: k},
			},
		}
		for i := range pr.Spec.PipelineSpec.Tasks {
			t := &pr.Spec.PipelineSpec.Tasks[i]
			if t.TaskSpec != nil {
				for j := range t.TaskSpec.Steps {
					t.TaskSpec.Steps[j].Env = append(t.TaskSpec.Steps[j].Env, *ref)
				}
			}
		}
	}
}
