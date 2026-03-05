package gitprovider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyGitLabSignature_Valid(t *testing.T) {
	err := VerifyGitLabSignature("my-secret-token", "my-secret-token")
	assert.NoError(t, err)
}

func TestVerifyGitLabSignature_Invalid(t *testing.T) {
	err := VerifyGitLabSignature("wrong-token", "my-secret-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthFailed)
}

func TestVerifyGitLabSignature_Empty(t *testing.T) {
	err := VerifyGitLabSignature("", "my-secret-token")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthFailed)
}

func TestVerifyGitHubSignature_Valid(t *testing.T) {
	secret := "my-github-secret"
	body := []byte(`{"ref":"refs/heads/main"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	err := VerifyGitHubSignature(sig, body, secret)
	assert.NoError(t, err)
}

func TestVerifyGitHubSignature_Invalid(t *testing.T) {
	err := VerifyGitHubSignature("sha256=deadbeef", []byte(`test`), "secret")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthFailed)
}

func TestVerifyGitHubSignature_MissingHeader(t *testing.T) {
	err := VerifyGitHubSignature("", []byte(`test`), "secret")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAuthFailed)
}

func TestVerifyGitHubSignature_BadFormat(t *testing.T) {
	err := VerifyGitHubSignature("invalid-format", []byte(`test`), "secret")
	assert.Error(t, err)
}

func TestIdempotencyKey(t *testing.T) {
	event := &WebhookEvent{
		EventType: "push",
		RepoURL:   "https://gitlab.example.com/group/project",
		CommitSHA: "abc123",
	}
	key := IdempotencyKey(event, 12345)
	require.Equal(t, "webhook:push:https://gitlab.example.com/group/project:abc123:12345", key)
}
