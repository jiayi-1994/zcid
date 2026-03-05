package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/pkg/database"
	"github.com/xjy/zcid/pkg/response"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	rdb       *redis.Client
	k8sStatus string
}

func NewAdminHandler(db *gorm.DB, rdb *redis.Client) *Handler {
	return &Handler{
		db:        db,
		rdb:       rdb,
		k8sStatus: "ok",
	}
}

func (h *Handler) GetSettings(c *gin.Context) {
	response.Success(c, GetSettings())
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}
	UpdateSettings(&req)
	response.Success(c, GetSettings())
}

func (h *Handler) GetHealth(c *gin.Context) {
	hc := CheckHealth(h.db, h.rdb, h.k8sStatus)
	httpStatus := http.StatusOK
	if hc.Status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}
	c.JSON(httpStatus, gin.H{
		"status": hc.Status,
		"checks": hc.Checks,
	})
}

func (h *Handler) GetIntegrationsStatus(c *gin.Context) {
	items := []IntegrationStatus{
		{Name: "database", Status: "ok", Detail: ""},
		{Name: "redis", Status: "ok", Detail: ""},
		{Name: "k8s", Status: "ok", Detail: "TODO: integrate real K8s/Tekton health check"},
	}
	if h.rdb != nil {
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			items[1].Status = "fail"
			items[1].Detail = err.Error()
		}
	} else {
		items[1].Status = "unavailable"
	}
	if err := database.PingPostgres(h.db); err != nil {
		items[0].Status = "fail"
		items[0].Detail = err.Error()
	}
	response.Success(c, gin.H{"integrations": items})
}
