package gitprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type gitHubProvider struct {
	serverURL   string
	accessToken string
	httpClient  *http.Client
}

func NewGitHubProvider(serverURL, accessToken string) (*gitHubProvider, error) {
	if serverURL == "" {
		serverURL = "https://api.github.com"
	}
	serverURL = strings.TrimRight(serverURL, "/")
	return &gitHubProvider{
		serverURL:   serverURL,
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (g *gitHubProvider) GetProviderType() ProviderType {
	return ProviderGitHub
}

func (g *gitHubProvider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.serverURL+"/user", nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("Authorization", "Bearer "+g.accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

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

func (g *gitHubProvider) ListRepos(ctx context.Context, page, pageSize int) ([]Repository, int, error) {
	url := fmt.Sprintf("%s/user/repos?per_page=%d&page=%d&sort=updated&affiliation=owner,collaborator,organization_member",
		g.serverURL, pageSize, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("Authorization", "Bearer "+g.accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

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

	var ghRepos []struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		HTMLURL  string `json:"html_url"`
		CloneURL string `json:"clone_url"`
		Private  bool   `json:"private"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghRepos); err != nil {
		return nil, 0, fmt.Errorf("%w: decode error: %v", ErrAPICall, err)
	}

	repos := make([]Repository, len(ghRepos))
	for i, r := range ghRepos {
		repos[i] = Repository{
			ID:       r.ID,
			FullName: r.FullName,
			Name:     r.Name,
			HTMLURL:  r.HTMLURL,
			CloneURL: r.CloneURL,
			Private:  r.Private,
		}
	}

	total := (page-1)*pageSize + len(repos)
	if len(repos) == pageSize {
		total = page*pageSize + 1
	}
	return repos, total, nil
}

func (g *gitHubProvider) ListBranches(ctx context.Context, repoFullName string) ([]Branch, error) {
	url := fmt.Sprintf("%s/repos/%s/branches?per_page=100", g.serverURL, repoFullName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPICall, err)
	}
	req.Header.Set("Authorization", "Bearer "+g.accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

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

	var ghBranches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghBranches); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrAPICall, err)
	}

	// GitHub needs a separate call for default branch; use repo info already available to caller
	branches := make([]Branch, len(ghBranches))
	for i, b := range ghBranches {
		branches[i] = Branch{Name: b.Name}
	}
	return branches, nil
}

// RefreshToken refreshes the OAuth token via GitHub's token endpoint.
// MVP: PAT mode does not support refresh. Returns ErrAuthFailed.
func (g *gitHubProvider) RefreshToken(_ context.Context, _ string) (*TokenPair, error) {
	return nil, fmt.Errorf("%w: PAT tokens cannot be refreshed", ErrAuthFailed)
}
