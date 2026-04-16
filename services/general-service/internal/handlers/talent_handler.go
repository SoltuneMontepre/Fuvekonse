package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/talent/requests"
	"general-service/internal/repositories"
	"general-service/internal/services"

	"github.com/gin-gonic/gin"
)

type TalentHandler struct {
	services *services.Services
}

func NewTalentHandler(services *services.Services) *TalentHandler {
	return &TalentHandler{services: services}
}

// CreateTalent godoc
// @Summary Submit a talent application
// @Tags talents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateTalentRequest true "Talent application"
// @Success 201 {object} map[string]interface{}
// @Router /talents [post]
func (h *TalentHandler) CreateTalent(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	var req requests.CreateTalentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	talent, err := h.services.Talent.CreateTalent(ctx, userID.(string), &req)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user must have a ticket to submit a talent application":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "user ticket must be approved to submit a talent application":
			utils.RespondError(c, 403, "FORBIDDEN", errMsg)
		case "failed to check user ticket":
			utils.RespondInternalServerError(c, errMsg)
		default:
			utils.RespondInternalServerError(c, "Failed to create talent application")
		}
		return
	}

	utils.RespondCreated(c, talent, "Talent application submitted successfully")
}

// GetMyTalents godoc
// @Summary List my talent applications
// @Tags talents
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /talents [get]
func (h *TalentHandler) GetMyTalents(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	talents, err := h.services.Talent.GetUserTalents(ctx, userID.(string))
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve talents")
		return
	}

	utils.RespondSuccess(c, &talents, "Successfully retrieved talents")
}

// GetTalentByID godoc
// @Summary Get one talent application by ID
// @Tags talents
// @Produce json
// @Security BearerAuth
// @Param id path string true "Talent ID" format(uuid)
// @Router /talents/{id} [get]
func (h *TalentHandler) GetTalentByID(c *gin.Context) {
	ctx := c.Request.Context()
	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	talent, err := h.services.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		utils.RespondInternalServerError(c, "Failed to retrieve talent")
		return
	}

	utils.RespondSuccess(c, talent, "Successfully retrieved talent")
}

// EditTalent godoc
// @Summary Update a pending talent application
// @Tags talents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Talent ID" format(uuid)
// @Param request body requests.UpdateTalentRequest true "Update request"
// @Router /talents/{id} [put]
func (h *TalentHandler) EditTalent(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	var req requests.UpdateTalentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	talent, err := h.services.Talent.EditTalent(ctx, userID.(string), talentID, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedTalent) {
			utils.RespondForbidden(c, "Cannot edit non-pending talent or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrTalentNotEditable) {
			utils.RespondForbidden(c, "Cannot edit non-pending talent")
			return
		}
		utils.RespondInternalServerError(c, "Failed to update talent")
		return
	}

	utils.RespondSuccess(c, talent, "Talent updated successfully")
}

// DeleteTalent godoc
// @Summary Delete a talent application
// @Tags talents
// @Produce json
// @Security BearerAuth
// @Param id path string true "Talent ID" format(uuid)
// @Router /talents/{id} [delete]
func (h *TalentHandler) DeleteTalent(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	err := h.services.Talent.DeleteTalent(ctx, userID.(string), talentID)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorizedTalent) {
			utils.RespondForbidden(c, "Cannot delete approved talent or you are not the owner")
			return
		}
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrTalentNotEditable) {
			utils.RespondForbidden(c, "Cannot delete approved talent")
			return
		}
		utils.RespondInternalServerError(c, "Failed to delete talent")
		return
	}

	c.JSON(204, nil)
}

func (h *TalentHandler) GetPendingTalents(c *gin.Context) {
	ctx := c.Request.Context()
	talents, err := h.services.Talent.GetPendingTalents(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve pending talents")
		return
	}
	utils.RespondSuccess(c, &talents, "Successfully retrieved pending talents")
}

func (h *TalentHandler) GetApprovedTalents(c *gin.Context) {
	ctx := c.Request.Context()
	talents, err := h.services.Talent.GetApprovedTalents(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve approved talents")
		return
	}
	utils.RespondSuccess(c, &talents, "Successfully retrieved approved talents")
}

func (h *TalentHandler) GetDeniedTalents(c *gin.Context) {
	ctx := c.Request.Context()
	talents, err := h.services.Talent.GetDeniedTalents(ctx)
	if err != nil {
		utils.RespondInternalServerError(c, "Failed to retrieve denied talents")
		return
	}
	utils.RespondSuccess(c, &talents, "Successfully retrieved denied talents")
}

// AssignTalentSchedule godoc
// @Summary Assign slot and start time to an approved talent
// @Tags admin-talents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Talent ID" format(uuid)
// @Param request body requests.AssignTalentScheduleRequest true "Slot and start time"
// @Router /admin/talents/{id}/schedule [patch]
func (h *TalentHandler) AssignTalentSchedule(c *gin.Context) {
	ctx := c.Request.Context()
	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	var req requests.AssignTalentScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	talent, err := h.services.Talent.AssignTalentSchedule(ctx, talentID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrTalentNotSchedulable) {
			utils.RespondBadRequest(c, "Talent must be approved before assigning a schedule")
			return
		}
		utils.RespondInternalServerError(c, "Failed to assign talent schedule")
		return
	}

	utils.RespondSuccess(c, talent, "Talent schedule updated successfully")
}

func (h *TalentHandler) ApproveTalent(c *gin.Context) {
	ctx := c.Request.Context()
	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	talent, err := h.services.Talent.ApproveTalent(ctx, talentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Talent already has approved status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to approve talent")
		return
	}

	utils.RespondSuccess(c, talent, "Talent approved successfully")
}

func (h *TalentHandler) DenyTalent(c *gin.Context) {
	ctx := c.Request.Context()
	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	talent, err := h.services.Talent.DenyTalent(ctx, talentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Talent already has denied status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to deny talent")
		return
	}

	utils.RespondSuccess(c, talent, "Talent denied successfully")
}

func (h *TalentHandler) MarkTalentPending(c *gin.Context) {
	ctx := c.Request.Context()
	talentID := c.Param("id")
	if talentID == "" {
		utils.RespondValidationError(c, "Talent ID is required")
		return
	}

	talent, err := h.services.Talent.MarkTalentPending(ctx, talentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTalentNotFound) {
			utils.RespondNotFound(c, "Talent not found")
			return
		}
		if errors.Is(err, services.ErrStatusUnchanged) {
			utils.RespondError(c, 409, "STATUS_UNCHANGED", "Talent already has pending status")
			return
		}
		utils.RespondInternalServerError(c, "Failed to set talent status to pending")
		return
	}

	utils.RespondSuccess(c, talent, "Talent status set to pending successfully")
}
