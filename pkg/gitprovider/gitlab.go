package gitprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type gitLabProvider struct {
	serverURL   string
	accessToken string
	httpClient  *http.Client
}

func NewGitLabProvider(serverURL, accessToken string) (*gitLabProvider, error) {
	serverURL = strings.TrimRight(serverURL, "/")
	return &gitLabProvider{
		serverURL:   serverURL,
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (g *gitLabProvider) GetProviderType() ProviderType {
	return ProviderGitLab
}

func (g *gitLabProvider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.serverURL+"/api/v4/user", nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("PRIVATE-TOKEN", g.accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrAuthFailed
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%w: status %d, body: %s", ErrAPICall, resp.StatusCode, string(body))
	}
	return nil
}

func (g *gitLabProvider) ListRepos(ctx context.Context, page, pageSize int) ([]Repository, int, error) {
	url := fmt.Sprintf("%s/api/v4/projects?membership=true&simple=true&per_page=%d&page=%d&order_by=updated_at",
		g.serverURL, pageSize, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("PRIVATE-TOKEN", g.accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, ErrAuthFailed
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, 0, fmt.Errorf("%w: status %d, body: %s", ErrAPICall, resp.StatusCode, string(body))
	}

	var glProjects []struct {
		ID                int64  `json:"id"`
		PathWithNamespace string `json:"path_with_namespace"`
		Name              string `json:"name"`
		WebURL            string `json:"web_url"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		Visibility        string `json:"visibility"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&glProjects); err != nil {
		return nil, 0, fmt.Errorf("%w: decode error: %v", ErrAPICall, err)
	}

	total := 0
	if v := resp.Header.Get("X-Total"); v != "" {
		total, _ = strconv.Atoi(v)
	}

	repos := make([]Repository, len(glProjects))
	for i, p := range glProjects {
		repos[i] = Repository{
			ID:       p.ID,
			FullName: p.PathWithNamespace,
			Name:     p.Name,
			HTMLURL:  p.WebURL,
			CloneURL: p.HTTPURLToRepo,
			Private:  p.Visibility != "public",
		}
	}
	return repos, total, nil
}

func (g *gitLabProvider) ListBranches(ctx context.Context, repoFullName string) ([]Branch, error) {
	encoded := strings.ReplaceAll(repoFullName, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches?per_page=100", g.serverURL, encoded)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("PRIVATE-TOKEN", g.accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAuthFailed
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrAPICall, resp.StatusCode, string(body))
	}

	var glBranches []struct {
		Name    string `json:"name"`
		Default bool   `json:"default"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&glBranches); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrAPICall, err)
	}

	branches := make([]Branch, len(glBranches))
	for i, b := range glBranches {
		branches[i] = Branch{Name: b.Name, IsDefault: b.Default}
	}
	return branches, nil
}

// RefreshToken refreshes the OAuth token via GitLab's token endpoint.
// MVP: PAT mode does not support refresh. Returns ErrAuthFailed.
func (g *gitLabProvider) RefreshToken(_ context.Context, _ string) (*TokenPair, error) {
	return nil, fmt.Errorf("%w: PAT tokens cannot be refreshed", ErrAuthFailed)
}
