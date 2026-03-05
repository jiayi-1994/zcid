package variable

const MaskedValue = "******"

type CreateVariableRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	VarType     string `json:"varType"`
	Description string `json:"description"`
}

type UpdateVariableRequest struct {
	Value       *string `json:"value"`
	Description *string `json:"description"`
}

type VariableResponse struct {
	ID          string `json:"id"`
	Scope       string `json:"scope"`
	ProjectID   string `json:"projectId,omitempty"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	VarType     string `json:"varType"`
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type VariableListResponse struct {
	Items []VariableResponse `json:"items"`
	Total int64              `json:"total"`
}

func ToVariableResponse(v *Variable, maskSecrets bool) VariableResponse {
	resp := VariableResponse{
		ID:          v.ID,
		Scope:       string(v.Scope),
		Key:         v.Key,
		Value:       v.Value,
		VarType:     string(v.VarType),
		Description: v.Description,
		CreatedBy:   v.CreatedBy,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if v.ProjectID != nil {
		resp.ProjectID = *v.ProjectID
	}
	if maskSecrets && v.VarType == TypeSecret {
		resp.Value = MaskedValue
	}
	return resp
}
