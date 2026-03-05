package audit

import (
	"context"
	"log/slog"

	"github.com/xjy/zcid/pkg/response"
)

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

func (s *Service) List(ctx context.Context, opts ListOpts) ([]*AuditLog, int64, error) {
	list, total, err := s.repo.List(ctx, opts)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeAuditQueryFailed, "failed to query audit logs", err.Error())
	}
	return list, total, nil
}
