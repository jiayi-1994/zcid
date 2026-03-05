package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xjy/zcid/pkg/response"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindUserByID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, userID string, updates map[string]any) error
	ListUsers(ctx context.Context) ([]*User, error)
	StoreRefreshToken(ctx context.Context, userID string, refreshToken string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID string) error
	UpsertUserRolePolicy(ctx context.Context, userID string, role SystemRole) error
	PublishPolicyUpdate(ctx context.Context) error
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: []byte(strings.TrimSpace(jwtSecret)),
		now:       time.Now,
	}
}

type Service struct {
	repo      Repository
	jwtSecret []byte
	now       func() time.Time
}

type tokenClaims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePasswordHash(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *Service) Login(ctx context.Context, username string, password string) (*TokenPair, error) {
	if !s.hasJWTSecret() {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	if user == nil || !ComparePasswordHash(user.PasswordHash, password) {
		return nil, response.NewBizError(response.CodeUnauthorized, "invalid username or password", "")
	}

	if user.Status == UserStatusDisabled {
		return nil, response.NewBizError(response.CodeAccountDisabled, "account disabled", "")
	}

	accessToken, err := s.signToken(user.ID, user.Username, string(user.Role), "access", AccessTokenTTL)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	refreshToken, err := s.signToken(user.ID, user.Username, string(user.Role), "refresh", RefreshTokenTTL)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	if err := s.repo.StoreRefreshToken(ctx, user.ID, refreshToken, RefreshTokenTTL); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) CreateUser(ctx context.Context, username string, password string, status string, role string) (*User, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "username and password are required")
	}

	userStatus, err := parseUserStatus(status)
	if err != nil {
		return nil, err
	}

	userRole, err := parseSystemRole(role)
	if err != nil {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	user := &User{
		Username:     username,
		PasswordHash: hash,
		Status:       userStatus,
		Role:         userRole,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return nil, response.NewBizError(response.CodeConflict, "username already exists", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	if err := s.repo.UpsertUserRolePolicy(ctx, user.ID, user.Role); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	if err := s.repo.PublishPolicyUpdate(ctx); err != nil {
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, userID string, username *string, password *string, status *string, role *string) (*User, error) {
	updates := map[string]any{}

	if username != nil {
		trimmed := strings.TrimSpace(*username)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "invalid request", "username cannot be empty")
		}
		updates["username"] = trimmed
	}

	if password != nil {
		trimmed := strings.TrimSpace(*password)
		if trimmed == "" {
			return nil, response.NewBizError(response.CodeValidation, "invalid request", "password cannot be empty")
		}
		hash, err := HashPassword(trimmed)
		if err != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
		updates["password_hash"] = hash
	}

	if status != nil {
		userStatus, err := parseUserStatus(*status)
		if err != nil {
			return nil, err
		}
		updates["status"] = userStatus
	}

	if role != nil {
		parsedRole, err := parseSystemRole(*role)
		if err != nil {
			return nil, err
		}
		updates["role"] = parsedRole
	}

	if len(updates) == 0 {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "at least one field is required")
	}

	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return nil, response.NewBizError(response.CodeNotFound, "user not found", "")
		case errors.Is(err, ErrUsernameTaken):
			return nil, response.NewBizError(response.CodeConflict, "username already exists", "")
		default:
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	if value, ok := updates["role"]; ok {
		newRole, ok := value.(SystemRole)
		if !ok {
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
		if err := s.repo.UpsertUserRolePolicy(ctx, userID, newRole); err != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
		if err := s.repo.PublishPolicyUpdate(ctx); err != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	if value, ok := updates["status"]; ok && value == UserStatusDisabled {
		if err := s.repo.DeleteRefreshToken(ctx, userID); err != nil {
			return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
		}
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, response.NewBizError(response.CodeNotFound, "user not found", "")
		}
		return nil, response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return user, nil
}

func (s *Service) DisableUser(ctx context.Context, userID string) (*User, error) {
	disabled := string(UserStatusDisabled)
	return s.UpdateUser(ctx, userID, nil, nil, &disabled, nil)
}

func (s *Service) AssignSystemRole(ctx context.Context, userID string, role string) (*User, error) {
	trimmedRole := strings.TrimSpace(role)
	if trimmedRole == "" {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "role is required")
	}
	return s.UpdateUser(ctx, userID, nil, nil, nil, &trimmedRole)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, error) {
	if !s.hasJWTSecret() {
		return "", response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	claims, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return "", err
	}

	storedToken, err := s.repo.GetRefreshToken(ctx, claims.Subject)
	if errors.Is(err, ErrRefreshSessionNotFound) {
		return "", response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	if err != nil {
		return "", response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	if subtle.ConstantTimeCompare([]byte(storedToken), []byte(refreshToken)) != 1 {
		return "", response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}

	user, err := s.repo.FindUserByID(ctx, claims.Subject)
	if errors.Is(err, ErrUserNotFound) {
		return "", response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	if err != nil {
		return "", response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}
	if user.Status == UserStatusDisabled {
		return "", response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}

	accessToken, err := s.signToken(claims.Subject, claims.Username, string(user.Role), "access", AccessTokenTTL)
	if err != nil {
		return "", response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return accessToken, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if !s.hasJWTSecret() {
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	claims, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return err
	}

	if err := s.repo.DeleteRefreshToken(ctx, claims.Subject); err != nil {
		return response.NewBizError(response.CodeInternalServerError, "internal server error", "")
	}

	return nil
}

func (s *Service) signToken(userID string, username string, role string, tokenType string, ttl time.Duration) (string, error) {
	now := s.now()
	claims := tokenClaims{
		Username:  username,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *Service) hasJWTSecret() bool {
	return len(s.jwtSecret) > 0
}

func (s *Service) parseToken(tokenString string, expectedType string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, response.NewBizError(response.CodeTokenExpired, "token expired", "")
		}
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}

	if !token.Valid {
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	if claims.TokenType != expectedType {
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}
	if claims.Subject == "" {
		return nil, response.NewBizError(response.CodeUnauthorized, "unauthorized", "")
	}

	return claims, nil
}

func parseUserStatus(raw string) (UserStatus, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return UserStatusActive, nil
	}

	switch UserStatus(value) {
	case UserStatusActive, UserStatusDisabled:
		return UserStatus(value), nil
	default:
		return "", response.NewBizError(response.CodeValidation, "invalid request", "invalid status")
	}
}

func parseSystemRole(raw string) (SystemRole, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return SystemRoleMember, nil
	}

	switch SystemRole(value) {
	case SystemRoleAdmin, SystemRoleProjectAdmin, SystemRoleMember:
		return SystemRole(value), nil
	default:
		return "", response.NewBizError(response.CodeValidation, "invalid request", "invalid role")
	}
}

func (s *Service) ListUsers(ctx context.Context) ([]*User, error) {
	return s.repo.ListUsers(ctx)
}
