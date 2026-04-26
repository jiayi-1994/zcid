package svcdef

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type ServiceStatus string

const (
	StatusActive  ServiceStatus = "active"
	StatusDeleted ServiceStatus = "deleted"
)

type ServiceDef struct {
	ID             string        `gorm:"column:id"`
	ProjectID      string        `gorm:"column:project_id"`
	Name           string        `gorm:"column:name"`
	Description    string        `gorm:"column:description"`
	RepoURL        string        `gorm:"column:repo_url"`
	ServiceType    string        `gorm:"column:service_type"`
	Language       string        `gorm:"column:language"`
	Owner          string        `gorm:"column:owner"`
	Tags           StringList    `gorm:"column:tags;type:jsonb"`
	PipelineIDs    StringList    `gorm:"column:pipeline_ids;type:jsonb"`
	EnvironmentIDs StringList    `gorm:"column:environment_ids;type:jsonb"`
	Status         ServiceStatus `gorm:"column:status"`
	CreatedAt      time.Time     `gorm:"column:created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at"`
}

func (ServiceDef) TableName() string {
	return "services"
}

type StringList []string

func NewStringList(values []string) StringList {
	if values == nil {
		return StringList{}
	}
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return StringList(out)
}

func (l StringList) Value() (driver.Value, error) {
	if l == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]string(l))
	if err != nil {
		return nil, fmt.Errorf("StringList.Value: %w", err)
	}
	return b, nil
}

func (l *StringList) Scan(value interface{}) error {
	if value == nil {
		*l = StringList{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("StringList.Scan: unsupported type %T", value)
	}
	if len(raw) == 0 {
		*l = StringList{}
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("StringList.Scan: %w", err)
	}
	*l = NewStringList(values)
	return nil
}
