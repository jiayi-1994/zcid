package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xjy/zcid/pkg/response"
)

func TestTokenServiceCreatePATStoresHashAndReturnsRawOnce(t *testing.T) {
	repo := newMockRepo()
	svc := NewTokenService(repo, nil)
	svc.now = func() time.Time { return time.Unix(100, 0) }

	created, err := svc.Create(context.Background(), CreateAccessTokenInput{
		Name:      "ci bot",
		Type:      AccessTokenTypePersonal,
		Scopes:    []string{ScopePipelinesTrigger},
		ExpiresAt: time.Unix(100, 0).Add(time.Hour),
		UserID:    "u-1",
		ActorID:   "u-1",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	if created.Raw == "" || created.Token.TokenHash == created.Raw {
		t.Fatal("expected raw token once and hashed storage")
	}
	if created.Token.TokenPrefix != PersonalTokenPrefix {
		t.Fatalf("expected PAT prefix, got %s", created.Token.TokenPrefix)
	}
	if _, ok := repo.accessTokens[created.Token.ID]; !ok {
		t.Fatal("expected token stored")
	}
}

func TestTokenServiceRejectsTooLongExpiryAndInvalidScope(t *testing.T) {
	repo := newMockRepo()
	svc := NewTokenService(repo, nil)
	now := time.Unix(100, 0)
	svc.now = func() time.Time { return now }

	_, err := svc.Create(context.Background(), CreateAccessTokenInput{Name: "bad", Type: AccessTokenTypePersonal, Scopes: []string{ScopePipelinesTrigger}, ExpiresAt: now.Add(MaxAccessTokenTTL + time.Hour), UserID: "u-1", ActorID: "u-1"})
	if err == nil {
		t.Fatal("expected max expiry validation error")
	}
	var bizErr *response.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != response.CodeValidation {
		t.Fatalf("expected validation error, got %v", err)
	}

	_, err = svc.Create(context.Background(), CreateAccessTokenInput{Name: "bad", Type: AccessTokenTypePersonal, Scopes: []string{"admin:*"}, ExpiresAt: now.Add(time.Hour), UserID: "u-1", ActorID: "u-1"})
	if err == nil {
		t.Fatal("expected invalid scope error")
	}
}

func TestNormalizeTokenScopesUsesCanonicalVocabulary(t *testing.T) {
	scopes, err := NormalizeTokenScopes([]string{
		" pipelines:read ",
		ScopePipelinesTrigger,
		ScopeDeploymentsRead,
		ScopeDeploymentsWrite,
		ScopeVariablesRead,
		ScopeNotificationsRead,
		ScopeAdminRead,
		ScopePipelinesRead,
	})
	if err != nil {
		t.Fatalf("normalize scopes failed: %v", err)
	}
	expected := []string{ScopeAdminRead, ScopeDeploymentsRead, ScopeDeploymentsWrite, ScopeNotificationsRead, ScopePipelinesRead, ScopePipelinesTrigger, ScopeVariablesRead}
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d: %#v", len(expected), len(scopes), scopes)
	}
	for i := range expected {
		if scopes[i] != expected[i] {
			t.Fatalf("expected scope[%d]=%s, got %s", i, expected[i], scopes[i])
		}
	}
}

func TestTokenServiceValidateRejectsRevokedExpiredAndMissingScope(t *testing.T) {
	repo := newMockRepo()
	svc := NewTokenService(repo, nil)
	now := time.Unix(100, 0)
	svc.now = func() time.Time { return now }
	created, err := svc.Create(context.Background(), CreateAccessTokenInput{Name: "ci", Type: AccessTokenTypeProject, Scopes: []string{ScopePipelinesRead}, ExpiresAt: now.Add(time.Hour), ProjectID: "p-1", ActorID: "u-1"})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}

	if _, err := svc.Validate(context.Background(), created.Raw, ScopePipelinesTrigger, ""); err == nil {
		t.Fatal("expected missing scope rejection")
	}

	if err := svc.Revoke(context.Background(), created.Token.ID, "u-1", string(SystemRoleAdmin), ""); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if _, err := svc.Validate(context.Background(), created.Raw, ScopePipelinesRead, ""); err == nil {
		t.Fatal("expected revoked token rejection")
	}
}

func TestTokenServiceRevokeEnforcesActorOwnership(t *testing.T) {
	repo := newMockRepo()
	svc := NewTokenService(repo, nil)
	now := time.Unix(100, 0)
	svc.now = func() time.Time { return now }

	owned, err := svc.Create(context.Background(), CreateAccessTokenInput{Name: "owned", Type: AccessTokenTypePersonal, Scopes: []string{ScopePipelinesRead}, ExpiresAt: now.Add(time.Hour), UserID: "u-1", ActorID: "u-1"})
	if err != nil {
		t.Fatalf("create owned token failed: %v", err)
	}
	other, err := svc.Create(context.Background(), CreateAccessTokenInput{Name: "other", Type: AccessTokenTypePersonal, Scopes: []string{ScopePipelinesRead}, ExpiresAt: now.Add(time.Hour), UserID: "u-2", ActorID: "u-2"})
	if err != nil {
		t.Fatalf("create other token failed: %v", err)
	}
	project, err := svc.Create(context.Background(), CreateAccessTokenInput{Name: "project", Type: AccessTokenTypeProject, Scopes: []string{ScopePipelinesRead}, ExpiresAt: now.Add(time.Hour), ProjectID: "p-1", ActorID: "admin-1"})
	if err != nil {
		t.Fatalf("create project token failed: %v", err)
	}

	if err := svc.Revoke(context.Background(), owned.Token.ID, "u-1", string(SystemRoleMember), ""); err != nil {
		t.Fatalf("owner should revoke own PAT: %v", err)
	}
	if err := svc.Revoke(context.Background(), other.Token.ID, "u-1", string(SystemRoleMember), ""); err == nil {
		t.Fatal("expected non-admin to be forbidden from revoking another user's PAT")
	}
	if err := svc.Revoke(context.Background(), project.Token.ID, "u-1", string(SystemRoleMember), ""); err == nil {
		t.Fatal("expected non-admin to be forbidden from revoking project tokens")
	}
	if err := svc.Revoke(context.Background(), project.Token.ID, "admin-1", string(SystemRoleAdmin), ""); err != nil {
		t.Fatalf("admin should revoke project token: %v", err)
	}
}
