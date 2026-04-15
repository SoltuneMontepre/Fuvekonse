package handlers

import (
	"general-service/internal/common/utils"
	"general-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	services *services.Services
}

func NewAnalyticsHandler(services *services.Services) *AnalyticsHandler {
	return &AnalyticsHandler{services: services}
}

// GetDashboard godoc
// @Summary Get dashboard analytics (admin only)
// @Description Returns consolidated dashboard data: ticket stats, sales timeline, revenue, user count, dealer count, users by country. Single request for all dashboard metrics.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param timeline_days query int false "Days for sales timeline" default(90) minimum(1) maximum(365)
// @Param revenue_days query int false "Days for revenue timeline" default(90) minimum(1) maximum(365)
// @Success 200 "Dashboard analytics"
// @Failure 401 "Unauthorized"
// @Failure 403 "Forbidden"
// @Failure 500 "Internal server error"
// @Router /admin/analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	timelineDays := 90
	revenueDays := 90

	if v := c.Query("timeline_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timelineDays = n
			if timelineDays > 365 {
				timelineDays = 365
			}
		}
	}
	if v := c.Query("revenue_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			revenueDays = n
			if revenueDays > 365 {
				revenueDays = 365
			}
		}
	}

	data, err := h.services.Analytics.GetDashboard(c.Request.Context(), timelineDays, revenueDays)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to load dashboard analytics")
		return
	}

	utils.RespondSuccess(c, data, "Dashboard analytics")
}
