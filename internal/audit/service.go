package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/xjy/zcid/pkg/response"
)

const (
	ResourceTypeAuthSecurity = "auth_security"
	ResultSuccess            = "success"
	ResultFailure            = "failure"
)

type AuthSecurityEvent struct {
	UserID     string
	Action     string
	ResourceID string
	Result     string
	IP         string
	Detail     AuthSecurityDetail
}

type AuthSecurityDetail struct {
	EventType     string         `json:"eventType"`
	PrincipalType string         `json:"principalType,omitempty"`
	TokenType     string         `json:"tokenType,omitempty"`
	TokenID       string         `json:"tokenId,omitempty"`
	TokenName     string         `json:"tokenName,omitempty"`
	TargetUserID  string         `json:"targetUserId,omitempty"`
	TargetProject string         `json:"targetProjectId,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	Fields        map[string]any `json:"fields,omitempty"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogAction(ctx context.Context, userID, action, resourceType, resourceID, result, ip, detail string) {
	log := &AuditLog{
		Action:       action,
		ResourceType: resourceType,
		Result:       result,
	}
	if userID != "" {
		log.UserID = &userID
	}
	if resourceID != "" {
		log.ResourceID = &resourceID
	}
	if ip != "" {
		log.IP = &ip
	}
	if detail != "" {
		log.Detail = &detail
	}
	if log.Result == "" {
		log.Result = "success"
	}
	go func() {
		if err := s.repo.Create(context.Background(), log); err != nil {
			slog.Error("audit log write failed", slog.Any("error", err), slog.String("action", action))
		}
	}()
}

func (s *Service) LogAuthSecurityEvent(ctx context.Context, event AuthSecurityEvent) {
	action := strings.TrimSpace(event.Action)
	if action == "" {
		action = strings.TrimSpace(event.Detail.EventType)
	}
	if action == "" {
		action = "auth.unknown"
	}

	result := strings.TrimSpace(event.Result)
	if result == "" {
		result = ResultSuccess
	}

	detail := event.Detail
	if strings.TrimSpace(detail.EventType) == "" {
		detail.EventType = action
	}
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		slog.Error("audit auth security detail marshal failed", slog.Any("error", err), slog.String("action", action))
		detailJSON, _ = json.Marshal(AuthSecurityDetail{EventType: action, Reason: "detail_marshal_failed"})
	}

	s.LogAction(ctx, event.UserID, action, ResourceTypeAuthSecurity, event.ResourceID, result, event.IP, string(detailJSON))
}

func (s *Service) List(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
	list, total, err := s.repo.List(ctx, opts)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeAuditQueryFailed, "failed to query audit logs", err.Error())
	}
	return list, total, nil
}
