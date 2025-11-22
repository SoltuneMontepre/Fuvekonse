package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"ticket-service/internal/dto"
	"ticket-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService services.PaymentServiceInterface
}

func NewPaymentHandler(paymentService services.PaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentLink godoc
// @Summary Create a payment link for ticket purchase
// @Description Create a payment link using payOS for purchasing a ticket
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePaymentLinkRequest true "Payment link request"
// @Success 200 {object} dto.CreatePaymentLinkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/payment-link [post]
func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	var req dto.CreatePaymentLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from JWT authentication context
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User ID not found in token"})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid user ID in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid user ID format in token"})
		return
	}

	// Parse tier ID from request
	tierID, err := uuid.Parse(req.TierID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid tier ID format"})
		return
	}

	response, err := h.paymentService.CreatePaymentLink(c.Request.Context(), tierID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CleanupStuckPayments godoc
// @Summary Clean up stuck payments (safety net)
// @Description Safety net for payments that don't receive webhooks (should be called periodically)
// @Tags payments
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/cleanup-stuck [post]
func (h *PaymentHandler) CleanupStuckPayments(c *gin.Context) {
	if err := h.paymentService.CleanupStuckPayments(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Stuck payments cleaned up successfully"})
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get the current status of a payment by order code
// @Tags payments
// @Produce json
// @Param orderCode path int true "Order Code"
// @Success 200 {object} dto.PaymentStatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/status/{orderCode} [get]
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	orderCodeStr := c.Param("orderCode")
	orderCode, err := strconv.ParseInt(orderCodeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order code"})
		return
	}

	status, err := h.paymentService.GetPaymentStatus(c.Request.Context(), orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// HandleWebhook godoc
// @Summary Handle payOS webhook
// @Description Handle payment webhook from payOS
// @Tags payments
// @Accept json
// @Produce json
// @Param webhook body dto.PaymentWebhookRequest true "Webhook request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments/webhook [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	var webhookReq dto.PaymentWebhookRequest

	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		c.JSON(http.StatusNoContent, dto.MessageResponse{Message: "Webhook received"})
		return
	}

	if err := h.paymentService.HandleWebhook(c.Request.Context(), &webhookReq); err != nil {
		c.JSON(http.StatusNoContent, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Webhook processed successfully"})
}

// CancelPaymentByOrderCode godoc
// @Summary Cancel payment by order code (frontend redirect)
// @Description Cancel a payment when user is redirected back from PayOS with cancellation
// @Tags payments
// @Produce json
// @Param orderCode query int64 true "Order Code from PayOS redirect"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /payments/cancel-by-order [post]
func (h *PaymentHandler) CancelPaymentByOrderCode(c *gin.Context) {
	orderCodeStr := c.Query("orderCode")
	if orderCodeStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "orderCode query parameter required"})
		return
	}

	orderCode, err := strconv.ParseInt(orderCodeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order code"})
		return
	}

	if err := h.paymentService.CancelPaymentByOrderCode(c.Request.Context(), orderCode); err != nil {
		// Only return error for paid payments that can't be cancelled
		if err.Error() == "cannot cancel paid payment" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			return
		}
		// For other cases (payment not found, already cancelled), treat as success
		// since frontend indicated cancellation intent
		fmt.Printf("Warning: CancelPaymentByOrderCode had issue but treated as success: %v\n", err)
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Payment cancellation processed"})
}
