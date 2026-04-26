package notification

import "time"

type CreateRuleRequest struct {
	Name         string      `json:"name" binding:"required"`
	EventType    EventType   `json:"eventType" binding:"required"`
	ChannelType  ChannelType `json:"channelType"`
	WebhookURL   string      `json:"webhookUrl"`
	SlackToken   string      `json:"slackToken"`
	SlackChannel string      `json:"slackChannel"`
	Enabled      *bool       `json:"enabled,omitempty"`
}

type UpdateRuleRequest struct {
	Name         *string      `json:"name,omitempty"`
	EventType    *EventType   `json:"eventType,omitempty"`
	ChannelType  *ChannelType `json:"channelType,omitempty"`
	WebhookURL   *string      `json:"webhookUrl,omitempty"`
	SlackToken   *string      `json:"slackToken,omitempty"`
	SlackChannel *string      `json:"slackChannel,omitempty"`
	Enabled      *bool        `json:"enabled,omitempty"`
}

type RuleResponse struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectId"`
	Name          string    `json:"name"`
	EventType     string    `json:"eventType"`
	ChannelType   string    `json:"channelType"`
	WebhookURL    string    `json:"webhookUrl"`
	SlackChannel  string    `json:"slackChannel"`
	HasSlackToken bool      `json:"hasSlackToken"`
	Enabled       bool      `json:"enabled"`
	CreatedBy     string    `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func ToRuleResponse(r *NotificationRule) RuleResponse {
	resp := RuleResponse{
		ID:            r.ID,
		ProjectID:     r.ProjectID,
		Name:          r.Name,
		EventType:     string(r.EventType),
		ChannelType:   string(normalizeChannelType(r.ChannelType)),
		WebhookURL:    r.WebhookURL,
		SlackChannel:  r.SlackChannel,
		HasSlackToken: r.SlackToken != "",
		Enabled:       r.Enabled,
		CreatedBy:     r.CreatedBy,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	return resp
}
