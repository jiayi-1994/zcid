package pipelinerun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/xjy/zcid/internal/notification"
	"github.com/xjy/zcid/internal/pipeline"
	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/internal/stepexec"
	"github.com/xjy/zcid/internal/variable"
	"github.com/xjy/zcid/pkg/response"
	"github.com/xjy/zcid/pkg/tekton"
)

// PipelineGetter retrieves pipeline by ID and project
type PipelineGetter interface {
	GetByIDAndProject(ctx context.Context, id, projectID string) (*pipeline.Pipeline, error)
}

type NotificationDispatcher interface {
	SendWebhook(ctx context.Context, projectID string, event notification.EventType, payload map[string]any) error
}

type Service struct {
	repo            Repository
	pipelineRepo    PipelineGetter
	variableService *variable.Service
	translator      *tekton.Translator
	k8sClient       K8sClient
	secretInjector  SecretInjector
	stepRepo        stepexec.Repository
	signals         *signal.Service
	notifications   NotificationDispatcher
}

func NewService(repo Repository, pipelineRepo PipelineGetter, variableService *variable.Service, translator *tekton.Translator, k8sClient K8sClient, secretInjector SecretInjector, stepRepo ...stepexec.Repository) *Service {
	s := &Service{
		repo:            repo,
		pipelineRepo:    pipelineRepo,
		variableService: variableService,
		translator:      translator,
		k8sClient:       k8sClient,
		secretInjector:  secretInjector,
	}
	if len(stepRepo) > 0 {
		s.stepRepo = stepRepo[0]
	}
	return s
}

func (s *Service) SetSignalService(signals *signal.Service) {
	s.signals = signals
}

func (s *Service) SetNotificationService(service *notification.Service) {
	s.notifications = service
}

func (s *Service) SetNotificationDispatcher(dispatcher NotificationDispatcher) {
	s.notifications = dispatcher
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
		resolved, err := s.variableService.ResolveVariables(ctx, projectID, pipelineID)
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
	pr.Metadata.Name = fmt.Sprintf("run-%s-%d", safePrefix(pipelineID, 8), runNumber)
	if len(pr.Metadata.Name) > 63 {
		pr.Metadata.Name = pr.Metadata.Name[:63]
	}

	tektonName := pr.Metadata.Name
	if err := s.repo.Update(ctx, run.ID, projectID, map[string]interface{}{"tekton_name": tektonName}); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "更新运行记录失败", err.Error())
	}

	if err := s.k8sClient.SubmitPipelineRun(ctx, namespace, pr); err != nil {
		_ = s.repo.UpdateStatus(ctx, run.ID, projectID, StatusFailed, ptr("提交到集群失败"))
		run.Status = StatusFailed
		run.ErrorMessage = ptr("提交到集群失败: " + err.Error())
		s.recordPipelineSignal(ctx, run.ID, pipelineID, projectID, StatusFailed, "pipeline.submit_failed", "Pipeline submit failed", err.Error())
		s.notifyPipelineRun(ctx, run, p.Name, notification.EventBuildFailed)
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
			now := time.Now()
			_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
				"status":        StatusFailed,
				"finished_at":   now,
				"updated_at":    now,
				"error_message": "运行超时",
			})
			run, _ := s.repo.GetByIDAndProject(context.Background(), runID, projectID)
			if run != nil {
				run.Status = StatusFailed
				run.FinishedAt = &now
				run.ErrorMessage = ptr("运行超时")
				s.recordPipelineSignal(context.Background(), runID, run.PipelineID, projectID, StatusFailed, "pipeline.timeout", "Pipeline run timed out", "")
				s.notifyPipelineRun(context.Background(), run, s.pipelineNameForNotification(context.Background(), run.PipelineID, projectID), notification.EventBuildFailed)
			}
			if s.stepRepo != nil {
				_ = s.stepRepo.FinalizeRun(context.Background(), runID, string(StatusFailed))
			}
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
				run, _ := s.repo.GetByIDAndProject(context.Background(), runID, projectID)
				_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
					"status":      StatusSucceeded,
					"finished_at": now,
					"updated_at":  now,
				})
				if run != nil {
					run.Status = StatusSucceeded
					run.FinishedAt = &now
					s.recordPipelineSignal(context.Background(), runID, run.PipelineID, projectID, StatusSucceeded, "pipeline.succeeded", "Pipeline run succeeded", "")
					s.notifyPipelineRun(context.Background(), run, s.pipelineNameForNotification(context.Background(), run.PipelineID, projectID), notification.EventBuildSuccess)
				}
				if s.stepRepo != nil {
					_ = s.stepRepo.FinalizeRun(context.Background(), runID, string(StatusSucceeded))
				}
				return
			case "Failed":
				now := time.Now()
				run, _ := s.repo.GetByIDAndProject(context.Background(), runID, projectID)
				_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
					"status":        StatusFailed,
					"finished_at":   now,
					"updated_at":    now,
					"error_message": "Pipeline execution failed",
				})
				if run != nil {
					run.Status = StatusFailed
					run.FinishedAt = &now
					run.ErrorMessage = ptr("Pipeline execution failed")
					s.recordPipelineSignal(context.Background(), runID, run.PipelineID, projectID, StatusFailed, "pipeline.failed", "Pipeline run failed", "Pipeline execution failed")
					s.notifyPipelineRun(context.Background(), run, s.pipelineNameForNotification(context.Background(), run.PipelineID, projectID), notification.EventBuildFailed)
				}
				if s.stepRepo != nil {
					_ = s.stepRepo.FinalizeRun(context.Background(), runID, string(StatusFailed))
				}
				return
			case "Cancelled":
				now := time.Now()
				run, _ := s.repo.GetByIDAndProject(context.Background(), runID, projectID)
				_ = s.repo.Update(context.Background(), runID, projectID, map[string]interface{}{
					"status":      StatusCancelled,
					"finished_at": now,
					"updated_at":  now,
				})
				if run != nil {
					s.recordPipelineSignal(context.Background(), runID, run.PipelineID, projectID, StatusCancelled, "pipeline.cancelled", "Pipeline run was cancelled", "")
				}
				if s.stepRepo != nil {
					_ = s.stepRepo.FinalizeRun(context.Background(), runID, string(StatusCancelled))
				}
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
	err = s.repo.Update(ctx, runID, projectID, map[string]interface{}{
		"status":      StatusCancelled,
		"finished_at": now,
		"updated_at":  now,
	})
	if err == nil && s.stepRepo != nil {
		_ = s.stepRepo.FinalizeRun(ctx, runID, string(StatusCancelled))
	}
	if err == nil {
		s.recordPipelineSignal(ctx, runID, run.PipelineID, projectID, StatusCancelled, "pipeline.cancelled", "Pipeline run was cancelled", "")
	}
	return err
}

func (s *Service) notifyPipelineRun(ctx context.Context, run *PipelineRun, pipelineName string, event notification.EventType) {
	if s.notifications == nil || run == nil {
		return
	}
	payload := map[string]any{
		"pipelineId":   run.PipelineID,
		"pipelineName": pipelineName,
		"runId":        run.ID,
		"runNumber":    run.RunNumber,
		"status":       string(run.Status),
	}
	if run.GitBranch != nil {
		payload["branch"] = *run.GitBranch
	}
	if run.GitCommit != nil {
		payload["commitSha"] = *run.GitCommit
	}
	if run.TriggeredBy != nil {
		payload["triggeredBy"] = *run.TriggeredBy
	}
	if run.ErrorMessage != nil {
		payload["errorMessage"] = *run.ErrorMessage
	}
	if run.StartedAt != nil && run.FinishedAt != nil && run.FinishedAt.After(*run.StartedAt) {
		payload["duration"] = run.FinishedAt.Sub(*run.StartedAt).Round(time.Second).String()
	}
	if err := s.notifications.SendWebhook(ctx, run.ProjectID, event, payload); err != nil {
		slog.Warn("failed to send pipeline notification", slog.Any("error", err), slog.String("runID", run.ID), slog.String("event", string(event)))
	}
}

func (s *Service) pipelineNameForNotification(ctx context.Context, pipelineID, projectID string) string {
	if s.pipelineRepo == nil {
		return pipelineID
	}
	p, err := s.pipelineRepo.GetByIDAndProject(ctx, pipelineID, projectID)
	if err != nil || p == nil || p.Name == "" {
		return pipelineID
	}
	return p.Name
}

func (s *Service) recordPipelineSignal(ctx context.Context, runID, pipelineID, projectID string, status RunStatus, reason, message, errorMessage string) {
	if s.signals == nil || pipelineID == "" || projectID == "" {
		return
	}
	sigStatus, severity := pipelineSignalStatus(status)
	staleAfter := time.Now().Add(30 * time.Minute)
	value := map[string]any{
		"runId":        runID,
		"pipelineId":   pipelineID,
		"runStatus":    string(status),
		"errorMessage": errorMessage,
	}
	if _, err := s.signals.Record(ctx, signal.RecordInput{
		ProjectID:     projectID,
		TargetType:    signal.TargetPipeline,
		TargetID:      pipelineID,
		Source:        "pipeline-run",
		Status:        sigStatus,
		Severity:      severity,
		Reason:        reason,
		Message:       message,
		ObservedValue: value,
		StaleAfter:    &staleAfter,
	}); err != nil {
		slog.Warn("failed to record pipeline health signal", slog.Any("error", err), slog.String("runID", runID), slog.String("pipelineID", pipelineID))
	}
}

func pipelineSignalStatus(status RunStatus) (signal.Status, signal.Severity) {
	switch status {
	case StatusSucceeded:
		return signal.StatusHealthy, signal.SeverityInfo
	case StatusFailed:
		return signal.StatusDegraded, signal.SeverityCritical
	case StatusCancelled:
		return signal.StatusWarning, signal.SeverityWarning
	default:
		return signal.StatusUnknown, signal.SeverityInfo
	}
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

func (s *Service) GetStepExecutions(ctx context.Context, projectID, pipelineID, runID string) (*StepExecutionListResponse, error) {
	if s.stepRepo == nil {
		return &StepExecutionListResponse{Items: []StepExecutionResponse{}}, nil
	}
	if _, err := s.repo.GetByIDProjectPipeline(ctx, runID, projectID, pipelineID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeRunNotFound, "运行记录不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询运行记录失败", err.Error())
	}
	rows, err := s.stepRepo.ListByPipelineRun(ctx, runID)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "查询步骤执行记录失败", err.Error())
	}
	items := make([]StepExecutionResponse, 0, len(rows))
	for i := range rows {
		items = append(items, toStepExecutionResponse(rows[i]))
	}
	return &StepExecutionListResponse{Items: items}, nil
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
