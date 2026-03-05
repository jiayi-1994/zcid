package variable

import "time"

type VariableScope string

const (
	ScopeGlobal   VariableScope = "global"
	ScopeProject  VariableScope = "project"
	ScopePipeline VariableScope = "pipeline"
)

type VariableType string

const (
	TypePlain  VariableType = "plain"
	TypeSecret VariableType = "secret"
)

type VariableStatus string

const (
	StatusActive  VariableStatus = "active"
	StatusDeleted VariableStatus = "deleted"
)

type Variable struct {
	ID          string         `gorm:"column:id"`
	Scope       VariableScope  `gorm:"column:scope"`
	ProjectID   *string        `gorm:"column:project_id"`
	PipelineID  *string        `gorm:"column:pipeline_id"`
	Key         string         `gorm:"column:key"`
	Value       string         `gorm:"column:value"`
	VarType     VariableType   `gorm:"column:var_type"`
	Description string         `gorm:"column:description"`
	Status      VariableStatus `gorm:"column:status"`
	CreatedBy   string         `gorm:"column:created_by"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (Variable) TableName() string {
	return "variables"
}
