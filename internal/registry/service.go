package registry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/response"
)

// Service implements registry business logic
type Service struct {
	repo       Repository
	crypto     *crypto.AESCrypto
	signals    *signal.Service
	httpClient *http.Client
}

// NewService creates a new Service
func NewService(repo Repository, aesCrypto *crypto.AESCrypto) *Service {
	return &Service{repo: repo, crypto: aesCrypto, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Service) SetSignalService(signals *signal.Service) {
	s.signals = signals
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
		PasswordEncrypted: passwordEnc,
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

func (s *Service) TestConnection(req TestConnectionRequest) (*TestConnectionResponse, error) {
	endpoint, err := registryPingURL(req.URL)
	if err != nil {
		s.recordRegistrySignal(req.ProjectID, false, "registry.invalid_url", err.Error())
		return nil, response.NewBizError(response.CodeValidation, "invalid registry url", err.Error())
	}
	httpReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		s.recordRegistrySignal(req.ProjectID, false, "registry.invalid_url", err.Error())
		return nil, response.NewBizError(response.CodeValidation, "invalid registry url", err.Error())
	}
	if strings.TrimSpace(req.Username) != "" || strings.TrimSpace(req.Password) != "" {
		httpReq.SetBasicAuth(strings.TrimSpace(req.Username), req.Password)
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		s.recordRegistrySignal(req.ProjectID, false, "registry.unreachable", err.Error())
		return &TestConnectionResponse{Success: false, Message: "Registry unreachable: " + err.Error()}, nil
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if success {
		s.recordRegistrySignal(req.ProjectID, true, "registry.reachable", "Registry API is reachable")
		return &TestConnectionResponse{Success: true, Message: "Registry API is reachable"}, nil
	}
	message := "Registry API returned " + resp.Status
	s.recordRegistrySignal(req.ProjectID, false, "registry.http_error", message)
	return &TestConnectionResponse{Success: false, Message: message}, nil
}

func registryPingURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url is required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("url scheme must be http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("url host is required")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v2/"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (s *Service) recordRegistrySignal(projectID string, ok bool, reason, message string) {
	if s.signals == nil || strings.TrimSpace(projectID) == "" {
		return
	}
	status := signal.StatusDegraded
	severity := signal.SeverityCritical
	if ok {
		status = signal.StatusHealthy
		severity = signal.SeverityInfo
	}
	staleAfter := time.Now().Add(10 * time.Minute)
	if _, err := s.signals.Record(context.Background(), signal.RecordInput{
		ProjectID:  strings.TrimSpace(projectID),
		TargetType: signal.TargetIntegration,
		TargetID:   "registry",
		Source:     "registry-test",
		Status:     status,
		Severity:   severity,
		Reason:     reason,
		Message:    message,
		ObservedValue: map[string]any{
			"ok":      ok,
			"message": message,
		},
		StaleAfter: &staleAfter,
	}); err != nil {
		slog.Warn("failed to record registry health signal", slog.Any("error", err), slog.String("projectID", projectID))
	}
}

// GetDefault returns the default registry for build chains
func (s *Service) GetDefault() (*Registry, error) {
	return s.repo.GetDefault()
}
