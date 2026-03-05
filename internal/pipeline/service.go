package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/response"
)

type Service struct {
	repo      Repository
	templates *TemplateRegistry
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:      repo,
		templates: NewTemplateRegistry(),
	}
}

func (s *Service) ListTemplates() []*PipelineTemplate {
	return s.templates.List()
}

func (s *Service) GetTemplate(templateID string) (*PipelineTemplate, error) {
	t := s.templates.Get(templateID)
	if t == nil {
		return nil, response.NewBizError(response.CodeNotFound, "模板不存在", fmt.Sprintf("未找到模板: %s", templateID))
	}
	return t, nil
}

func (s *Service) CreatePipeline(ctx context.Context, projectID string, req CreatePipelineRequest, createdBy string) (*Pipeline, error) {
	if req.TemplateID != "" {
		tmpl := s.templates.Get(req.TemplateID)
		if tmpl == nil {
			return nil, response.NewBizError(response.CodeNotFound, "模板不存在", fmt.Sprintf("未找到模板: %s", req.TemplateID))
		}
		if req.TemplateParams == nil {
			req.TemplateParams = make(map[string]string)
		}
		mergedParams, valErr := validateTemplateParams(tmpl, req.TemplateParams)
		if valErr != nil {
			return nil, response.NewBizError(response.CodeValidation, "模板参数校验失败", valErr.Error())
		}
		req.Config = applyTemplateParams(tmpl.Config, mergedParams)
	}

	if req.Config.SchemaVersion == "" {
		req.Config.SchemaVersion = "1.0"
	}
	if req.Config.Stages == nil {
		req.Config.Stages = []StageConfig{}
	}

	triggerType := TriggerManual
	if req.TriggerType != "" {
		tt := TriggerType(req.TriggerType)
		if !isValidTriggerType(tt) {
			return nil, response.NewBizError(response.CodeValidation, "无效的触发类型", fmt.Sprintf("不支持的触发类型: %s", req.TriggerType))
		}
		triggerType = tt
	}

	concurrencyPolicy := ConcurrencyQueue
	if req.ConcurrencyPolicy != "" {
		cp := ConcurrencyPolicy(req.ConcurrencyPolicy)
		if !isValidConcurrencyPolicy(cp) {
			return nil, response.NewBizError(response.CodeValidation, "无效的并发策略", fmt.Sprintf("不支持的并发策略: %s", req.ConcurrencyPolicy))
		}
		concurrencyPolicy = cp
	}

	p := &Pipeline{
		ID:                uuid.NewString(),
		ProjectID:         projectID,
		Name:              req.Name,
		Description:       req.Description,
		Status:            StatusDraft,
		Config:            req.Config,
		TriggerType:       triggerType,
		ConcurrencyPolicy: concurrencyPolicy,
		CreatedBy:         createdBy,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, ErrNameDuplicate) {
			return nil, response.NewBizError(response.CodePipelineNameDup, "流水线名称已存在", fmt.Sprintf("项目内已存在名为 '%s' 的流水线", req.Name))
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "创建流水线失败", err.Error())
	}

	return p, nil
}

func (s *Service) GetPipeline(ctx context.Context, id, projectID string) (*Pipeline, error) {
	p, err := s.repo.GetByIDAndProject(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询流水线失败", err.Error())
	}
	return p, nil
}

func (s *Service) ListPipelines(ctx context.Context, projectID string, page, pageSize int) ([]*Pipeline, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pipelines, total, err := s.repo.List(ctx, projectID, page, pageSize)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeInternalServerError, "查询流水线列表失败", err.Error())
	}
	return pipelines, total, nil
}

func (s *Service) UpdatePipeline(ctx context.Context, id, projectID string, req UpdatePipelineRequest) (*Pipeline, error) {
	existing, err := s.repo.GetByIDAndProject(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询流水线失败", err.Error())
	}

	updates := make(map[string]any)

	if req.Name != nil {
		exists, checkErr := s.repo.ExistsByNameAndProject(ctx, existing.ProjectID, *req.Name, id)
		if checkErr != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "检查流水线名称失败", checkErr.Error())
		}
		if exists {
			return nil, response.NewBizError(response.CodePipelineNameDup, "流水线名称已存在", fmt.Sprintf("项目内已存在名为 '%s' 的流水线", *req.Name))
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		st := PipelineStatus(*req.Status)
		if !isValidStatus(st) {
			return nil, response.NewBizError(response.CodeValidation, "无效的流水线状态", fmt.Sprintf("不支持的状态: %s", *req.Status))
		}
		updates["status"] = string(st)
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.TriggerType != nil {
		tt := TriggerType(*req.TriggerType)
		if !isValidTriggerType(tt) {
			return nil, response.NewBizError(response.CodeValidation, "无效的触发类型", fmt.Sprintf("不支持的触发类型: %s", *req.TriggerType))
		}
		updates["trigger_type"] = string(tt)
	}
	if req.ConcurrencyPolicy != nil {
		cp := ConcurrencyPolicy(*req.ConcurrencyPolicy)
		if !isValidConcurrencyPolicy(cp) {
			return nil, response.NewBizError(response.CodeValidation, "无效的并发策略", fmt.Sprintf("不支持的并发策略: %s", *req.ConcurrencyPolicy))
		}
		updates["concurrency_policy"] = string(cp)
	}

	if len(updates) == 0 {
		return existing, nil
	}

	if err := s.repo.Update(ctx, id, projectID, updates); err != nil {
		if errors.Is(err, ErrNameDuplicate) {
			return nil, response.NewBizError(response.CodePipelineNameDup, "流水线名称已存在", "")
		}
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "更新流水线失败", err.Error())
	}

	return s.repo.GetByIDAndProject(ctx, id, projectID)
}

func (s *Service) DeletePipeline(ctx context.Context, id, projectID string) error {
	if err := s.repo.SoftDelete(ctx, id, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return response.NewBizError(response.CodeInternalServerError, "删除流水线失败", err.Error())
	}
	return nil
}

func (s *Service) CopyPipeline(ctx context.Context, id, projectID, createdBy string) (*Pipeline, error) {
	source, err := s.repo.GetByIDAndProject(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodePipelineNotFound, "流水线不存在", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "查询流水线失败", err.Error())
	}

	configBytes, err := json.Marshal(source.Config)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "复制流水线配置失败", err.Error())
	}
	var copiedConfig PipelineConfig
	if err := json.Unmarshal(configBytes, &copiedConfig); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "复制流水线配置失败", err.Error())
	}

	copyName := source.Name + "-copy"
	for i := 2; ; i++ {
		exists, checkErr := s.repo.ExistsByNameAndProject(ctx, source.ProjectID, copyName, "")
		if checkErr != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "检查流水线名称失败", checkErr.Error())
		}
		if !exists {
			break
		}
		copyName = fmt.Sprintf("%s-copy-%d", source.Name, i)
	}

	copied := &Pipeline{
		ID:                uuid.NewString(),
		ProjectID:         source.ProjectID,
		Name:              copyName,
		Description:       source.Description,
		Status:            StatusDraft,
		Config:            copiedConfig,
		TriggerType:       source.TriggerType,
		ConcurrencyPolicy: source.ConcurrencyPolicy,
		CreatedBy:         createdBy,
	}

	if err := s.repo.Create(ctx, copied); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "复制流水线失败", err.Error())
	}

	return copied, nil
}

func isValidStatus(s PipelineStatus) bool {
	return s == StatusDraft || s == StatusActive || s == StatusDisabled
}

func isValidTriggerType(t TriggerType) bool {
	return t == TriggerManual || t == TriggerWebhook || t == TriggerScheduled
}

func isValidConcurrencyPolicy(c ConcurrencyPolicy) bool {
	return c == ConcurrencyQueue || c == ConcurrencyCancelOld || c == ConcurrencyReject
}
