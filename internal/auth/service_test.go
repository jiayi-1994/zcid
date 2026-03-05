package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

type mockRepo struct {
	usersByName  map[string]*User
	usersByID    map[string]*User
	sessions     map[string]string
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
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		usersByName: map[string]*User{},
		usersByID:   map[string]*User{},
		sessions:    map[string]string{},
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
	svc := NewService(repo, "test-secret")

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
