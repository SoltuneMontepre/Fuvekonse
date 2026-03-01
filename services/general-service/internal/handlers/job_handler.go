package handlers

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/ticket/requests"
	"general-service/internal/queue"
	"general-service/internal/repositories"
	"general-service/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProcessTicketJob handles internal ticket job requests from the SQS worker.
// Expects X-Internal-Api-Key header and JSON body matching queue.TicketJobMessage.
func (h *TicketHandler) ProcessTicketJob(c *gin.Context) {
	ctx := c.Request.Context()

	var msg queue.TicketJobMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		utils.RespondBadRequest(c, "Invalid job payload: "+err.Error())
		return
	}

	switch msg.Action {
	case queue.ActionPurchaseTicket:
		_, err := h.services.Ticket.PurchaseTicket(ctx, msg.UserID, &requests.PurchaseTicketRequest{TierID: msg.TierID})
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess[any](c, nil, "Purchase processed")
	case queue.ActionConfirmPayment:
		ticket, err := h.services.Ticket.ConfirmPayment(ctx, msg.UserID)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess(c, ticket, "Confirm payment processed")
	case queue.ActionCancelTicket:
		err := h.services.Ticket.CancelTicket(ctx, msg.UserID)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess[any](c, nil, "Cancel processed")
	case queue.ActionUpdateBadge:
		req := &requests.UpdateBadgeDetailsRequest{
			ConBadgeName:   msg.ConBadgeName,
			BadgeImage:     msg.BadgeImage,
			IsFursuiter:    msg.IsFursuiter,
			IsFursuitStaff: msg.IsFursuitStaff,
		}
		ticket, err := h.services.Ticket.UpdateBadgeDetails(ctx, msg.UserID, req)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess(c, ticket, "Badge update processed")
	case queue.ActionApproveTicket:
		ticket, err := h.services.Ticket.ApproveTicket(ctx, msg.TicketID, msg.StaffID)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess(c, ticket, "Approve processed")
	case queue.ActionDenyTicket:
		req := &requests.DenyTicketRequest{Reason: msg.Reason}
		ticket, err := h.services.Ticket.DenyTicket(ctx, msg.TicketID, msg.StaffID, req)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess(c, ticket, "Deny processed")
	case queue.ActionUpgradeTicket:
		result, err := h.services.Ticket.UpgradeTicket(ctx, msg.UserID, &requests.UpgradeTicketRequest{NewTierID: msg.TierID})
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess(c, result, "Upgrade processed")
	case queue.ActionBlacklistUser:
		req := &requests.BlacklistUserRequest{Reason: msg.Reason}
		err := h.services.Ticket.BlacklistUser(ctx, msg.TargetUserID, req)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess[any](c, nil, "Blacklist processed")
	case queue.ActionUnblacklistUser:
		err := h.services.Ticket.UnblacklistUser(ctx, msg.TargetUserID)
		if err != nil {
			respondTicketJobError(c, err)
			return
		}
		utils.RespondSuccess[any](c, nil, "Unblacklist processed")
	default:
		utils.RespondBadRequest(c, "Unknown ticket job action: "+string(msg.Action))
	}
}

func respondTicketJobError(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, services.ErrInvalidTierID), errors.Is(err, services.ErrInvalidTicketID), errors.Is(err, services.ErrInvalidUserID):
		utils.RespondBadRequest(c, err.Error())
	case errors.Is(err, repositories.ErrTicketTierNotFound), errors.Is(err, repositories.ErrTicketNotFound), errors.Is(err, services.ErrNoTicketFound):
		utils.RespondNotFound(c, err.Error())
	case errors.Is(err, repositories.ErrOutOfStock):
		utils.RespondError(c, http.StatusConflict, "OUT_OF_STOCK", err.Error())
	case errors.Is(err, repositories.ErrUserAlreadyHasTicket):
		utils.RespondError(c, http.StatusConflict, "ALREADY_HAS_TICKET", err.Error())
	case errors.Is(err, repositories.ErrUserBlacklisted):
		utils.RespondForbidden(c, err.Error())
	case errors.Is(err, repositories.ErrInvalidTicketStatus):
		utils.RespondError(c, http.StatusConflict, "INVALID_STATUS", err.Error())
	case errors.Is(err, repositories.ErrCannotDowngrade):
		utils.RespondError(c, http.StatusConflict, "CANNOT_DOWNGRADE", err.Error())
	case errors.Is(err, repositories.ErrTicketDenied):
		utils.RespondError(c, http.StatusConflict, "TICKET_DENIED", err.Error())
	default:
		log.Printf("Job processing failed (unhandled error): %v", err)
		utils.RespondInternalServerError(c, "Job processing failed")
	}
}
