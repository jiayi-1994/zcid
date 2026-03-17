package variable

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

type Repository interface {
	Create(ctx context.Context, v *Variable) error
	GetByID(ctx context.Context, id string) (*Variable, error)
	ListByProject(ctx context.Context, projectID string) ([]Variable, int64, error)
	ListGlobal(ctx context.Context) ([]Variable, int64, error)
	ListByPipelineScope(ctx context.Context, projectID, pipelineID string) ([]Variable, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, id string) error
	ListGlobalAndProject(ctx context.Context, projectID string) ([]Variable, error)
}

type Service struct {
	repo   Repository
	crypto *crypto.AESCrypto
}

func NewService(repo Repository, aesCrypto *crypto.AESCrypto) *Service {
	return &Service{repo: repo, crypto: aesCrypto}
}

func (s *Service) CreateVariable(ctx context.Context, scope VariableScope, projectID *string, pipelineID *string, req CreateVariableRequest, createdBy string) (*Variable, error) {
	varType := TypePlain
	if req.VarType == string(TypeSecret) {
		varType = TypeSecret
	}

	value := req.Value
	if varType == TypeSecret {
		if s.crypto == nil {
			return nil, response.NewBizError(response.CodeDecryptFailed, "加密服务未配置", "encryption key not set")
		}
		encrypted, err := s.crypto.Encrypt(value)
		if err != nil {
			return nil, response.NewBizError(response.CodeDecryptFailed, "加密失败", err.Error())
		}
		value = encrypted
	}

	v := &Variable{
		ID:          uuid.New().String(),
		Scope:       scope,
		ProjectID:   projectID,
		PipelineID:  pipelineID,
		Key:         req.Key,
		Value:       value,
		VarType:     varType,
		Description: req.Description,
		Status:      StatusActive,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		if errors.Is(err, ErrKeyDuplicate) {
			return nil, response.NewBizError(response.CodeVarDuplicate, "变量名已存在", "")
		}
		return nil, err
	}

	return v, nil
}

func (s *Service) GetVariable(ctx context.Context, id string) (*Variable, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "变量不存在", "")
		}
		return nil, err
	}
	return v, nil
}

func (s *Service) ListProjectVariables(ctx context.Context, projectID string) ([]Variable, int64, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) ListPipelineVariables(ctx context.Context, projectID, pipelineID string) ([]Variable, int64, error) {
	vars, err := s.repo.ListByPipelineScope(ctx, projectID, pipelineID)
	if err != nil {
		return nil, 0, err
	}
	return vars, int64(len(vars)), nil
}

func (s *Service) ListGlobalVariables(ctx context.Context) ([]Variable, int64, error) {
	return s.repo.ListGlobal(ctx)
}

func (s *Service) UpdateVariable(ctx context.Context, id string, req UpdateVariableRequest, isSecret bool) error {
	updates := make(map[string]interface{})
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Value != nil {
		value := *req.Value
		if isSecret {
			if s.crypto == nil {
				return response.NewBizError(response.CodeDecryptFailed, "加密服务未配置", "encryption key not set")
			}
			encrypted, err := s.crypto.Encrypt(value)
			if err != nil {
				return response.NewBizError(response.CodeDecryptFailed, "加密失败", err.Error())
			}
			value = encrypted
		}
		updates["value"] = value
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotFound, "变量不存在", "")
		}
		if errors.Is(err, ErrKeyDuplicate) {
			return response.NewBizError(response.CodeVarDuplicate, "变量名已存在", "")
		}
		return err
	}
	return nil
}

func (s *Service) DeleteVariable(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeNotFound, "变量不存在", "")
		}
		return err
	}
	return nil
}

func (s *Service) GetMergedVariables(ctx context.Context, projectID string) ([]Variable, error) {
	vars, err := s.repo.ListGlobalAndProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]Variable)
	for _, v := range vars {
		if _, exists := merged[v.Key]; !exists || v.Scope == ScopeProject {
			merged[v.Key] = v
		}
	}

	result := make([]Variable, 0, len(merged))
	for _, v := range merged {
		result = append(result, v)
	}
	return result, nil
}

// GetMergedVariablesWithPipeline returns merged variables including pipeline scope (pipeline overrides project overrides global).
func (s *Service) GetMergedVariablesWithPipeline(ctx context.Context, projectID, pipelineID string) ([]Variable, error) {
	base, err := s.repo.ListGlobalAndProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	merged := make(map[string]Variable)
	for _, v := range base {
		merged[v.Key] = v
	}
	if pipelineID != "" {
		pipeVars, err := s.repo.ListByPipelineScope(ctx, projectID, pipelineID)
		if err == nil {
			for _, v := range pipeVars {
				merged[v.Key] = v
			}
		}
	}
	result := make([]Variable, 0, len(merged))
	for _, v := range merged {
		result = append(result, v)
	}
	return result, nil
}

// ResolveVariables returns merged variables with secrets decrypted (for internal use only).
// If pipelineID is non-empty, pipeline-scope variables override project/global.
func (s *Service) ResolveVariables(ctx context.Context, projectID string, pipelineID ...string) ([]Variable, error) {
	pid := ""
	if len(pipelineID) > 0 {
		pid = pipelineID[0]
	}
	merged, err := s.GetMergedVariablesWithPipeline(ctx, projectID, pid)
	if err != nil {
		return nil, err
	}

	for i, v := range merged {
		if v.VarType == TypeSecret && s.crypto != nil {
			decrypted, err := s.crypto.Decrypt(v.Value)
			if err != nil {
				slog.Warn("变量解密失败", slog.String("key", v.Key), slog.Any("error", err))
				continue
			}
			merged[i].Value = decrypted
		}
	}
	return merged, nil
}

// FilterForRole filters out secret variables for non-admin/non-project_admin users (FR5).
func FilterForRole(vars []Variable, role string) []Variable {
	if role == "admin" || role == "project_admin" {
		return vars
	}
	filtered := make([]Variable, 0)
	for _, v := range vars {
		if v.VarType != TypeSecret {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
