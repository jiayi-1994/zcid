package gitprovider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// WebhookEvent represents a parsed webhook event from GitLab or GitHub.
type WebhookEvent struct {
	Provider    ProviderType `json:"provider"`
	EventType   string       `json:"eventType"`
	RepoURL     string       `json:"repoUrl"`
	RepoName    string       `json:"repoName"`
	Branch      string       `json:"branch"`
	CommitSHA   string       `json:"commitSha"`
	CommitMsg   string       `json:"commitMsg"`
	AuthorName  string       `json:"authorName"`
	AuthorEmail string       `json:"authorEmail"`
}

// VerifyGitLabSignature checks X-Gitlab-Token header against the stored secret.
func VerifyGitLabSignature(headerToken, secret string) error {
	if headerToken == "" {
		return fmt.Errorf("%w: missing X-Gitlab-Token header", ErrAuthFailed)
	}
	if headerToken != secret {
		return fmt.Errorf("%w: X-Gitlab-Token mismatch", ErrAuthFailed)
	}
	return nil
}

// VerifyGitHubSignature checks X-Hub-Signature-256 header using HMAC-SHA256.
func VerifyGitHubSignature(signatureHeader string, body []byte, secret string) error {
	if signatureHeader == "" {
		return fmt.Errorf("%w: missing X-Hub-Signature-256 header", ErrAuthFailed)
	}

	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return fmt.Errorf("%w: invalid signature format", ErrAuthFailed)
	}

	expectedMAC := signatureHeader[len(prefix):]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(actualMAC), []byte(expectedMAC)) {
		return fmt.Errorf("%w: HMAC signature mismatch", ErrAuthFailed)
	}
	return nil
}

// IdempotencyKey generates the deduplication key for a webhook event.
func IdempotencyKey(event *WebhookEvent, timestampMinute int64) string {
	return fmt.Sprintf("webhook:%s:%s:%s:%d", event.EventType, event.RepoURL, event.CommitSHA, timestampMinute)
}

// IdempotencyKeyWithDelivery uses delivery ID when present for precise dedup (FR19).
// GitLab: X-Gitlab-Event-UUID, GitHub: X-GitHub-Delivery.
// Falls back to IdempotencyKey(event, timestampMinute) if deliveryID is empty.
func IdempotencyKeyWithDelivery(event *WebhookEvent, deliveryID string) string {
	if deliveryID != "" {
		return fmt.Sprintf("webhook:delivery:%s", deliveryID)
	}
	return IdempotencyKey(event, time.Now().Unix()/60)
}
