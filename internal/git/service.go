package git

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/xjy/zcid/pkg/cache"
	"github.com/xjy/zcid/pkg/crypto"
	"github.com/xjy/zcid/pkg/gitprovider"
	"github.com/xjy/zcid/pkg/response"
)

type Service struct {
	repo   Repository
	crypto *crypto.AESCrypto
	cache  *cache.RedisCache
}

func NewService(repo Repository, aesCrypto *crypto.AESCrypto) *Service {
	return &Service{repo: repo, crypto: aesCrypto}
}

// SetCache injects the Redis cache for Git repo/branch list caching.
func (s *Service) SetCache(c *cache.RedisCache) {
	s.cache = c
}

func (s *Service) CreateConnection(req CreateConnectionRequest, createdBy string) (*GitConnection, error) {
	if !gitprovider.ValidProvider(req.ProviderType) {
		return nil, response.NewBizError(response.CodeGitProviderUnsupported, "不支持的 Git 提供商", req.ProviderType)
	}

	encToken, err := s.encryptToken(req.AccessToken)
	if err != nil {
		return nil, err
	}

	var encRefresh *string
	if req.RefreshToken != "" {
		enc, err := s.encryptToken(req.RefreshToken)
		if err != nil {
			return nil, err
		}
		encRefresh = &enc
	}

	tokenType := TokenPAT
	if req.TokenType == string(TokenOAuth) {
		tokenType = TokenOAuth
	}

	webhookSecret := generateWebhookSecret()
	encWebhookSecret, err := s.encryptToken(webhookSecret)
	if err != nil {
		return nil, err
	}

	conn := &GitConnection{
		ID:            uuid.New().String(),
		Name:          req.Name,
		ProviderType:  req.ProviderType,
		ServerURL:     req.ServerURL,
		AccessToken:   encToken,
		RefreshToken:  encRefresh,
		TokenType:     tokenType,
		WebhookSecret: encWebhookSecret,
		Status:        StatusConnected,
		Description:   req.Description,
		CreatedBy:     createdBy,
	}

	if err := s.repo.Create(conn); err != nil {
		if errors.Is(err, ErrNameDuplicate) {
			return nil, response.NewBizError(response.CodeGitNameDuplicate, "连接名称已存在", req.Name)
		}
		return nil, err
	}

	conn.PlainTokenMask = maskToken(req.AccessToken)
	return conn, nil
}

func (s *Service) GetConnection(id string) (*GitConnection, error) {
	conn, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return nil, err
	}
	if plain, decErr := s.decryptToken(conn.AccessToken); decErr == nil {
		conn.PlainTokenMask = maskToken(plain)
	}
	return conn, nil
}

func (s *Service) ListConnections() ([]GitConnection, int64, error) {
	return s.repo.List()
}

func (s *Service) UpdateConnection(id string, req UpdateConnectionRequest) error {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.AccessToken != nil {
		enc, err := s.encryptToken(*req.AccessToken)
		if err != nil {
			return err
		}
		updates["access_token"] = enc
		updates["status"] = string(StatusConnected)
	}
	if req.RefreshToken != nil {
		enc, err := s.encryptToken(*req.RefreshToken)
		if err != nil {
			return err
		}
		updates["refresh_token"] = enc
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.repo.Update(id, updates); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		if errors.Is(err, ErrNameDuplicate) {
			return response.NewBizError(response.CodeGitNameDuplicate, "连接名称已存在", "")
		}
		return err
	}
	return nil
}

func (s *Service) DeleteConnection(id string) error {
	if err := s.repo.SoftDelete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return err
	}
	return nil
}

func (s *Service) TestConnection(ctx context.Context, id string) (*TestConnectionResponse, error) {
	conn, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return nil, err
	}

	plainToken, err := s.decryptToken(conn.AccessToken)
	if err != nil {
		return &TestConnectionResponse{Success: false, Message: "Token 解密失败: " + err.Error()}, nil
	}

	provider, err := gitprovider.New(gitprovider.ProviderType(conn.ProviderType), conn.ServerURL, plainToken)
	if err != nil {
		return &TestConnectionResponse{Success: false, Message: "创建 Provider 失败: " + err.Error()}, nil
	}

	if err := provider.TestConnection(ctx); err != nil {
		if errors.Is(err, gitprovider.ErrAuthFailed) {
			_ = s.repo.Update(id, map[string]interface{}{"status": string(StatusTokenExpired)})
			return &TestConnectionResponse{Success: false, Message: "认证失败，Token 可能已过期或被撤销"}, nil
		}
		slog.Warn("Git 连接测试失败", slog.String("connectionId", id), slog.Any("error", err))
		_ = s.repo.Update(id, map[string]interface{}{"status": string(StatusDisconnected)})
		return &TestConnectionResponse{Success: false, Message: "连接失败: " + err.Error()}, nil
	}

	if conn.Status != StatusConnected {
		_ = s.repo.Update(id, map[string]interface{}{"status": string(StatusConnected)})
	}
	return &TestConnectionResponse{Success: true, Message: "连接正常"}, nil
}

// GetDecryptedToken returns the decrypted access token for internal use (e.g. listing repos).
func (s *Service) GetDecryptedToken(conn *GitConnection) (string, error) {
	return s.decryptToken(conn.AccessToken)
}

func (s *Service) ListRepos(ctx context.Context, connID string, page, pageSize int, refresh bool) ([]gitprovider.Repository, int, error) {
	conn, err := s.repo.GetByID(connID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, 0, response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return nil, 0, err
	}
	if conn.Status != StatusConnected {
		return nil, 0, response.NewBizError(response.CodeGitConnectionDown, "Git 连接不可用", string(conn.Status))
	}

	cacheKey := fmt.Sprintf("repos:%s:%d:%d", connID, page, pageSize)

	if !refresh && s.cache != nil {
		if cached, cacheErr := s.cache.Get(ctx, cacheKey); cacheErr == nil {
			var result reposCacheEntry
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result.Repos, result.Total, nil
			}
		}
	}

	plainToken, err := s.decryptToken(conn.AccessToken)
	if err != nil {
		return nil, 0, err
	}

	provider, err := gitprovider.New(gitprovider.ProviderType(conn.ProviderType), conn.ServerURL, plainToken)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeGitAPIFailed, "创建 Provider 失败", err.Error())
	}

	repos, total, err := provider.ListRepos(ctx, page, pageSize)
	if err != nil {
		if errors.Is(err, gitprovider.ErrAuthFailed) {
			_ = s.repo.Update(connID, map[string]interface{}{"status": string(StatusTokenExpired)})
			return nil, 0, response.NewBizError(response.CodeGitTokenInvalid, "Token 已失效", "")
		}
		return nil, 0, response.NewBizError(response.CodeGitAPIFailed, "获取仓库列表失败", err.Error())
	}

	if s.cache != nil {
		entry := reposCacheEntry{Repos: repos, Total: total}
		if data, marshalErr := json.Marshal(entry); marshalErr == nil {
			_ = s.cache.Set(ctx, cacheKey, string(data), 0)
		}
	}

	return repos, total, nil
}

func (s *Service) ListBranches(ctx context.Context, connID, repoFullName string, refresh bool) ([]gitprovider.Branch, error) {
	conn, err := s.repo.GetByID(connID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return nil, err
	}
	if conn.Status != StatusConnected {
		return nil, response.NewBizError(response.CodeGitConnectionDown, "Git 连接不可用", string(conn.Status))
	}

	cacheKey := fmt.Sprintf("branches:%s:%s", connID, repoFullName)

	if !refresh && s.cache != nil {
		if cached, cacheErr := s.cache.Get(ctx, cacheKey); cacheErr == nil {
			var branches []gitprovider.Branch
			if json.Unmarshal([]byte(cached), &branches) == nil {
				return branches, nil
			}
		}
	}

	plainToken, err := s.decryptToken(conn.AccessToken)
	if err != nil {
		return nil, err
	}

	provider, err := gitprovider.New(gitprovider.ProviderType(conn.ProviderType), conn.ServerURL, plainToken)
	if err != nil {
		return nil, response.NewBizError(response.CodeGitAPIFailed, "创建 Provider 失败", err.Error())
	}

	branches, err := provider.ListBranches(ctx, repoFullName)
	if err != nil {
		if errors.Is(err, gitprovider.ErrAuthFailed) {
			_ = s.repo.Update(connID, map[string]interface{}{"status": string(StatusTokenExpired)})
			return nil, response.NewBizError(response.CodeGitTokenInvalid, "Token 已失效", "")
		}
		return nil, response.NewBizError(response.CodeGitAPIFailed, "获取分支列表失败", err.Error())
	}

	if s.cache != nil {
		if data, marshalErr := json.Marshal(branches); marshalErr == nil {
			_ = s.cache.Set(ctx, cacheKey, string(data), 0)
		}
	}

	return branches, nil
}

// GetWebhookSecret returns the decrypted webhook secret for a connection.
func (s *Service) GetWebhookSecret(ctx context.Context, id string) (string, error) {
	conn, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", response.NewBizError(response.CodeGitConnectionNotFound, "Git 连接不存在", "")
		}
		return "", err
	}
	if conn.WebhookSecret == "" {
		return "", response.NewBizError(response.CodeGitConnectionDown, "Webhook Secret 未配置", "")
	}
	return s.decryptToken(conn.WebhookSecret)
}

// FindConnectionByServerURL finds a connection matching the given server URL.
func (s *Service) FindConnectionByServerURL(serverURL string) (*GitConnection, error) {
	conn, err := s.repo.GetByServerURL(serverURL)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, response.NewBizError(response.CodeGitConnectionNotFound, "无匹配的 Git 连接", serverURL)
		}
		return nil, err
	}
	return conn, nil
}

// VerifyGitLabWebhook verifies a GitLab webhook token against all GitLab connections.
func (s *Service) VerifyGitLabWebhook(gitlabToken string) (*GitConnection, error) {
	conns, err := s.repo.ListByProviderType("gitlab")
	if err != nil {
		return nil, err
	}
	for i := range conns {
		secret, decErr := s.decryptToken(conns[i].WebhookSecret)
		if decErr != nil {
			continue
		}
		if gitprovider.VerifyGitLabSignature(gitlabToken, secret) == nil {
			return &conns[i], nil
		}
	}
	return nil, response.NewBizError(response.CodeGitWebhookSigInvalid, "Webhook 签名验证失败", "")
}

// VerifyGitHubWebhook verifies a GitHub webhook signature against all GitHub connections.
func (s *Service) VerifyGitHubWebhook(signature string, body []byte) (*GitConnection, error) {
	conns, err := s.repo.ListByProviderType("github")
	if err != nil {
		return nil, err
	}
	for i := range conns {
		secret, decErr := s.decryptToken(conns[i].WebhookSecret)
		if decErr != nil {
			continue
		}
		if gitprovider.VerifyGitHubSignature(signature, body, secret) == nil {
			return &conns[i], nil
		}
	}
	return nil, response.NewBizError(response.CodeGitWebhookSigInvalid, "Webhook 签名验证失败", "")
}

func generateWebhookSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return uuid.New().String()
	}
	return hex.EncodeToString(b)
}

type reposCacheEntry struct {
	Repos []gitprovider.Repository `json:"repos"`
	Total int                      `json:"total"`
}

func (s *Service) encryptToken(plaintext string) (string, error) {
	if s.crypto == nil {
		return "", response.NewBizError(response.CodeEncryptFailed, "加密服务未配置", "ZCID_ENCRYPTION_KEY not set")
	}
	encrypted, err := s.crypto.Encrypt(plaintext)
	if err != nil {
		return "", response.NewBizError(response.CodeEncryptFailed, "Token 加密失败", err.Error())
	}
	return encrypted, nil
}

func (s *Service) decryptToken(ciphertext string) (string, error) {
	if s.crypto == nil {
		return "", response.NewBizError(response.CodeDecryptFailed, "解密服务未配置", "ZCID_ENCRYPTION_KEY not set")
	}
	decrypted, err := s.crypto.Decrypt(ciphertext)
	if err != nil {
		return "", response.NewBizError(response.CodeDecryptFailed, "Token 解密失败", err.Error())
	}
	return decrypted, nil
}
