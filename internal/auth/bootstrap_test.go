package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

func TestEnsureBootstrapTokenCreatesTokenWhenOnlyLegacyAdminExists(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("admin123")
	repo.seedUser(&User{ID: legacyBootstrapAdminID, Username: "admin", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleAdmin})
	svc := NewService(repo, "test-secret")
	svc.now = func() time.Time { return time.Unix(100, 0) }

	token, generated, err := svc.EnsureBootstrapToken(context.Background())
	if err != nil {
		t.Fatalf("ensure bootstrap token failed: %v", err)
	}
	if !generated {
		t.Fatal("expected bootstrap token to be generated")
	}
	if token == "" {
		t.Fatal("expected plaintext token")
	}
	if len(repo.bootstrap) != 1 {
		t.Fatalf("expected one stored token verifier, got %d", len(repo.bootstrap))
	}
	if _, exists := repo.bootstrap[token]; exists {
		t.Fatal("plaintext token must not be used as storage key")
	}
	if repo.usersByID[legacyBootstrapAdminID].Status != UserStatusDisabled {
		t.Fatal("expected legacy admin to be disabled before bootstrap")
	}
}

func TestEnsureBootstrapTokenSkippedWhenConfiguredUserExists(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleAdmin})
	svc := NewService(repo, "test-secret")

	token, generated, err := svc.EnsureBootstrapToken(context.Background())
	if err != nil {
		t.Fatalf("ensure bootstrap token failed: %v", err)
	}
	if generated || token != "" {
		t.Fatalf("expected no token, generated=%v token=%q", generated, token)
	}
}

func TestRedeemBootstrapTokenActivatesLegacyAdmin(t *testing.T) {
	repo := newMockRepo()
	oldHash, _ := HashPassword("admin123")
	repo.seedUser(&User{ID: legacyBootstrapAdminID, Username: "admin", PasswordHash: oldHash, Status: UserStatusDisabled, Role: SystemRoleAdmin})
	now := time.Unix(100, 0)
	token := "zcid_bootstrap_test-token"
	repo.bootstrap[hashBootstrapToken(token)] = &BootstrapToken{ID: "bt-1", TokenHash: hashBootstrapToken(token), ExpiresAt: now.Add(time.Minute)}
	svc := NewService(repo, "test-secret")
	svc.now = func() time.Time { return now }

	user, err := svc.RedeemBootstrapToken(context.Background(), token, "root", "new-pass")
	if err != nil {
		t.Fatalf("redeem bootstrap token failed: %v", err)
	}
	if user.ID != legacyBootstrapAdminID {
		t.Fatalf("expected legacy admin id, got %s", user.ID)
	}
	if user.Username != "root" {
		t.Fatalf("expected username root, got %s", user.Username)
	}
	if user.Status != UserStatusActive || user.Role != SystemRoleAdmin {
		t.Fatalf("expected active admin, got status=%s role=%s", user.Status, user.Role)
	}
	if ComparePasswordHash(oldHash, "new-pass") {
		t.Fatal("expected password hash to change")
	}
	if repo.lastPolicyID != legacyBootstrapAdminID || repo.lastPolicy != SystemRoleAdmin {
		t.Fatalf("expected admin policy upsert, got id=%s role=%s", repo.lastPolicyID, repo.lastPolicy)
	}
	stored := repo.bootstrap[hashBootstrapToken(token)]
	if stored.UsedAt == nil {
		t.Fatal("expected bootstrap token to be marked used")
	}
}

func TestRedeemBootstrapTokenRejectsExpiredToken(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("admin123")
	repo.seedUser(&User{ID: legacyBootstrapAdminID, Username: "admin", PasswordHash: hash, Status: UserStatusDisabled, Role: SystemRoleAdmin})
	now := time.Unix(100, 0)
	token := "zcid_bootstrap_test-token"
	repo.bootstrap[hashBootstrapToken(token)] = &BootstrapToken{ID: "bt-1", TokenHash: hashBootstrapToken(token), ExpiresAt: now.Add(-time.Minute)}
	svc := NewService(repo, "test-secret")
	svc.now = func() time.Time { return now }

	_, err := svc.RedeemBootstrapToken(context.Background(), token, "root", "new-pass")
	if err == nil {
		t.Fatal("expected expired token error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeTokenExpired {
		t.Fatalf("expected token expired biz error, got %v", err)
	}
	if repo.usersByID[legacyBootstrapAdminID].Status != UserStatusDisabled {
		t.Fatal("expected legacy admin to remain disabled")
	}
}

func TestRedeemBootstrapTokenRejectsAfterConfiguredUserExists(t *testing.T) {
	repo := newMockRepo()
	hash, _ := HashPassword("pass123")
	repo.seedUser(&User{ID: "u-1", Username: "alice", PasswordHash: hash, Status: UserStatusActive, Role: SystemRoleAdmin})
	now := time.Unix(100, 0)
	token := "zcid_bootstrap_test-token"
	repo.bootstrap[hashBootstrapToken(token)] = &BootstrapToken{ID: "bt-1", TokenHash: hashBootstrapToken(token), ExpiresAt: now.Add(time.Minute)}
	svc := NewService(repo, "test-secret")
	svc.now = func() time.Time { return now }

	_, err := svc.RedeemBootstrapToken(context.Background(), token, "root", "new-pass")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeConflict {
		t.Fatalf("expected conflict biz error, got %v", err)
	}
}
