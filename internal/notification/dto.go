package notification

import "time"

type CreateRuleRequest struct {
	Name       string    `json:"name" binding:"required"`
	EventType  EventType `json:"eventType" binding:"required"`
	WebhookURL string    `json:"webhookUrl" binding:"required,url"`
	Enabled    *bool     `json:"enabled,omitempty"`
}

type UpdateRuleRequest struct {
	Name       *string    `json:"name,omitempty"`
	EventType  *EventType `json:"eventType,omitempty"`
	WebhookURL *string    `json:"webhookUrl,omitempty"`
	Enabled    *bool      `json:"enabled,omitempty"`
}

type RuleResponse struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"projectId"`
	Name       string    `json:"name"`
	EventType  string    `json:"eventType"`
	WebhookURL string    `json:"webhookUrl"`
	Enabled    bool      `json:"enabled"`
	CreatedBy  string    `json:"createdBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func ToRuleResponse(r *NotificationRule) RuleResponse {
	resp := RuleResponse{
		ID:         r.ID,
		ProjectID:  r.ProjectID,
		Name:       r.Name,
		EventType:  string(r.EventType),
		WebhookURL: r.WebhookURL,
		Enabled:    r.Enabled,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
	return resp
}
