package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/response"
	"gorm.io/gorm"
)

const ContextKeyProjectRole = "projectRole"

// RequireProjectScope checks that the current user is a member of the project
// referenced by the :id URL parameter, or is a system admin.
// It sets ContextKeyProjectRole to the user's role within the project.
func RequireProjectScope(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		if projectID == "" {
			c.Next()
			return
		}

		role, _ := c.Get(ContextKeyRole)
		roleStr, _ := role.(string)

		if roleStr == "admin" {
			c.Set(ContextKeyProjectRole, "admin")
			c.Next()
			return
		}

		principalType, _ := c.Get(ContextKeyPrincipalType)
		if principalType == PrincipalTypeProjectToken {
			tokenProjectID, _ := c.Get(ContextKeyTokenProjectID)
			if tokenProjectID == projectID {
				c.Set(ContextKeyProjectRole, "project_token")
				c.Next()
				return
			}
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权访问该项目", "project token is scoped to another project"))
			c.Abort()
			return
		}

		userID, _ := c.Get(ContextKeyUserID)
		userIDStr, _ := userID.(string)
		if userIDStr == "" {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权访问该项目", ""))
			c.Abort()
			return
		}

		var roles []string
		err := db.Table("project_members").
			Where("project_id = ? AND user_id = ?", projectID, userIDStr).
			Limit(1).
			Pluck("role", &roles).Error

		if err != nil || len(roles) == 0 {
			response.HandleError(c, response.NewBizError(response.CodeForbidden, "无权访问该项目", ""))
			c.Abort()
			return
		}

		c.Set(ContextKeyProjectRole, roles[0])
		c.Next()
	}
}
