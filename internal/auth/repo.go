package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrUsernameTaken          = errors.New("username already exists")
)

const policyUpdateChannel = "rbac:policy:update"

type Repo struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepo(db *gorm.DB, redisClient *redis.Client) *Repo {
	return &Repo{db: db, redis: redisClient}
}

func (r *Repo) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return &user, nil
}

func (r *Repo) FindUserByID(ctx context.Context, userID string) (*User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	var user User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return &user, nil
}

func (r *Repo) CreateUser(ctx context.Context, user *User) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	if strings.TrimSpace(user.ID) == "" {
		user.ID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if isUniqueConstraintError(err) {
		return ErrUsernameTaken
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repo) UpdateUser(ctx context.Context, userID string, updates map[string]any) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if len(updates) == 0 {
		return nil
	}

	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(updates)
	if isUniqueConstraintError(res.Error) {
		return ErrUsernameTaken
	}
	if res.Error != nil {
		return fmt.Errorf("update user: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repo) StoreRefreshToken(ctx context.Context, userID string, refreshToken string, ttl time.Duration) error {
	if r.redis == nil {
		return fmt.Errorf("redis client is nil")
	}
	if err := r.redis.Set(ctx, refreshSessionKey(userID), refreshToken, ttl).Err(); err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *Repo) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	if r.redis == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	value, err := r.redis.Get(ctx, refreshSessionKey(userID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrRefreshSessionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}

	return value, nil
}

func (r *Repo) DeleteRefreshToken(ctx context.Context, userID string) error {
	if r.redis == nil {
		return fmt.Errorf("redis client is nil")
	}

	if err := r.redis.Del(ctx, refreshSessionKey(userID)).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *Repo) UpsertUserRolePolicy(ctx context.Context, userID string, role SystemRole) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("userID is required")
	}
	if strings.TrimSpace(string(role)) == "" {
		return fmt.Errorf("role is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM casbin_rule WHERE ptype = ? AND v0 = ?", "g", userID).Error; err != nil {
			return fmt.Errorf("delete user role policy: %w", err)
		}
		if err := tx.Exec("INSERT INTO casbin_rule (ptype, v0, v1) VALUES (?, ?, ?)", "g", userID, string(role)).Error; err != nil {
			return fmt.Errorf("insert user role policy: %w", err)
		}
		return nil
	})
}

func (r *Repo) PublishPolicyUpdate(ctx context.Context) error {
	if r.redis == nil {
		return fmt.Errorf("redis client is nil")
	}
	if err := r.redis.Publish(ctx, policyUpdateChannel, "reload").Err(); err != nil {
		return fmt.Errorf("publish policy update: %w", err)
	}
	return nil
}

func refreshSessionKey(userID string) string {
	return "auth:refresh:" + userID
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

func (r *Repo) ListUsers(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
