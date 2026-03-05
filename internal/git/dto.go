package git

const tokenMaskSuffix = 4

type CreateConnectionRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	ProviderType string `json:"providerType" binding:"required"`
	ServerURL    string `json:"serverUrl" binding:"required,url"`
	AccessToken  string `json:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	Description  string `json:"description"`
}

type UpdateConnectionRequest struct {
	Name         *string `json:"name"`
	AccessToken  *string `json:"accessToken"`
	RefreshToken *string `json:"refreshToken"`
	Description  *string `json:"description"`
}

type ConnectionResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"providerType"`
	ServerURL    string `json:"serverUrl"`
	TokenType    string `json:"tokenType"`
	TokenMask    string `json:"tokenMask"`
	Status       string `json:"status"`
	Description  string `json:"description"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type ConnectionListResponse struct {
	Items []ConnectionResponse `json:"items"`
	Total int64                `json:"total"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func maskToken(token string) string {
	if len(token) <= tokenMaskSuffix {
		return "****"
	}
	return "****" + token[len(token)-tokenMaskSuffix:]
}

func ToConnectionResponse(c *GitConnection) ConnectionResponse {
	mask := c.PlainTokenMask
	if mask == "" {
		mask = "****"
	}
	return ConnectionResponse{
		ID:           c.ID,
		Name:         c.Name,
		ProviderType: c.ProviderType,
		ServerURL:    c.ServerURL,
		TokenType:    string(c.TokenType),
		TokenMask:    mask,
		Status:       string(c.Status),
		Description:  c.Description,
		CreatedBy:    c.CreatedBy,
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
