package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xjy/zcid/internal/audit"
	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	usersByName  map[string]*User
	usersByID    map[string]*User
	sessions     map[string]string
	bootstrap    map[string]*BootstrapToken
	findErr      error
	storeErr     error
	getErr       error
	deleteErr    error
	createErr    error
	updateErr    error
	upsertErr    error
	publishErr   error
	storedTTL    time.Duration
	storedKeyID  string
	updateCalls  int
	createCalls  int
	publishCalls int
	lastUserID   string
	lastUpdateKV map[string]any
	lastPolicyID string
	lastPolicy   SystemRole
	accessTokens map[string]*AccessToken
}

type mockAuthAuditRecorder struct {
	events []audit.AuthSecurityEvent
}

func (m *mockAuthAuditRecorder) LogAuthSecurityEvent(_ context.Context, event audit.AuthSecurityEvent) {
	m.events = append(m.events, event)
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		usersByName:  map[string]*User{},
		usersByID:    map[string]*User{},
		sessions:     map[string]string{},
		bootstrap:    map[string]*BootstrapToken{},
		accessTokens: map[string]*AccessToken{},
	}
}

func (m *mockRepo) seedUser(user *User) {
	copy := *user
	m.usersByName[user.Username] = &copy
	m.usersByID[user.ID] = &copy
}

func (m *mockRepo) FindUserByUsername(_ context.Context, username string) (*User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if user, ok := m.usersByName[username]; ok {
		copy := *user
		return &copy, nil
	}
	return nil, nil
}

func (m *mockRepo) FindUserByID(_ context.Context, userID string) (*User, error) {
	if user, ok := m.usersByID[userID]; ok {
		copy := *user
		return &copy, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockRepo) CreateUser(_ context.Context, user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createCalls++
	if _, exists := m.usersByName[user.Username]; exists {
		return ErrUsernameTaken
	}
	if user.ID == "" {
		user.ID = "generated-id"
	}
	copy := *user
	m.usersByName[user.Username] = &copy
	m.usersByID[user.ID] = &copy
	return nil
}

func (m *mockRepo) UpdateUser(_ context.Context, userID string, updates map[string]any) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	user, ok := m.usersByID[userID]
	if !ok {
		return ErrUserNotFound
	}

	m.updateCalls++
	m.lastUserID = userID
	m.lastUpdateKV = updates

	if rawUsername, ok := updates["username"]; ok {
		newUsername := rawUsername.(string)
		if existing, exists := m.usersByName[newUsername]; exists && existing.ID != userID {
			return ErrUsernameTaken
		}
		delete(m.usersByName, user.Username)
		user.Username = newUsername
		m.usersByName[newUsername] = user
	}

	if rawHash, ok := updates["password_hash"]; ok {
		user.PasswordHash = rawHash.(string)
	}

	if rawStatus, ok := updates["status"]; ok {
		switch v := rawStatus.(type) {
		case UserStatus:
			user.Status = v
		case string:
			user.Status = UserStatus(v)
		}
	}

	if rawRole, ok := updates["role"]; ok {
		switch v := rawRole.(type) {
		case SystemRole:
			user.Role = v
		case string:
			user.Role = SystemRole(v)
		}
	}

	return nil
}

func (m *mockRepo) StoreRefreshToken(_ context.Context, userID string, refreshToken string, ttl time.Duration) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.storedTTL = ttl
	m.storedKeyID = userID
	m.sessions[userID] = refreshToken
	return nil
}

func (m *mockRepo) GetRefreshToken(_ context.Context, userID string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	token, ok := m.sessions[userID]
	if !ok {
		return "", ErrRefreshSessionNotFound
	}
	return token, nil
}

func (m *mockRepo) DeleteRefreshToken(_ context.Context, userID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.sessions, userID)
	return nil
}

func (m *mockRepo) UpsertUserRolePolicy(_ context.Context, userID string, role SystemRole) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.lastPolicyID = userID
	m.lastPolicy = role
	return nil
}

func (m *mockRepo) PublishPolicyUpdate(_ context.Context) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishCalls++
	return nil
}

func (m *mockRepo) ListUsers(_ context.Context) ([]*User, error) {
	var users []*User
	for _, u := range m.usersByID {
		copy := *u
		users = append(users, &copy)
	}
	return users, nil
}

func (m *mockRepo) CountConfiguredUsers(_ context.Context) (int64, error) {
	var count int64
	for _, user := range m.usersByID {
		if user.ID == legacyBootstrapAdminID && user.Status == UserStatusDisabled {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockRepo) DisableLegacyBootstrapAdmin(_ context.Context) error {
	user, ok := m.usersByID[legacyBootstrapAdminID]
	if !ok || user.Username != "admin" {
		return nil
	}
	user.Status = UserStatusDisabled
	if byName, ok := m.usersByName["admin"]; ok && byName.ID == legacyBootstrapAdminID {
		byName.Status = UserStatusDisabled
	}
	return nil
}

func (m *mockRepo) FindActiveBootstrapToken(_ context.Context, now time.Time) (*BootstrapToken, error) {
	for _, token := range m.bootstrap {
		if token.UsedAt == nil && token.ExpiresAt.After(now) {
			copy := *token
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) StoreBootstrapToken(_ context.Context, token *BootstrapToken) error {
	if token.ID == "" {
		token.ID = "bootstrap-id"
	}
	copy := *token
	m.bootstrap[token.TokenHash] = &copy
	return nil
}

func (m *mockRepo) FindBootstrapTokenByHash(_ context.Context, tokenHash string) (*BootstrapToken, error) {
	if token, ok := m.bootstrap[tokenHash]; ok {
		copy := *token
		return &copy, nil
	}
	return nil, nil
}

func (m *mockRepo) MarkBootstrapTokenUsed(_ context.Context, tokenID string, usedAt time.Time) error {
	for _, token := range m.bootstrap {
		if token.ID == tokenID && token.UsedAt == nil {
			t := usedAt
			token.UsedAt = &t
			return nil
		}
	}
	return errors.New("bootstrap token not found or already used")
}

func (m *mockRepo) CreateAccessToken(_ context.Context, token *AccessToken) error {
	if token.ID == "" {
		token.ID = "token-id-" + string(rune('1'+len(m.accessTokens)))
	}
	copy := *token
	m.accessTokens[token.ID] = &copy
	return nil
}

func (m *mockRepo) ListAccessTokens(_ context.Context, ownerUserID string, includeProject bool) ([]*AccessToken, error) {
	var out []*AccessToken
	for _, token := range m.accessTokens {
		if ownerUserID == "" || includeProject || (token.UserID != nil && *token.UserID == ownerUserID) {
			copy := *token
			out = append(out, &copy)
		}
	}
	return out, nil
}

func (m *mockRepo) FindAccessTokenByID(_ context.Context, tokenID string) (*AccessToken, error) {
	if token, ok := m.accessTokens[tokenID]; ok {
		copy := *token
		return &copy, nil
	}
	return nil, ErrAccessTokenNotFound
}

func (m *mockRepo) FindAccessTokenByHash(_ context.Context, tokenHash string) (*AccessToken, error) {
	for _, token := range m.accessTokens {
		if token.TokenHash == tokenHash {
			copy := *token
			return &copy, nil
		}
	}
	return nil, ErrAccessTokenNotFound
}

func (m *mockRepo) RevokeAccessToken(_ context.Context, tokenID string, actorID string, revokedAt time.Time) error {
	if token, ok := m.accessTokens[tokenID]; ok && token.RevokedAt == nil {
		t := revokedAt
		token.RevokedAt = &t
		token.RevokedBy = &actorID
		return nil
	}
	return ErrAccessTokenNotFound
}

func (m *mockRepo) UpdateAccessTokenLastUsed(_ context.Context, tokenID string, usedAt time.Time) error {
	if token, ok := m.accessTokens[tokenID]; ok {
		t := usedAt
		token.LastUsedAt = &t
		return nil
	}
	return ErrAccessTokenNotFound
}

func TestHashPasswordAndCompare(t *testing.T) {
	hash, err := HashPassword("pass123")
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	if hash == "pass123" {
		t.Fatal("password hash should not equal plain password")
	}
	if !ComparePasswordHash(hash, "pass123") {
		t.Fatal("expected password compare to succeed")
	}
	if ComparePasswordHash(hash, "wrong") {
		t.Fatal("expected password compare to fail for wrong password")
	}
}

func TestLoginSuccessStoresRefreshSession(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})

	svc := NewService(repo, "test-secret")
	pair, err := svc.Login(context.Background(), "alice", "pass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected both access and refresh token")
	}
	if repo.sessions["u-1"] != pair.RefreshToken {
		t.Fatal("expected refresh token stored in session")
	}
	if repo.storedTTL != RefreshTokenTTL {
		t.Fatalf("expected refresh ttl %v, got %v", RefreshTokenTTL, repo.storedTTL)
	}
}

func TestLoginFailureAuditUsesRequestIPFromContext(t *testing.T) {
	repo := newMockRepo()
	recorder := &mockAuthAuditRecorder{}
	svc := NewService(repo, "test-secret")
	svc.SetAuditRecorder(recorder)

	_, err := svc.Login(ContextWithRequestIP(context.Background(), "203.0.113.10"), "missing", "wrong")
	if err == nil {
		t.Fatal("expected login failure")
	}
	if len(recorder.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(recorder.events))
	}
	if recorder.events[0].Action != "auth.login_failed" {
		t.Fatalf("expected auth.login_failed event, got %s", recorder.events[0].Action)
	}
	if recorder.events[0].IP != "203.0.113.10" {
		t.Fatalf("expected request IP in audit event, got %q", recorder.events[0].IP)
	}
}

func TestLoginFailUserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.Login(context.Background(), "ghost", "pass123")
	if err == nil {
		t.Fatal("expected login error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized biz error, got %v", err)
	}
}

func TestLoginFailWrongPassword(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("right-pass")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	svc := NewService(repo, "test-secret")

	_, err := svc.Login(context.Background(), "alice", "wrong")
	if err == nil {
		t.Fatal("expected login error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized biz error, got %v", err)
	}
}

func TestLoginFailDisabledAccount(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusDisabled})
	svc := NewService(repo, "test-secret")

	_, err := svc.Login(context.Background(), "alice", "pass123")
	if err == nil {
		t.Fatal("expected disabled account error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeAccountDisabled {
		t.Fatalf("expected account disabled biz error, got %v", err)
	}
}

func TestCreateUserHashesPassword(t *testing.T) {
	repo := newMockRepo()
	recorder := &mockAuthAuditRecorder{}
	svc := NewService(repo, "test-secret")
	svc.SetAuditRecorder(recorder)

	user, err := svc.CreateUser(context.Background(), "alice", "pass123", "", "")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if user.ID == "" {
		t.Fatal("expected created user id")
	}
	if user.PasswordHash == "pass123" {
		t.Fatal("expected password to be hashed")
	}
	if user.Status != UserStatusActive {
		t.Fatalf("expected default status active, got %s", user.Status)
	}
	if !ComparePasswordHash(user.PasswordHash, "pass123") {
		t.Fatal("expected stored hash to match original password")
	}
	if len(recorder.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(recorder.events))
	}
	event := recorder.events[0]
	if event.Action != "auth.user_created" {
		t.Fatalf("expected auth.user_created event, got %s", event.Action)
	}
	if event.Result != audit.ResultSuccess {
		t.Fatalf("expected success result, got %s", event.Result)
	}
	if event.Detail.EventType != "auth.user_created" {
		t.Fatalf("expected auth.user_created event type, got %s", event.Detail.EventType)
	}
	if event.Detail.PrincipalType != "user" {
		t.Fatalf("expected user principal type, got %s", event.Detail.PrincipalType)
	}
	if event.UserID != user.ID || event.Detail.TargetUserID != user.ID {
		t.Fatalf("expected created user id in audit event, got user=%q target=%q", event.UserID, event.Detail.TargetUserID)
	}
	if _, ok := event.Detail.Fields["password"]; ok {
		t.Fatal("audit event must not include plaintext password")
	}
	if _, ok := event.Detail.Fields["passwordHash"]; ok {
		t.Fatal("audit event must not include password hash")
	}
}

func TestCreateUserDoesNotAuditWhenPolicyUpdateFails(t *testing.T) {
	repo := newMockRepo()
	repo.upsertErr = errors.New("casbin down")
	recorder := &mockAuthAuditRecorder{}
	svc := NewService(repo, "test-secret")
	svc.SetAuditRecorder(recorder)

	_, err := svc.CreateUser(context.Background(), "alice", "pass123", "", "")
	if err == nil {
		t.Fatal("expected create user to fail when policy update fails")
	}
	if len(recorder.events) != 0 {
		t.Fatalf("expected no success audit event on policy failure, got %d", len(recorder.events))
	}
}

func TestCreateUserDoesNotAuditWhenPolicyPublishFails(t *testing.T) {
	repo := newMockRepo()
	repo.publishErr = errors.New("redis down")
	recorder := &mockAuthAuditRecorder{}
	svc := NewService(repo, "test-secret")
	svc.SetAuditRecorder(recorder)

	_, err := svc.CreateUser(context.Background(), "alice", "pass123", "", "")
	if err == nil {
		t.Fatal("expected create user to fail when policy publish fails")
	}
	if len(recorder.events) != 0 {
		t.Fatalf("expected no success audit event on policy publish failure, got %d", len(recorder.events))
	}
}

func TestCreateUserConflict(t *testing.T) {
	repo := newMockRepo()
	repo.createErr = ErrUsernameTaken
	svc := NewService(repo, "test-secret")

	_, err := svc.CreateUser(context.Background(), "alice", "pass123", "", "")
	if err == nil {
		t.Fatal("expected create user conflict error")
	}

	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeConflict {
		t.Fatalf("expected conflict biz error, got %v", err)
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("old-pass")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	svc := NewService(repo, "test-secret")

	newUsername := "alice-new"
	newPassword := "new-pass"
	active := string(UserStatusActive)

	user, err := svc.UpdateUser(context.Background(), "u-1", &newUsername, &newPassword, &active, nil)
	if err != nil {
		t.Fatalf("update user failed: %v", err)
	}

	if user.Username != newUsername {
		t.Fatalf("expected username %s, got %s", newUsername, user.Username)
	}
	if user.Status != UserStatusActive {
		t.Fatalf("expected status active, got %s", user.Status)
	}
	if ComparePasswordHash(hash, newPassword) {
		t.Fatal("expected password hash to be changed")
	}
	if !ComparePasswordHash(user.PasswordHash, newPassword) {
		t.Fatal("expected updated hash to match new password")
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	newUsername := "ghost"
	_, err := svc.UpdateUser(context.Background(), "missing-user", &newUsername, nil, nil, nil)
	if err == nil {
		t.Fatal("expected not found error")
	}

	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeNotFound {
		t.Fatalf("expected not found biz error, got %v", err)
	}
}

func TestAssignSystemRoleSuccess(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleMember})
	svc := NewService(repo, "test-secret")

	user, err := svc.AssignSystemRole(context.Background(), "u-1", string(SystemRoleAdmin))
	if err != nil {
		t.Fatalf("assign role failed: %v", err)
	}
	if user.Role != SystemRoleAdmin {
		t.Fatalf("expected role admin, got %s", user.Role)
	}
}

func TestAssignSystemRoleInvalidRole(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleMember})
	svc := NewService(repo, "test-secret")

	_, err := svc.AssignSystemRole(context.Background(), "u-1", "invalid-role")
	if err == nil {
		t.Fatal("expected validation error")
	}

	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeValidation {
		t.Fatalf("expected validation biz error, got %v", err)
	}
}

func TestDisableUserClearsSession(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	repo.sessions["u-1"] = "refresh-token"
	svc := NewService(repo, "test-secret")

	user, err := svc.DisableUser(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("disable user failed: %v", err)
	}
	if user.Status != UserStatusDisabled {
		t.Fatalf("expected status disabled, got %s", user.Status)
	}
	if _, exists := repo.sessions["u-1"]; exists {
		t.Fatal("expected refresh token session to be deleted")
	}
}

func TestDisableUserRedisError(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	repo.sessions["u-1"] = "refresh-token"
	repo.deleteErr = errors.New("redis down")
	svc := NewService(repo, "test-secret")

	_, err := svc.DisableUser(context.Background(), "u-1")
	if err == nil {
		t.Fatal("expected internal server error")
	}

	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeInternalServerError {
		t.Fatalf("expected internal server biz error, got %v", err)
	}
}

func TestRefreshSuccess(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	svc := NewService(repo, "test-secret")

	pair, err := svc.Login(context.Background(), "alice", "pass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	newAccess, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newAccess == "" {
		t.Fatal("expected new access token")
	}
	if repo.sessions["u-1"] != pair.RefreshToken {
		t.Fatal("expected refresh token unchanged")
	}
}

func TestRefreshFailSessionMissing(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")
	tok, err := svc.signToken("u-1", "alice", string(SystemRoleMember), "refresh", RefreshTokenTTL)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}

	_, err = svc.Refresh(context.Background(), tok)
	if err == nil {
		t.Fatal("expected refresh error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized biz error, got %v", err)
	}
}

func TestRefreshFailExpiredToken(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")
	svc.now = func() time.Time { return time.Now().Add(-(RefreshTokenTTL + time.Hour)) }

	expiredToken, err := svc.signToken("u-1", "alice", string(SystemRoleMember), "refresh", RefreshTokenTTL)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}

	_, err = svc.Refresh(context.Background(), expiredToken)
	if err == nil {
		t.Fatal("expected refresh error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeTokenExpired {
		t.Fatalf("expected token expired biz error, got %v", err)
	}
}

func TestRefreshFailForgedToken(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")
	forgedSvc := NewService(repo, "other-secret")
	forgedToken, err := forgedSvc.signToken("u-1", "alice", string(SystemRoleMember), "refresh", RefreshTokenTTL)
	if err != nil {
		t.Fatalf("sign forged token failed: %v", err)
	}

	_, err = svc.Refresh(context.Background(), forgedToken)
	if err == nil {
		t.Fatal("expected refresh error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized biz error, got %v", err)
	}
}

func TestRefreshUsesCurrentRoleAfterDemotion(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleAdmin})
	svc := NewService(repo, "test-secret")

	pair, err := svc.Login(context.Background(), "alice", "pass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	repo.usersByID["u-1"].Role = SystemRoleMember
	repo.usersByName["alice"].Role = SystemRoleMember

	newAccess, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	claims, err := svc.parseToken(newAccess, "access")
	if err != nil {
		t.Fatalf("parse access token failed: %v", err)
	}
	if claims.Role != string(SystemRoleMember) {
		t.Fatalf("expected refreshed access token role member, got %s", claims.Role)
	}
}

func TestLogoutInvalidatesRefreshToken(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive})
	svc := NewService(repo, "test-secret")

	pair, err := svc.Login(context.Background(), "alice", "pass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if err == nil {
		t.Fatal("expected refresh to fail after logout")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeUnauthorized {
		t.Fatalf("expected unauthorized biz error, got %v", err)
	}
}
