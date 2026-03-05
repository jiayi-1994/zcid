package gitprovider

type ProviderType string

const (
	ProviderGitLab ProviderType = "gitlab"
	ProviderGitHub ProviderType = "github"
)

func ValidProvider(s string) bool {
	switch ProviderType(s) {
	case ProviderGitLab, ProviderGitHub:
		return true
	}
	return false
}

type Repository struct {
	ID       int64  `json:"id"`
	FullName string `json:"fullName"`
	Name     string `json:"name"`
	HTMLURL  string `json:"htmlUrl"`
	CloneURL string `json:"cloneUrl"`
	Private  bool   `json:"private"`
}

type Branch struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn,omitempty"`
}
