package gitprovider

import "context"

// GitProvider abstracts Git hosting platform operations.
// Implementations: GitLab (REST API v4), GitHub (REST API v3).
type GitProvider interface {
	TestConnection(ctx context.Context) error
	ListRepos(ctx context.Context, page, pageSize int) (repos []Repository, total int, err error)
	ListBranches(ctx context.Context, repoFullName string) ([]Branch, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	GetProviderType() ProviderType
}

// New creates a GitProvider for the given type, server URL, and access token.
func New(providerType ProviderType, serverURL, accessToken string) (GitProvider, error) {
	switch providerType {
	case ProviderGitLab:
		return NewGitLabProvider(serverURL, accessToken)
	case ProviderGitHub:
		return NewGitHubProvider(serverURL, accessToken)
	default:
		return nil, ErrUnsupportedProvider
	}
}
