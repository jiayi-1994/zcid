package argocd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RestClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewRestClient(serverURL, token string, insecure bool) *RestClient {
	transport := &http.Transport{}
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &RestClient{
		baseURL: normalizeURL(serverURL),
		token:   token,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

var _ ArgoClient = (*RestClient)(nil)

func (c *RestClient) CreateOrUpdateApp(ctx context.Context, app *ArgoApp) error {
	body := buildAppSpec(app)

	_, err := c.doGet(ctx, "/api/v1/applications/"+app.Name)
	if err == nil {
		return c.doRequest(ctx, http.MethodPut, "/api/v1/applications/"+app.Name, body, nil)
	}

	return c.doRequest(ctx, http.MethodPost, "/api/v1/applications", body, nil)
}

func (c *RestClient) SyncApp(ctx context.Context, appName string) error {
	syncReq := map[string]interface{}{
		"name": appName,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/applications/"+appName+"/sync", syncReq, nil)
}

func (c *RestClient) GetAppStatus(ctx context.Context, appName string) (*AppStatus, error) {
	data, err := c.doGet(ctx, "/api/v1/applications/"+appName)
	if err != nil {
		return nil, err
	}

	var appResp struct {
		Status struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			Resources []struct {
				Kind    string `json:"kind"`
				Name    string `json:"name"`
				Status  string `json:"status"`
				Message string `json:"message"`
			} `json:"resources"`
		} `json:"status"`
	}

	if err := json.Unmarshal(data, &appResp); err != nil {
		return nil, fmt.Errorf("decode app status: %w", err)
	}

	result := &AppStatus{
		Health: appResp.Status.Health.Status,
		Sync:   appResp.Status.Sync.Status,
	}

	for _, r := range appResp.Status.Resources {
		result.Resources = append(result.Resources, ResourceStatus{
			Kind:    r.Kind,
			Name:    r.Name,
			Status:  r.Status,
			Message: r.Message,
		})
	}

	return result, nil
}

func (c *RestClient) DeleteApp(ctx context.Context, appName string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/applications/"+appName, nil, nil)
}

func (c *RestClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ArgoCD request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ArgoCD %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

func (c *RestClient) doGet(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ArgoCD GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ArgoCD GET %s returned %d", path, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func normalizeURL(serverURL string) string {
	if len(serverURL) > 0 && serverURL[len(serverURL)-1] == '/' {
		return serverURL[:len(serverURL)-1]
	}
	if len(serverURL) > 8 {
		return serverURL
	}
	return "https://" + serverURL
}

func buildAppSpec(app *ArgoApp) map[string]interface{} {
	spec := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      app.Name,
			"namespace": "argocd",
			"labels": map[string]string{
				"zcid.io/managed-by": "zcid",
			},
		},
		"spec": map[string]interface{}{
			"project": func() string {
				if app.Project != "" {
					return app.Project
				}
				return "zcid-managed"
			}(),
			"source": map[string]interface{}{
				"repoURL":        app.RepoURL,
				"path":           app.Path,
				"targetRevision": app.TargetRevision,
			},
			"destination": map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": app.Namespace,
			},
		},
	}
	return spec
}
