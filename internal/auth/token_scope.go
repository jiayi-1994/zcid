package auth

import (
	"sort"
	"strings"

	"github.com/xjy/zcid/pkg/response"
)

const (
	ScopePipelinesRead     = "pipelines:read"
	ScopePipelinesTrigger  = "pipelines:trigger"
	ScopeDeploymentsRead   = "deployments:read"
	ScopeDeploymentsWrite  = "deployments:write"
	ScopeVariablesRead     = "variables:read"
	ScopeNotificationsRead = "notifications:read"
	ScopeAdminRead         = "admin:read"
)

var allowedTokenScopes = map[string]struct{}{
	ScopePipelinesRead:     {},
	ScopePipelinesTrigger:  {},
	ScopeDeploymentsRead:   {},
	ScopeDeploymentsWrite:  {},
	ScopeVariablesRead:     {},
	ScopeNotificationsRead: {},
	ScopeAdminRead:         {},
}

func NormalizeTokenScopes(scopes []string) ([]string, error) {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(scopes))
	for _, raw := range scopes {
		scope := strings.ToLower(strings.TrimSpace(raw))
		if scope == "" {
			continue
		}
		if _, ok := allowedTokenScopes[scope]; !ok {
			return nil, response.NewBizError(response.CodeValidation, "invalid token scope", scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	if len(normalized) == 0 {
		return nil, response.NewBizError(response.CodeValidation, "invalid request", "at least one scope is required")
	}
	sort.Strings(normalized)
	return normalized, nil
}

func EncodeTokenScopes(scopes []string) string {
	return strings.Join(scopes, ",")
}

func DecodeTokenScopes(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope != "" {
			out = append(out, scope)
		}
	}
	return out
}

func TokenHasScope(scopes []string, required string) bool {
	required = strings.TrimSpace(required)
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}
