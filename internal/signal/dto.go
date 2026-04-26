package signal

import "time"

type RecordInput struct {
	ProjectID     string
	TargetType    TargetType
	TargetID      string
	Source        string
	Status        Status
	Severity      Severity
	Reason        string
	Message       string
	ObservedValue any
	ObservedAt    time.Time
	StaleAfter    *time.Time
}

type HealthSignalResponse struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"projectId"`
	TargetType      string     `json:"targetType"`
	TargetID        string     `json:"targetId"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	EffectiveStatus string     `json:"effectiveStatus"`
	Severity        string     `json:"severity"`
	Reason          string     `json:"reason"`
	Message         string     `json:"message"`
	ObservedValue   JSONRaw    `json:"observedValue"`
	ObservedAt      time.Time  `json:"observedAt"`
	StaleAfter      *time.Time `json:"staleAfter,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func ToResponse(row HealthSignal, now time.Time) HealthSignalResponse {
	return HealthSignalResponse{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		TargetType:      string(row.TargetType),
		TargetID:        row.TargetID,
		Source:          row.Source,
		Status:          string(row.Status),
		EffectiveStatus: string(row.EffectiveStatus(now)),
		Severity:        string(row.Severity),
		Reason:          row.Reason,
		Message:         row.Message,
		ObservedValue:   row.ObservedValue,
		ObservedAt:      row.ObservedAt,
		StaleAfter:      row.StaleAfter,
		CreatedAt:       row.CreatedAt,
	}
}
