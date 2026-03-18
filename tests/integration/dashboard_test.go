//go:build integration

package integration

import (
	"testing"
)

func TestAdminHealth(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("get system health", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/admin/health")
		requireCode(t, resp, 0)
		t.Logf("health check passed")
	})
}

func TestAdminSettings(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("get settings", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/admin/settings")
		requireCode(t, resp, 0)
		t.Logf("admin settings retrieved successfully")
	})
}

func TestAuditLog(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("list audit logs", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/admin/audit-logs?page=1&pageSize=10")
		requireCode(t, resp, 0)

		var data struct {
			Items []struct {
				ID     string `json:"id"`
				Action string `json:"action"`
			} `json:"items"`
			Total int64 `json:"total"`
		}
		mustUnmarshalData(t, resp, &data)
		t.Logf("audit logs: total=%d", data.Total)
	})
}
