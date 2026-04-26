package svcdef

type CreateServiceRequest struct {
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description"`
	RepoURL        string   `json:"repoUrl"`
	ServiceType    string   `json:"serviceType"`
	Language       string   `json:"language"`
	Owner          string   `json:"owner"`
	Tags           []string `json:"tags"`
	PipelineIDs    []string `json:"pipelineIds"`
	EnvironmentIDs []string `json:"environmentIds"`
}

type UpdateServiceRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	RepoURL        *string  `json:"repoUrl"`
	ServiceType    *string  `json:"serviceType"`
	Language       *string  `json:"language"`
	Owner          *string  `json:"owner"`
	Tags           []string `json:"tags"`
	PipelineIDs    []string `json:"pipelineIds"`
	EnvironmentIDs []string `json:"environmentIds"`
}

type ServiceResponse struct {
	ID             string   `json:"id"`
	ProjectID      string   `json:"projectId"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RepoURL        string   `json:"repoUrl"`
	ServiceType    string   `json:"serviceType"`
	Language       string   `json:"language"`
	Owner          string   `json:"owner"`
	Tags           []string `json:"tags"`
	PipelineIDs    []string `json:"pipelineIds"`
	EnvironmentIDs []string `json:"environmentIds"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

type ServiceListResponse struct {
	Items    []ServiceResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

func ToServiceResponse(s *ServiceDef) ServiceResponse {
	return ServiceResponse{
		ID:             s.ID,
		ProjectID:      s.ProjectID,
		Name:           s.Name,
		Description:    s.Description,
		RepoURL:        s.RepoURL,
		ServiceType:    s.ServiceType,
		Language:       s.Language,
		Owner:          s.Owner,
		Tags:           []string(s.Tags),
		PipelineIDs:    []string(s.PipelineIDs),
		EnvironmentIDs: []string(s.EnvironmentIDs),
		Status:         string(s.Status),
		CreatedAt:      s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
