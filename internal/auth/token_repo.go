package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const accessTokenCacheTTL = 5 * time.Minute

func (r *Repo) CreateAccessToken(ctx context.Context, token *AccessToken) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if token == nil {
		return fmt.Errorf("access token is nil")
	}
	if strings.TrimSpace(token.ID) == "" {
		token.ID = uuid.NewString()
	}
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("create access token: %w", err)
	}
	return nil
}

func (r *Repo) ListAccessTokens(ctx context.Context, ownerUserID string, includeProject bool) ([]*AccessToken, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	query := r.db.WithContext(ctx).Model(&AccessToken{})
	if strings.TrimSpace(ownerUserID) != "" {
		if includeProject {
			query = query.Where("user_id = ? OR token_type = ?", ownerUserID, AccessTokenTypeProject)
		} else {
			query = query.Where("user_id = ?", ownerUserID)
		}
	}
	var tokens []*AccessToken
	if err := query.Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("list access tokens: %w", err)
	}
	return tokens, nil
}

func (r *Repo) FindAccessTokenByID(ctx context.Context, tokenID string) (*AccessToken, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	var token AccessToken
	err := r.db.WithContext(ctx).Where("id = ?", tokenID).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAccessTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find access token by id: %w", err)
	}
	return &token, nil
}

func (r *Repo) FindAccessTokenByHash(ctx context.Context, tokenHash string) (*AccessToken, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	cacheKey := accessTokenCacheKey(tokenHash)
	if r.redis != nil {
		cached, cacheErr := r.redis.Get(ctx, cacheKey).Result()
		if cacheErr == nil {
			var token AccessToken
			if err := json.Unmarshal([]byte(cached), &token); err == nil {
				return &token, nil
			}
		}
	}
	var token AccessToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAccessTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find access token by hash: %w", err)
	}
	if r.redis != nil {
		if payload, marshalErr := json.Marshal(token); marshalErr == nil {
			_ = r.redis.Set(ctx, cacheKey, payload, accessTokenCacheTTL).Err()
		}
	}
	return &token, nil
}

func (r *Repo) RevokeAccessToken(ctx context.Context, tokenID string, actorID string, revokedAt time.Time) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	updates := map[string]any{"revoked_at": revokedAt, "revoked_by": actorID, "updated_at": revokedAt}
	res := r.db.WithContext(ctx).Model(&AccessToken{}).Where("id = ? AND revoked_at IS NULL", tokenID).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("revoke access token: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrAccessTokenNotFound
	}
	if r.redis != nil {
		var token AccessToken
		if err := r.db.WithContext(ctx).Where("id = ?", tokenID).First(&token).Error; err == nil {
			_ = r.redis.Del(ctx, accessTokenCacheKey(token.TokenHash)).Err()
		}
	}
	return nil
}

func (r *Repo) UpdateAccessTokenLastUsed(ctx context.Context, tokenID string, usedAt time.Time) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := r.db.WithContext(ctx).Model(&AccessToken{}).Where("id = ?", tokenID).Updates(map[string]any{"last_used_at": usedAt, "updated_at": usedAt}).Error; err != nil {
		return fmt.Errorf("update access token last used: %w", err)
	}
	return nil
}

func accessTokenCacheKey(tokenHash string) string {
	return "auth:access-token:" + strings.TrimSpace(tokenHash)
}
