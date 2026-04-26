package pipeline

type CreatePipelineRequest struct {
	Name              string            `json:"name" binding:"required,max=200"`
	Description       string            `json:"description"`
	Config            PipelineConfig    `json:"config"`
	TriggerType       string            `json:"triggerType"`
	ConcurrencyPolicy string            `json:"concurrencyPolicy"`
	TemplateID        string            `json:"templateId,omitempty"`
	TemplateParams    map[string]string `json:"templateParams,omitempty"`
}

type FromTemplateRequest struct {
	TemplateID string            `json:"templateId" binding:"required"`
	Name       string            `json:"name" binding:"required,max=200"`
	Params     map[string]string `json:"params"`
}

type UpdatePipelineRequest struct {
	Name              *string         `json:"name"`
	Description       *string         `json:"description"`
	Status            *string         `json:"status"`
	Config            *PipelineConfig `json:"config"`
	TriggerType       *string         `json:"triggerType"`
	ConcurrencyPolicy *string         `json:"concurrencyPolicy"`
}

type PipelineResponse struct {
	ID                string         `json:"id"`
	ProjectID         string         `json:"projectId"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Status            string         `json:"status"`
	Config            PipelineConfig `json:"config"`
	TriggerType       string         `json:"triggerType"`
	ConcurrencyPolicy string         `json:"concurrencyPolicy"`
	CreatedBy         string         `json:"createdBy"`
	CreatedAt         string         `json:"createdAt"`
	UpdatedAt         string         `json:"updatedAt"`
}

type PipelineSummaryResponse struct {
	ID                string `json:"id"`
	ProjectID         string `json:"projectId"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	TriggerType       string `json:"triggerType"`
	ConcurrencyPolicy string `json:"concurrencyPolicy"`
	CreatedBy         string `json:"createdBy"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type PipelineListResponse struct {
	Items    []PipelineSummaryResponse `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

func ToPipelineResponse(p *Pipeline) PipelineResponse {
	return PipelineResponse{
		ID:                p.ID,
		ProjectID:         p.ProjectID,
		Name:              p.Name,
		Description:       p.Description,
		Status:            string(p.Status),
		Config:            p.Config,
		TriggerType:       string(p.TriggerType),
		ConcurrencyPolicy: string(p.ConcurrencyPolicy),
		CreatedBy:         p.CreatedBy,
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToPipelineSummaryResponse(p *Pipeline) PipelineSummaryResponse {
	return PipelineSummaryResponse{
		ID:                p.ID,
		ProjectID:         p.ProjectID,
		Name:              p.Name,
		Description:       p.Description,
		Status:            string(p.Status),
		TriggerType:       string(p.TriggerType),
		ConcurrencyPolicy: string(p.ConcurrencyPolicy),
		CreatedBy:         p.CreatedBy,
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
