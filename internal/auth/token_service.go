package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/xjy/zcid/internal/audit"
	"github.com/xjy/zcid/pkg/middleware"
	"github.com/xjy/zcid/pkg/response"
)

const (
	PersonalTokenPrefix = "zcid_pat_"
	ProjectTokenPrefix  = "zcid_proj_"
	MaxAccessTokenTTL   = 366 * 24 * time.Hour
	accessTokenBytes    = 32
)

var ErrAccessTokenNotFound = errors.New("access token not found")

type TokenRepository interface {
	CreateAccessToken(ctx context.Context, token *AccessToken) error
	ListAccessTokens(ctx context.Context, ownerUserID string, includeProject bool) ([]*AccessToken, error)
	FindAccessTokenByID(ctx context.Context, tokenID string) (*AccessToken, error)
	FindAccessTokenByHash(ctx context.Context, tokenHash string) (*AccessToken, error)
	RevokeAccessToken(ctx context.Context, tokenID string, actorID string, revokedAt time.Time) error
	UpdateAccessTokenLastUsed(ctx context.Context, tokenID string, usedAt time.Time) error
}

type TokenAuditRecorder interface {
	LogAuthSecurityEvent(ctx context.Context, event audit.AuthSecurityEvent)
}

type TokenService struct {
	repo  TokenRepository
	audit TokenAuditRecorder
	now   func() time.Time
}

type CreateAccessTokenInput struct {
	Name      string
	Type      AccessTokenType
	Scopes    []string
	ExpiresAt time.Time
	UserID    string
	ProjectID string
	ActorID   string
	IP        string
}

type AccessTokenWithSecret struct {
	Token *AccessToken
	Raw   string
}

type TokenPrincipal struct {
	TokenID   string
	TokenType AccessTokenType
	UserID    string
	ProjectID string
	Scopes    []string
}

func NewTokenService(repo TokenRepository, auditRecorder TokenAuditRecorder) *TokenService {
	return &TokenService{repo: repo, audit: auditRecorder, now: time.Now}
}

func (s *TokenService) Create(ctx context.Context, input CreateAccessTokenInput) (*AccessTokenWithSecret, error) {
	name := strings.TrimSpace(input.Name)
	actorID := strings.TrimSpace(input.ActorID)
	if name == "" || actorID == "" {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "name and actor are required")
	}
	if input.Type != AccessTokenTypePersonal && input.Type != AccessTokenTypeProject {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "invalid token type")
	}

	scopes, err := NormalizeTokenScopes(input.Scopes)
	if err != nil {
		return nil, err
	}
	now := s.now()
	if input.ExpiresAt.IsZero() || !input.ExpiresAt.After(now) {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "expiry must be in the future")
	}
	if input.ExpiresAt.After(now.Add(MaxAccessTokenTTL)) {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "expiry exceeds maximum lifetime")
	}

	raw, prefix, err := generateAccessTokenSecret(input.Type)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	record := &AccessToken{
		TokenType:   input.Type,
		Name:        name,
		TokenPrefix: prefix,
		TokenHash:   hashAccessToken(raw),
		Scopes:      EncodeTokenScopes(scopes),
		CreatedBy:   actorID,
		ExpiresAt:   input.ExpiresAt,
	}
	if input.Type == AccessTokenTypePersonal {
		userID := strings.TrimSpace(input.UserID)
		if userID == "" {
			userID = actorID
		}
		record.UserID = &userID
	} else {
		projectID := strings.TrimSpace(input.ProjectID)
		if projectID == "" {
			return nil, response.NewBizError(response.CodeValidation, "invalid request", "project id is required")
		}
		record.ProjectID = &projectID
	}

	if err := s.repo.CreateAccessToken(ctx, record); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	s.logTokenEvent(ctx, audit.ResultSuccess, "auth.token_created", record, actorID, input.IP, "")
	return &AccessTokenWithSecret{Token: record, Raw: raw}, nil
}

func (s *TokenService) List(ctx context.Context, actorID string, includeProject bool) ([]*AccessToken, error) {
	list, err := s.repo.ListAccessTokens(ctx, strings.TrimSpace(actorID), includeProject)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	return list, nil
}

func (s *TokenService) Revoke(ctx context.Context, tokenID string, actorID string, actorRole string, ip string) error {
	tokenID = strings.TrimSpace(tokenID)
	actorID = strings.TrimSpace(actorID)
	if tokenID == "" || actorID == "" {
		return response.NewBizError(response.CodeValidation, "invalid request", "token id and actor are required")
	}
	record, err := s.repo.FindAccessTokenByID(ctx, tokenID)
	if errors.Is(err, ErrAccessTokenNotFound) {
		return response.NewBizError(response.CodeNotFound, "access token not found", "")
	}
	if err != nil {
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	if strings.TrimSpace(actorRole) != string(SystemRoleAdmin) {
		if record.TokenType != AccessTokenTypePersonal || record.UserID == nil || *record.UserID != actorID {
			return response.NewBizError(response.CodeForbidden, "forbidden", "cannot revoke another principal's token")
		}
	}
	if err := s.repo.RevokeAccessToken(ctx, tokenID, actorID, s.now()); err != nil {
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	s.logTokenEvent(ctx, audit.ResultSuccess, "auth.token_revoked", record, actorID, ip, "")
	return nil
}

func (s *TokenService) Validate(ctx context.Context, raw string, requiredScope string, ip string) (*TokenPrincipal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || (!strings.HasPrefix(raw, PersonalTokenPrefix) && !strings.HasPrefix(raw, ProjectTokenPrefix)) {
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	record, err := s.repo.FindAccessTokenByHash(ctx, hashAccessToken(raw))
	if err != nil {
		s.logTokenFailure(ctx, "unknown", requiredScope, ip, "not_found")
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	now := s.now()
	if record.RevokedAt != nil {
		s.logTokenEvent(ctx, audit.ResultFailure, "auth.token_used", record, "", ip, "revoked")
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	if !now.Before(record.ExpiresAt) {
		s.logTokenEvent(ctx, audit.ResultFailure, "auth.token_used", record, "", ip, "expired")
		return nil, response.NewBizError(response.CodeTokenExpired, "token expired", "")
	}
	scopes := DecodeTokenScopes(record.Scopes)
	if requiredScope != "" && !TokenHasScope(scopes, requiredScope) {
		s.logTokenEvent(ctx, audit.ResultFailure, "auth.token_used", record, "", ip, "missing_scope")
		return nil, response.NewBizError(response.CodeForbidden, "forbidden", "missing token scope")
	}
	if err := s.repo.UpdateAccessTokenLastUsed(ctx, record.ID, now); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	s.logTokenEvent(ctx, audit.ResultSuccess, "auth.token_used", record, "", ip, "")
	principal := &TokenPrincipal{TokenID: record.ID, TokenType: record.TokenType, Scopes: scopes}
	if record.UserID != nil {
		principal.UserID = *record.UserID
	}
	if record.ProjectID != nil {
		principal.ProjectID = *record.ProjectID
	}
	return principal, nil
}

func (s *TokenService) ValidateProgrammaticToken(ctx context.Context, raw string, requiredScope string, ip string) (*middleware.ProgrammaticTokenPrincipal, error) {
	principal, err := s.Validate(ctx, raw, requiredScope, ip)
	if err != nil {
		return nil, err
	}
	return &middleware.ProgrammaticTokenPrincipal{
		TokenID:   principal.TokenID,
		TokenType: string(principal.TokenType),
		UserID:    principal.UserID,
		ProjectID: principal.ProjectID,
		Scopes:    principal.Scopes,
	}, nil
}

func (s *TokenService) logTokenFailure(ctx context.Context, tokenID string, scope string, ip string, reason string) {
	if s.audit == nil {
		return
	}
	s.audit.LogAuthSecurityEvent(ctx, audit.AuthSecurityEvent{
		Action: "auth.token_used",
		Result: audit.ResultFailure,
		IP:     ip,
		Detail: audit.AuthSecurityDetail{EventType: "auth.token_used", TokenID: tokenID, Reason: reason, Fields: map[string]any{"scope": scope}},
	})
}

func (s *TokenService) logTokenEvent(ctx context.Context, result string, action string, record *AccessToken, actorID string, ip string, reason string) {
	if s.audit == nil || record == nil {
		return
	}
	userID := actorID
	if userID == "" && record.UserID != nil {
		userID = *record.UserID
	}
	detail := audit.AuthSecurityDetail{
		EventType: action,
		TokenType: string(record.TokenType),
		TokenID:   record.ID,
		TokenName: record.Name,
		Reason:    reason,
		Fields:    map[string]any{"scopes": DecodeTokenScopes(record.Scopes)},
	}
	if record.ProjectID != nil {
		detail.TargetProject = *record.ProjectID
	}
	s.audit.LogAuthSecurityEvent(ctx, audit.AuthSecurityEvent{
		UserID:     userID,
		Action:     action,
		ResourceID: record.ID,
		Result:     result,
		IP:         ip,
		Detail:     detail,
	})
}

func generateAccessTokenSecret(tokenType AccessTokenType) (string, string, error) {
	raw := make([]byte, accessTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	prefix := PersonalTokenPrefix
	if tokenType == AccessTokenTypeProject {
		prefix = ProjectTokenPrefix
	}
	secret := prefix + base64.RawURLEncoding.EncodeToString(raw)
	return secret, prefix, nil
}

func hashAccessToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
