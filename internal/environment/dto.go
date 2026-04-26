package environment

type CreateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Namespace   string `json:"namespace" binding:"required"`
	Description string `json:"description"`
}

type UpdateEnvironmentRequest struct {
	Name        *string `json:"name"`
	Namespace   *string `json:"namespace"`
	Description *string `json:"description"`
}

type EnvironmentResponse struct {
	ID          string                     `json:"id"`
	ProjectID   string                     `json:"projectId"`
	Name        string                     `json:"name"`
	Namespace   string                     `json:"namespace"`
	Description string                     `json:"description"`
	Status      string                     `json:"status"`
	Health      *EnvironmentHealthResponse `json:"health,omitempty"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
}

type EnvironmentHealthResponse struct {
	Status       string `json:"status"`
	Reason       string `json:"reason"`
	LastSignalAt string `json:"lastSignalAt,omitempty"`
	Stale        bool   `json:"stale"`
}

type EnvironmentListResponse struct {
	Items    []EnvironmentResponse `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

func ToEnvironmentResponse(e *Environment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Name:        e.Name,
		Namespace:   e.Namespace,
		Description: e.Description,
		Status:      string(e.Status),
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToEnvironmentResponseWithHealth(e *Environment, health EnvironmentHealthResponse) EnvironmentResponse {
	resp := ToEnvironmentResponse(e)
	resp.Health = &health
	return resp
}
