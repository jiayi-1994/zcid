package admin

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xjy/zcid/internal/signal"
	"github.com/xjy/zcid/pkg/database"
	"github.com/xjy/zcid/pkg/response"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	rdb       *redis.Client
	k8sStatus string
	signals   *signal.Service
}

func NewAdminHandler(db *gorm.DB, rdb *redis.Client) *Handler {
	return &Handler{
		db:        db,
		rdb:       rdb,
		k8sStatus: "ok",
	}
}

func (h *Handler) SetSignalService(signals *signal.Service) {
	h.signals = signals
}

// GetSettings godoc
// @Summary Get system settings
// @Description Retrieve the current system settings (admin only)
// @Tags admin
// @Produce json
// @Success 200 {object} response.Response{data=SystemSettings}
// @Router /api/v1/admin/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	response.Success(c, GetSettings())
}

// UpdateSettings godoc
// @Summary Update system settings
// @Description Update the system settings (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param request body SystemSettings true "System settings payload"
// @Success 200 {object} response.Response{data=SystemSettings}
// @Failure 400 {object} response.Response
// @Router /api/v1/admin/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context) {
	var req SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, response.NewBizError(response.CodeValidation, "invalid request", err.Error()))
		return
	}
	UpdateSettings(&req)
	response.Success(c, GetSettings())
}

// GetHealth godoc
// @Summary Get system health
// @Description Check the health status of the system and its dependencies
// @Tags admin
// @Produce json
// @Success 200 {object} object{status=string,checks=object}
// @Failure 503 {object} object{status=string,checks=object}
// @Router /api/v1/admin/health [get]
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

// GetIntegrationsStatus godoc
// @Summary Get integrations status
// @Description Check the status of all system integrations (database, redis, k8s)
// @Tags admin
// @Produce json
// @Success 200 {object} response.Response{data=object{integrations=[]IntegrationStatus}}
// @Router /api/v1/admin/integrations/status [get]
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
	h.recordIntegrationSignals(c.Request.Context(), items)
	response.Success(c, gin.H{"integrations": items})
}

func (h *Handler) recordIntegrationSignals(ctx context.Context, items []IntegrationStatus) {
	if h.signals == nil || h.db == nil {
		return
	}
	var projectIDs []string
	if err := h.db.WithContext(ctx).Table("projects").Where("status != ?", "deleted").Pluck("id", &projectIDs).Error; err != nil {
		slog.Warn("failed to list projects for integration health signals", slog.Any("error", err))
		return
	}
	staleAfter := time.Now().Add(10 * time.Minute)
	for _, projectID := range projectIDs {
		for _, item := range items {
			status, severity := integrationSignalStatus(item.Status)
			if _, err := h.signals.Record(ctx, signal.RecordInput{
				ProjectID:  projectID,
				TargetType: signal.TargetIntegration,
				TargetID:   item.Name,
				Source:     "admin-health",
				Status:     status,
				Severity:   severity,
				Reason:     "integration." + item.Status,
				Message:    item.Detail,
				ObservedValue: map[string]any{
					"name":   item.Name,
					"status": item.Status,
					"detail": item.Detail,
				},
				StaleAfter: &staleAfter,
			}); err != nil {
				slog.Warn("failed to record integration health signal", slog.Any("error", err), slog.String("projectID", projectID), slog.String("integration", item.Name))
			}
		}
	}
}

func integrationSignalStatus(status string) (signal.Status, signal.Severity) {
	switch status {
	case "ok":
		return signal.StatusHealthy, signal.SeverityInfo
	case "fail":
		return signal.StatusDegraded, signal.SeverityCritical
	default:
		return signal.StatusUnknown, signal.SeverityWarning
	}
}
