package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/panel/requests"
	"general-service/internal/repositories"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type PanelHandler struct {
	services *services.Services
}

func NewPanelHandler(services *services.Services) *PanelHandler {
	return &PanelHandler{services: services}
}

// CreatePanel godoc
// @Summary Submit a performance / panel application
// @Tags panels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreatePanelRequest true "Panel application"
// @Success 201 {object} map[string]interface{}
// @Router /panels [post]
func (h *PanelHandler) CreatePanel(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	var req requests.CreatePanelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	panel, err := h.services.Panel.CreatePanel(ctx, userID.(string), &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user must have a ticket to submit a panel application":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "user ticket must be approved to submit a panel application":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "failed to check user ticket":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to create panel application")
		}
		return
	}

	utils.RespondCreated(c, panel, "Panel application submitted successfully")
}

// GetMyPanels godoc
// @Summary List my panel applications
// @Tags panels
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /panels [get]
func (h *PanelHandler) GetMyPanels(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	panels, err := h.services.Panel.GetUserPanels(ctx, userID.(string))
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve panels")
		return
	}

	utils.RespondSuccess(c, &panels, "Successfully retrieved panels")
}

// GetPanelByID godoc
// @Summary Get one panel application by ID
// @Tags panels
// @Produce json
// @Security BearerAuth
// @Param id path string true "Panel ID" format(uuid)
// @Router /panels/{id} [get]
func (h *PanelHandler) GetPanelByID(c *gin.Context) {
	ctx := c.Request.Context()
	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	panel, err := h.services.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve panel")
		return
	}

	utils.RespondSuccess(c, panel, "Successfully retrieved panel")
}

// EditPanel godoc
// @Summary Update a pending panel application
// @Tags panels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Panel ID" format(uuid)
// @Param request body requests.UpdatePanelRequest true "Update request"
// @Router /panels/{id} [put]
func (h *PanelHandler) EditPanel(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	var req requests.UpdatePanelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	panel, err := h.services.Panel.EditPanel(ctx, userID.(string), panelID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedPanel) {
			utils.RespondForbidden(c, "Cannot edit non-pending panel or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrPanelNotEditable) {
			utils.RespondForbidden(c, "Cannot edit non-pending panel")
			return
		}
		utils.RespondInternalServerError(c, "Failed to update panel")
		return
	}

	utils.RespondSuccess(c, panel, "Panel updated successfully")
}

// DeletePanel godoc
// @Summary Delete a panel application
// @Tags panels
// @Produce json
// @Security BearerAuth
// @Param id path string true "Panel ID" format(uuid)
// @Router /panels/{id} [delete]
func (h *PanelHandler) DeletePanel(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	err := h.services.Panel.DeletePanel(ctx, userID.(string), panelID)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedPanel) {
			utils.RespondForbidden(c, "Cannot delete approved panel or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrPanelNotEditable) {
			utils.RespondForbidden(c, "Cannot delete approved panel")
			return
		}
		utils.RespondInternalServerError(c, "Failed to delete panel")
		return
	}

	c.JSON(204, nil)
}

func (h *PanelHandler) GetPendingPanels(c *gin.Context) {
	ctx := c.Request.Context()
	panels, err := h.services.Panel.GetPendingPanels(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve pending panels")
		return
	}
	utils.RespondSuccess(c, &panels, "Successfully retrieved pending panels")
}

func (h *PanelHandler) GetApprovedPanels(c *gin.Context) {
	ctx := c.Request.Context()
	panels, err := h.services.Panel.GetApprovedPanels(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve approved panels")
		return
	}
	utils.RespondSuccess(c, &panels, "Successfully retrieved approved panels")
}

func (h *PanelHandler) GetDeniedPanels(c *gin.Context) {
	ctx := c.Request.Context()
	panels, err := h.services.Panel.GetDeniedPanels(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve denied panels")
		return
	}
	utils.RespondSuccess(c, &panels, "Successfully retrieved denied panels")
}

// AssignPanelSchedule godoc
// @Summary Assign slot and start time to an approved panel
// @Tags admin-panels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Panel ID" format(uuid)
// @Param request body requests.AssignPanelScheduleRequest true "Slot and start time"
// @Router /admin/panels/{id}/schedule [patch]
func (h *PanelHandler) AssignPanelSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	var req requests.AssignPanelScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	panel, err := h.services.Panel.AssignPanelSchedule(ctx, panelID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrPanelNotSchedulable) {
			utils.RespondBadRequest(c, "Panel must be approved before assigning a schedule")
			return
		}
		utils.RespondInternalServerError(c, "Failed to assign panel schedule")
		return
	}

	utils.RespondSuccess(c, panel, "Panel schedule updated successfully")
}

func (h *PanelHandler) ApprovePanel(c *gin.Context) {
	ctx := c.Request.Context()
	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	panel, err := h.services.Panel.ApprovePanel(ctx, panelID)
	if err != nil {
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Panel already has approved status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to approve panel")
		return
	}

	utils.RespondSuccess(c, panel, "Panel approved successfully")
}

func (h *PanelHandler) DenyPanel(c *gin.Context) {
	ctx := c.Request.Context()
	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	panel, err := h.services.Panel.DenyPanel(ctx, panelID)
	if err != nil {
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Panel already has denied status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to deny panel")
		return
	}

	utils.RespondSuccess(c, panel, "Panel denied successfully")
}

func (h *PanelHandler) MarkPanelPending(c *gin.Context) {
	ctx := c.Request.Context()
	panelID := c.Param("id")
	if panelID == "" {
		utils.RespondValidationError(c, "Panel ID is required")
		return
	}

	panel, err := h.services.Panel.MarkPanelPending(ctx, panelID)
	if err != nil {
		if errors.Is(err, repositories.ErrPanelNotFound) {
			utils.RespondNotFound(c, "Panel not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Panel already has pending status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to set panel status to pending")
		return
	}

	utils.RespondSuccess(c, panel, "Panel status set to pending successfully")
}
