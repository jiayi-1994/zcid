//go:build integration

package integration

import (
	"testing"
)

func TestAuthLogin(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := NewAPIClient()

	t.Run("successful login", func(t *testing.T) {
		username, password := adminCredentials(t)
		resp := c.PostJSON(t, "/api/v1/auth/login", map[string]string{
			"username": username,
			"password": password,
		})
		requireCode(t, resp, 0)

		var data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.AccessToken == "" {
			t.Error("accessToken should not be empty")
		}
		if data.RefreshToken == "" {
			t.Error("refreshToken should not be empty")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		resp := c.PostJSON(t, "/api/v1/auth/login", map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		})
		if resp.Code == 0 {
			t.Error("expected non-zero code for invalid credentials")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		resp := c.PostJSON(t, "/api/v1/auth/login", map[string]string{
			"username": "",
			"password": "",
		})
		if resp.Code == 0 {
			t.Error("expected non-zero code for empty credentials")
		}
	})
}

func TestAuthTokenRefresh(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := NewAPIClient()
	username, password := adminCredentials(t)

	loginResp := c.PostJSON(t, "/api/v1/auth/login", map[string]string{
		"username": username,
		"password": password,
	})
	requireCode(t, loginResp, 0)

	var loginData struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	mustUnmarshalData(t, loginResp, &loginData)

	t.Run("refresh token", func(t *testing.T) {
		resp := c.PostJSON(t, "/api/v1/auth/refresh", map[string]string{
			"refreshToken": loginData.RefreshToken,
		})
		requireCode(t, resp, 0)

		var data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}
		mustUnmarshalData(t, resp, &data)
		if data.AccessToken == "" {
			t.Error("refreshed accessToken should not be empty")
		}
	})
}

func TestAuthUserManagement(t *testing.T) {
	waitForServer(t)
	resetRateLimit(t)
	c := adminClient(t)

	t.Run("list users", func(t *testing.T) {
		resp := c.GetJSON(t, "/api/v1/admin/users")
		requireCode(t, resp, 0)

		var data []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		mustUnmarshalData(t, resp, &data)
		if len(data) < 1 {
			t.Error("expected at least 1 user (admin)")
		}
		found := false
		for _, u := range data {
			if u.Username == "admin" {
				found = true
				break
			}
		}
		if !found {
			t.Error("admin user not found in user list")
		}
	})
}
