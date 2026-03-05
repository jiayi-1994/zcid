package project

import "time"

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"ownerId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ProjectListResponse struct {
	Items    []ProjectResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

func ToProjectResponse(p *Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type AddMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type MemberResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type MemberListResponse struct {
	Items []MemberResponse `json:"items"`
	Total int64            `json:"total"`
}

type MemberWithUsername struct {
	UserID    string      `gorm:"column:user_id"`
	Username  string      `gorm:"column:username"`
	Role      ProjectRole `gorm:"column:role"`
	CreatedAt time.Time   `gorm:"column:created_at"`
}
