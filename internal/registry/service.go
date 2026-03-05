package registry

import (
	"errors"

	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

// Service implements registry business logic
type Service struct {
	repo   Repository
	crypto *crypto.AESCrypto
}

// NewService creates a new Service
func NewService(repo Repository, aesCrypto *crypto.AESCrypto) *Service {
	return &Service{repo: repo, crypto: aesCrypto}
}

// Create creates a new registry with encrypted password
func (s *Service) Create(req CreateRegistryRequest, createdBy string) (*Registry, error) {
	regType := RegistryType(req.Type)
	if regType == "" {
		regType = RegistryTypeHarbor
	}

	passwordEnc := ""
	if req.Password != "" {
		if s.crypto == nil {
			return nil, response.NewBizError(response.CodeEncryptFailed, "加密服务未配置", "encryption key not set")
		}
		enc, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, response.NewBizError(response.CodeEncryptFailed, "加密失败", err.Error())
		}
		passwordEnc = enc
	}

	reg := &Registry{
		Name:              req.Name,
		Type:              regType,
		URL:               req.URL,
		Username:          req.Username,
		PasswordEncrypted:  passwordEnc,
		IsDefault:         req.IsDefault,
		Status:            StatusActive,
		CreatedBy:         createdBy,
	}

	if err := s.repo.Create(reg); err != nil {
		if errors.Is(err, ErrNameDuplicate) {
			return nil, response.NewBizError(response.CodeRegistryNameDup, "Registry name already exists", req.Name)
		}
		return nil, err
	}

	if req.IsDefault {
		_ = s.repo.SetDefault(reg.ID)
	}

	return reg, nil
}

// Get retrieves a registry by ID
func (s *Service) Get(id string) (*Registry, error) {
	reg, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeRegistryNotFound, "Registry not found", "")
		}
		return nil, err
	}
	return reg, nil
}

// List returns all registries
func (s *Service) List() ([]Registry, int64, error) {
	return s.repo.List()
}

// Update updates a registry
func (s *Service) Update(id string, req UpdateRegistryRequest) (*Registry, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil && *req.Password != "" {
		if s.crypto == nil {
			return nil, response.NewBizError(response.CodeEncryptFailed, "加密服务未配置", "encryption key not set")
		}
		enc, err := s.crypto.Encrypt(*req.Password)
		if err != nil {
			return nil, response.NewBizError(response.CodeEncryptFailed, "加密失败", err.Error())
		}
		updates["password_encrypted"] = enc
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.repo.Update(id, updates); err != nil {
			if errors.Is(err, ErrNameDuplicate) {
				return nil, response.NewBizError(response.CodeRegistryNameDup, "Registry name already exists", "")
			}
			if errors.Is(err, ErrNotFound) {
				return nil, response.NewBizError(response.CodeRegistryNotFound, "Registry not found", "")
			}
			return nil, err
		}
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := s.repo.SetDefault(id); err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, response.NewBizError(response.CodeRegistryNotFound, "Registry not found", "")
			}
			return nil, err
		}
	}

	return s.Get(id)
}

// Delete soft-deletes a registry
func (s *Service) Delete(id string) error {
	if err := s.repo.SoftDelete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeRegistryNotFound, "Registry not found", "")
		}
		return err
	}
	return nil
}

// TestConnection tests connectivity to a registry.
// TODO: Integrate with Harbor/Docker Hub/GHCR API when external dependencies are available.
func (s *Service) TestConnection(req TestConnectionRequest) (*TestConnectionResponse, error) {
	// Mock implementation - always succeeds for now
	// TODO: Implement actual HTTP ping to registry URL with auth
	_ = req
	return &TestConnectionResponse{
		Success: true,
		Message: "Connection test not implemented (mock)",
	}, nil
}

// GetDefault returns the default registry for build chains
func (s *Service) GetDefault() (*Registry, error) {
	return s.repo.GetDefault()
}
