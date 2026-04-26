//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var apiBaseURL string

func init() {
	apiBaseURL = os.Getenv("ZCID_API_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}
}

type APIClient struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
}

func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL:    apiBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type APIResponse struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	Detail    string          `json:"detail,omitempty"`
	RequestID string          `json:"requestId"`
}

func (c *APIClient) Login(t *testing.T, username, password string) {
	t.Helper()
	body := map[string]string{"username": username, "password": password}
	resp := c.PostJSON(t, "/api/v1/auth/login", body)
	requireCode(t, resp, 0)

	var data struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	mustUnmarshalData(t, resp, &data)
	if data.AccessToken == "" {
		t.Fatal("login returned empty access token")
	}
	c.AccessToken = data.AccessToken
}

func (c *APIClient) doRequest(t *testing.T, method, path string, body interface{}) *APIResponse {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		t.Fatalf("unmarshal response (status=%d, body=%s): %v", resp.StatusCode, string(respBody), err)
	}
	return &apiResp
}

func (c *APIClient) GetJSON(t *testing.T, path string) *APIResponse {
	t.Helper()
	return c.doRequest(t, http.MethodGet, path, nil)
}

func (c *APIClient) PostJSON(t *testing.T, path string, body interface{}) *APIResponse {
	t.Helper()
	return c.doRequest(t, http.MethodPost, path, body)
}

func (c *APIClient) PutJSON(t *testing.T, path string, body interface{}) *APIResponse {
	t.Helper()
	return c.doRequest(t, http.MethodPut, path, body)
}

func (c *APIClient) DeleteJSON(t *testing.T, path string) *APIResponse {
	t.Helper()
	return c.doRequest(t, http.MethodDelete, path, nil)
}

func requireCode(t *testing.T, resp *APIResponse, code int) {
	t.Helper()
	if resp.Code != code {
		t.Fatalf("expected response code %d, got %d (message=%s)", code, resp.Code, resp.Message)
	}
}

func mustUnmarshalData(t *testing.T, resp *APIResponse, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(resp.Data, v); err != nil {
		t.Fatalf("unmarshal data: %v (raw=%s)", err, string(resp.Data))
	}
}

func adminClient(t *testing.T) *APIClient {
	t.Helper()
	c := NewAPIClient()
	username, password := adminCredentials(t)
	c.Login(t, username, password)
	return c
}

func adminCredentials(t *testing.T) (string, string) {
	t.Helper()
	username := os.Getenv("ZCID_ADMIN_USERNAME")
	password := os.Getenv("ZCID_ADMIN_PASSWORD")
	if username == "" || password == "" {
		t.Skip("set ZCID_ADMIN_USERNAME and ZCID_ADMIN_PASSWORD for integration tests; default admin credentials are intentionally disabled")
	}
	return username, password
}

func waitForServer(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(apiBaseURL + "/healthz")
		if err == nil {
			statusCode := resp.StatusCode
			resp.Body.Close()
			if statusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatal("server did not become ready within 30s")
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%100000)
}

func resetRateLimit(t *testing.T) {
	t.Helper()

	redisAddr := os.Getenv("ZCID_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	keys, err := rdb.Keys(ctx, "ratelimit:*").Result()
	if err != nil {
		t.Logf("could not list rate limit keys: %v", err)
		return
	}
	if len(keys) > 0 {
		if err := rdb.Del(ctx, keys...).Err(); err != nil {
			t.Logf("could not delete rate limit keys: %v", err)
		}
	}
}
