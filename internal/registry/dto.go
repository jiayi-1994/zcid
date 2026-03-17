package registry

type CreateRegistryRequest struct {
	Name      string `json:"name" binding:"required,max=200"`
	Type      string `json:"type" binding:"required"`
	URL       string `json:"url" binding:"required,max=500"`
	Username  string `json:"username" binding:"max=200"`
	Password  string `json:"password"`
	IsDefault bool   `json:"isDefault"`
}

type UpdateRegistryRequest struct {
	Name      *string `json:"name" binding:"omitempty,max=200"`
	Type      *string `json:"type" binding:"omitempty"`
	URL       *string `json:"url" binding:"omitempty,max=500"`
	Username  *string `json:"username" binding:"omitempty,max=200"`
	Password  *string `json:"password"`
	IsDefault *bool   `json:"isDefault"`
	Status    *string `json:"status" binding:"omitempty"`
}

type RegistryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	IsDefault bool   `json:"isDefault"`
	Status    string `json:"status"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type RegistryListResponse struct {
	Items []RegistryResponse `json:"items"`
	Total int64              `json:"total"`
}

type TestConnectionRequest struct {
	URL      string `json:"url" binding:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
