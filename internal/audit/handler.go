package audit

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xjy/zcid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("", h.List)
}

func (h *Handler) List(c *gin.Context) {
	opts := ListOpts{
		Page:     1,
		PageSize: 20,
	}
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			opts.Page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			opts.PageSize = v
		}
	}
	if uid := c.Query("userId"); uid != "" {
		opts.UserID = &uid
	}
	if action := c.Query("action"); action != "" {
		opts.Action = &action
	}
	if rt := c.Query("resourceType"); rt != "" {
		opts.ResourceType = &rt
	}
	if rid := c.Query("resourceId"); rid != "" {
		opts.ResourceID = &rid
	}
	if start := c.Query("startTime"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			opts.StartTime = &t
		}
	}
	if end := c.Query("endTime"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			opts.EndTime = &t
		}
	}

	list, total, err := h.service.List(c.Request.Context(), opts)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	items := make([]gin.H, len(list))
	for i, log := range list {
		item := gin.H{
			"id":           log.ID,
			"action":       log.Action,
			"resourceType": log.ResourceType,
			"result":       log.Result,
			"createdAt":    log.CreatedAt,
		}
		if log.UserID != nil {
			item["userId"] = *log.UserID
		}
		if log.ResourceID != nil {
			item["resourceId"] = *log.ResourceID
		}
		if log.IP != nil {
			item["ip"] = *log.IP
		}
		if log.Detail != nil {
			item["detail"] = *log.Detail
		}
		items[i] = item
	}
	response.Success(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     opts.Page,
		"pageSize": opts.PageSize,
	})
}
