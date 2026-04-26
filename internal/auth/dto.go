package auth

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Status   string `json:"status"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
	Status   *string `json:"status"`
	Role     *string `json:"role"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type AssignRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type BootstrapRedeemRequest struct {
	Token    string `json:"token" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type BootstrapStatusResponse struct {
	Required bool `json:"required"`
}

type CreateAccessTokenRequest struct {
	Name      string   `json:"name" binding:"required"`
	Type      string   `json:"type" binding:"required"`
	Scopes    []string `json:"scopes" binding:"required"`
	ExpiresAt string   `json:"expiresAt" binding:"required"`
	ProjectID string   `json:"projectId"`
}

type AccessTokenResponse struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	TokenPrefix string   `json:"tokenPrefix"`
	Scopes      []string `json:"scopes"`
	UserID      *string  `json:"userId,omitempty"`
	ProjectID   *string  `json:"projectId,omitempty"`
	CreatedBy   string   `json:"createdBy"`
	ExpiresAt   string   `json:"expiresAt"`
	LastUsedAt  *string  `json:"lastUsedAt,omitempty"`
	RevokedAt   *string  `json:"revokedAt,omitempty"`
	CreatedAt   string   `json:"createdAt"`
}

type CreateAccessTokenResponse struct {
	Token AccessTokenResponse `json:"token"`
	Raw   string              `json:"rawToken"`
}
