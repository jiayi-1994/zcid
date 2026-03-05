package rbac

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	PolicyUpdateChannel = "rbac:policy:update"
)

const modelText = `
[request_definition]
r = sub, proj, obj, act

[policy_definition]
p = sub, proj, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (p.proj == "*" || r.proj == p.proj) && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`

func NewEnforcer(db *gorm.DB) (*casbin.SyncedEnforcer, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", "casbin_rule")
	if err != nil {
		return nil, fmt.Errorf("init gorm adapter: %w", err)
	}

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("init casbin model: %w", err)
	}

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("init casbin enforcer: %w", err)
	}
	enforcer.EnableAutoSave(true)

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load casbin policy: %w", err)
	}

	if err := ensureBasePolicies(enforcer); err != nil {
		return nil, err
	}

	return enforcer, nil
}

func StartWatcher(ctx context.Context, enforcer *casbin.SyncedEnforcer, rdb *redis.Client) {
	if enforcer == nil || rdb == nil {
		return
	}

	go func() {
		pubsub := rdb.Subscribe(ctx, PolicyUpdateChannel)
		defer func() {
			if err := pubsub.Close(); err != nil {
				slog.Warn("rbac watcher close failed", slog.Any("error", err))
			}
		}()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if strings.TrimSpace(msg.Payload) == "" {
					continue
				}
				if err := enforcer.LoadPolicy(); err != nil {
					slog.Error("rbac policy reload failed", slog.Any("error", err))
				}
			}
		}
	}()
}

func ensureBasePolicies(enforcer *casbin.SyncedEnforcer) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer is nil")
	}

	if _, err := enforcer.AddGroupingPolicy("admin", "project_admin"); err != nil {
		return fmt.Errorf("add admin inheritance: %w", err)
	}
	if _, err := enforcer.AddGroupingPolicy("project_admin", "member"); err != nil {
		return fmt.Errorf("add project_admin inheritance: %w", err)
	}

	if _, err := enforcer.AddPolicy("admin", "*", "/api/v1/admin/*", "GET|POST|PUT|PATCH|DELETE"); err != nil {
		return fmt.Errorf("add admin route policy: %w", err)
	}

	return nil
}
