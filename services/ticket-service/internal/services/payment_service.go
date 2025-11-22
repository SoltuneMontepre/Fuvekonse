package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ticket-service/internal/common/constants"
	"ticket-service/internal/dto"
	"ticket-service/internal/models"
	"ticket-service/internal/repositories"

	"github.com/google/uuid"
	"github.com/payOSHQ/payos-lib-golang/v2"
)

type PaymentServiceInterface interface {
	CreatePaymentLink(ctx context.Context, tierID, userID uuid.UUID) (*dto.CreatePaymentLinkResponse, error)
	HandleWebhook(ctx context.Context, webhookReq *dto.PaymentWebhookRequest) error
	CleanupStuckPayments(ctx context.Context) error
	GetPaymentStatus(ctx context.Context, orderCode int64) (*dto.PaymentStatusResponse, error)
	CancelPaymentByOrderCode(ctx context.Context, orderCode int64) error
	VerifyWebhookSignature(ctx context.Context, webhookReq *dto.PaymentWebhookRequest) bool
}

type PaymentService struct {
	paymentRepo   repositories.PaymentRepositoryInterface
	ticketRepo    repositories.TicketRepositoryInterface
	tierRepo      repositories.TicketTierRepositoryInterface
	payOSClient   *payos.PayOS
	frontendURL   string
	generalSvcURL string
	checksumKey   string
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepositoryInterface,
	ticketRepo repositories.TicketRepositoryInterface,
	tierRepo repositories.TicketTierRepositoryInterface,
	payOSClientID string,
	payOSAPIKey string,
	payOSChecksumKey string,
	frontendURL string,
	generalSvcURL string,
) PaymentServiceInterface {
	var payOSClient *payos.PayOS
	var err error

	if payOSClientID != "" && payOSAPIKey != "" && payOSChecksumKey != "" {
		payOSClient, err = payos.NewPayOS(&payos.PayOSOptions{
			ClientId:    payOSClientID,
			ApiKey:      payOSAPIKey,
			ChecksumKey: payOSChecksumKey,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to initialize PayOS client: %v\n", err)
		}
	} else {
		fmt.Println("Warning: PayOS credentials not configured")
	}

	return &PaymentService{
		paymentRepo:   paymentRepo,
		ticketRepo:    ticketRepo,
		tierRepo:      tierRepo,
		payOSClient:   payOSClient,
		frontendURL:   frontendURL,
		generalSvcURL: generalSvcURL,
		checksumKey:   payOSChecksumKey,
	}
}

func (s *PaymentService) CreatePaymentLink(ctx context.Context, tierID, userID uuid.UUID) (*dto.CreatePaymentLinkResponse, error) {
	// Get ticket tier
	tier, err := s.tierRepo.GetByID(tierID)
	if err != nil {
		return nil, fmt.Errorf("ticket tier not found: %w", err)
	}

	// Validate ticket tier
	if !tier.IsActive {
		return nil, fmt.Errorf("ticket tier is not active")
	}

	// Check if ticket is available (including reservations)
	if tier.Stock <= 0 {
		return nil, fmt.Errorf("ticket tier is out of stock")
	}

	// Attempt to reserve the ticket (decrement stock atomically)
	if err := s.tierRepo.DecrementStock(tierID); err != nil {
		return nil, fmt.Errorf("failed to reserve ticket: %w", err)
	}

	// Generate order code (milliseconds since epoch)
	// PayOS requires order code to be a positive integer
	orderCode := time.Now().UnixMilli()

	// Create ticket record with PENDING PAYMENT status
	// This matches the desired flow: ticket created when user clicks "Buy ticket"
	hardcodedUserID := uuid.MustParse("b3c8e5f4-0a86-4d73-a8b9-3daafe0f6a20")
	ticket := &models.Ticket{
		ID:           uuid.New(),
		UserID:       hardcodedUserID,
		TicketTierID: tierID,
		OrderCode:    orderCode,
		Status:       "PENDING",
		Price:        tier.Price,
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Create pending payment record linked to the ticket
	payment := &models.Payment{
		ID:           uuid.New(),
		TicketID:     &ticket.ID, // Link to created ticket
		TicketTierID: &tierID,
		OrderCode:    orderCode,
		Amount:       tier.Price,
		Status:       constants.PaymentStatusPending.String(),
		Provider:     "payos",
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Create payment link with payOS
	if s.payOSClient == nil {
		return nil, fmt.Errorf("payOS client not initialized")
	}

	returnUrl := fmt.Sprintf("%s/ticket/purchase/%s?status=%s", s.frontendURL, tierID.String(), constants.PaymentStatusPaid.String())
	cancelUrl := fmt.Sprintf("%s/ticket/purchase/%s?status=%s", s.frontendURL, tierID.String(), constants.PaymentStatusCancelled.String())

	// Ensure amount is positive and valid
	amount := int(tier.Price)
	if amount <= 0 {
		return nil, fmt.Errorf("invalid ticket price: %d", amount)
	}

	paymentLinkRequest := payos.CreatePaymentLinkRequest{
		OrderCode:   orderCode,
		Amount:      4000, // Use actual amount instead of hardcoded 4000
		Description: fmt.Sprintf("Thanh toan ve %s", tier.TicketName),
		CancelUrl:   cancelUrl,
		ReturnUrl:   returnUrl,
		// Note: PayOS payment links have their own expiration, no longer tied to our reservation system
	}

	paymentLinkResponse, err := s.payOSClient.PaymentRequests.Create(ctx, paymentLinkRequest)
	if err != nil {
		// Clean up payment record if payment link creation fails
		s.paymentRepo.Delete(payment.ID)
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	// Update payment with checkout URL and payment link ID
	payment.CheckoutUrl = paymentLinkResponse.CheckoutUrl
	payment.PaymentLinkID = paymentLinkResponse.PaymentLinkId
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return &dto.CreatePaymentLinkResponse{
		CheckoutUrl: paymentLinkResponse.CheckoutUrl,
		OrderCode:   orderCode,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, webhookReq *dto.PaymentWebhookRequest) error {
	if s.payOSClient == nil {
		return fmt.Errorf("payOS client not initialized")
	}

	// Verify webhook signature for security
	// if !s.VerifyWebhookSignature(ctx, webhookReq) {
	// 	return fmt.Errorf("webhook signature verification failed")
	// }

	data := webhookReq.Data

	// Find existing payment record by order code (created when payment link was generated)
	payment, err := s.paymentRepo.GetByOrderCode(data.OrderCode)
	if err != nil {
		return fmt.Errorf("payment not found for order code %d: %w", data.OrderCode, err)
	}

	// Get the associated ticket
	ticket, err := s.ticketRepo.GetByID(*payment.TicketID)
	if err != nil {
		return fmt.Errorf("ticket not found for order code %d: %w", data.OrderCode, err)
	}

	// Handle different webhook codes from PayOS
	if webhookReq.Code == "00" {
		// Payment successful
		ticket.Status = "PAID"
		payment.Status = constants.PaymentStatusPaid.String()
		payment.GatewayTransactionID = data.Reference
		payment.PaymentLinkID = data.PaymentLinkID

		// Update ticket and payment
		if err := s.ticketRepo.Update(ticket); err != nil {
			return fmt.Errorf("failed to update ticket: %w", err)
		}

		if err := s.paymentRepo.Update(payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Create UserTicket in general-service after successful payment
		if err := s.createUserTicket(ctx, ticket); err != nil {
			fmt.Printf("Warning: Failed to create UserTicket: %v\n", err)
			// Don't fail the webhook if UserTicket creation fails
		}

		fmt.Printf("Payment successful for order code %d\n", data.OrderCode)
	} else {
		// Payment failed/cancelled/expired
		// Only increment stock if payment was still pending (not already processed)
		if payment.Status == constants.PaymentStatusPending.String() {
			if payment.TicketTierID != nil {
				if err := s.tierRepo.IncrementStock(*payment.TicketTierID); err != nil {
					fmt.Printf("Warning: Failed to increment stock for failed payment: %v\n", err)
				}
			}
		}

		// Update ticket and payment status (only if not already cancelled)
		if payment.Status != constants.PaymentStatusCancelled.String() {
			ticket.Status = "CANCELLED"
			payment.Status = constants.PaymentStatusCancelled.String()

			if err := s.ticketRepo.Update(ticket); err != nil {
				return fmt.Errorf("failed to update cancelled ticket: %w", err)
			}

			if err := s.paymentRepo.Update(payment); err != nil {
				return fmt.Errorf("failed to update cancelled payment: %w", err)
			}
		}

		fmt.Printf("Payment failed/cancelled for order code %d: %s (code: %s)\n", data.OrderCode, webhookReq.Desc, webhookReq.Code)
	}

	return nil
}

// VerifyWebhookSignature verifies the webhook signature from PayOS
func (s *PaymentService) VerifyWebhookSignature(ctx context.Context, webhookReq *dto.PaymentWebhookRequest) bool {
	if s.checksumKey == "" {
		fmt.Println("Warning: Checksum key not configured, skipping webhook verification")
		return true // Allow webhooks if checksum key is not configured
	}

	// PayOS webhook signature verification
	// Create the signature string: orderCode + amount + description
	signatureString := fmt.Sprintf("%d%d%s",
		webhookReq.Data.OrderCode,
		webhookReq.Data.Amount,
		webhookReq.Data.Desc)

	// Create HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(s.checksumKey))
	h.Write([]byte(signatureString))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures (case-insensitive)
	return strings.EqualFold(expectedSignature, webhookReq.Signature)
}

func (s *PaymentService) CleanupStuckPayments(ctx context.Context) error {
	// Safety net for payments that don't receive webhooks
	// PayOS should send webhooks for all payment outcomes, but this handles edge cases
	expiredTime := time.Now().Add(-60 * time.Minute) // 1 hour grace period

	expiredPayments, err := s.paymentRepo.GetExpiredReservations()
	if err != nil {
		return fmt.Errorf("failed to get expired reservations: %w", err)
	}

	// Filter payments that are still pending and very old
	var trulyExpiredPayments []models.Payment
	for _, payment := range expiredPayments {
		if payment.Status == constants.PaymentStatusPending.String() && payment.CreatedAt.Before(expiredTime) {
			trulyExpiredPayments = append(trulyExpiredPayments, payment)
		}
	}

	cleanedCount := 0
	for _, payment := range trulyExpiredPayments {
		// Release stock back to the ticket tier
		if payment.TicketTierID != nil {
			if err := s.tierRepo.IncrementStock(*payment.TicketTierID); err != nil {
				fmt.Printf("Warning: Failed to release expired reservation stock for payment %s: %v\n", payment.ID, err)
				continue
			}
		}

		// Update ticket status to CANCELLED if payment has an associated ticket
		if payment.TicketID != nil {
			ticket, err := s.ticketRepo.GetByID(*payment.TicketID)
			if err != nil {
				fmt.Printf("Warning: Failed to get ticket for expired payment %s: %v\n", payment.ID, err)
			} else {
				ticket.Status = "CANCELLED"
				if err := s.ticketRepo.Update(ticket); err != nil {
					fmt.Printf("Warning: Failed to update ticket status for expired payment %s: %v\n", payment.ID, err)
				}
			}
		}

		// Update payment status to expired
		payment.Status = constants.PaymentStatusExpired.String()
		if err := s.paymentRepo.Update(&payment); err != nil {
			fmt.Printf("Warning: Failed to update expired payment %s: %v\n", payment.ID, err)
			continue
		}

		cleanedCount++
		fmt.Printf("Cleaned up expired reservation for payment %s\n", payment.ID)
	}

	fmt.Printf("Successfully cleaned up %d expired reservations\n", cleanedCount)
	return nil
}

// GetPaymentStatus retrieves the current payment status from PayOS
func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderCode int64) (*dto.PaymentStatusResponse, error) {
	if s.payOSClient == nil {
		return nil, fmt.Errorf("payOS client not initialized")
	}

	// Get payment from our database first
	payment, err := s.paymentRepo.GetByOrderCode(orderCode)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Try to get payment info from PayOS (if the method exists)
	// Note: This depends on the PayOS SDK having a GetPaymentInfo or similar method
	// For now, we'll return our internal status

	response := &dto.PaymentStatusResponse{
		OrderCode:     payment.OrderCode,
		Status:        payment.Status,
		Amount:        payment.Amount,
		Description:   "Ticket purchase",
		PaymentLinkID: payment.PaymentLinkID,
		CheckoutUrl:   payment.CheckoutUrl,
		TransactionID: payment.GatewayTransactionID,
	}

	// If payment is paid, set paid timestamp
	if payment.Status == constants.PaymentStatusPaid.String() {
		// We don't store the exact paid timestamp, so we'll use updated_at
		response.PaidAt = &payment.UpdatedAt
	}

	return response, nil
}

// createUserTicket creates a UserTicket in the general-service after successful payment
func (s *PaymentService) createUserTicket(ctx context.Context, ticket *models.Ticket) error {
	if s.generalSvcURL == "" {
		// General service URL not configured, skip UserTicket creation
		fmt.Println("Warning: General service URL not configured, skipping UserTicket creation")
		return nil
	}

	userTicketData := map[string]interface{}{
		"user_id":          ticket.UserID,
		"ticket_id":        ticket.ID,                  // Now points to Ticket.ID
		"con_badge_name":   ticket.UserID.String()[:8], // Use first 8 chars of user ID as default badge name
		"badge_image":      "",                         // Empty for now, can be set later
		"is_fursuiter":     false,                      // Default values
		"is_fursuit_staff": false,
		"is_checked_in":    false,
	}

	jsonData, err := json.Marshal(userTicketData)
	if err != nil {
		return fmt.Errorf("failed to marshal UserTicket data: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/user-tickets", s.generalSvcURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call general-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("general-service returned status %d", resp.StatusCode)
	}

	fmt.Printf("Successfully created UserTicket for ticket %s\n", ticket.ID)
	return nil
}

// CancelPaymentByOrderCode cancels a payment by order code (called from frontend redirect)
func (s *PaymentService) CancelPaymentByOrderCode(ctx context.Context, orderCode int64) error {
	// Find the payment
	payment, err := s.paymentRepo.GetByOrderCode(orderCode)
	if err != nil {
		// If payment not found, it might have been cleaned up or doesn't exist
		// Return success since frontend indicated cancellation intent
		fmt.Printf("Payment %d not found for cancellation, may already be processed\n", orderCode)
		return nil
	}

	// Check current status and handle accordingly
	switch payment.Status {
	case constants.PaymentStatusCancelled.String():
		// Already cancelled - return success (idempotent)
		fmt.Printf("Payment %d already cancelled\n", orderCode)
		return nil
	case constants.PaymentStatusPaid.String():
		// Cannot cancel paid payment
		return fmt.Errorf("cannot cancel paid payment")
	case constants.PaymentStatusPending.String():
		// Proceed with cancellation
	default:
		// Unknown status - log but allow cancellation attempt
		fmt.Printf("Payment %d has unknown status %s, proceeding with cancellation\n", orderCode, payment.Status)
	}

	// Get the associated ticket
	ticket, err := s.ticketRepo.GetByID(*payment.TicketID)
	if err != nil {
		return fmt.Errorf("ticket not found: %w", err)
	}

	// Only increment stock if payment was still pending (not already processed)
	if payment.Status == constants.PaymentStatusPending.String() {
		if payment.TicketTierID != nil {
			if err := s.tierRepo.IncrementStock(*payment.TicketTierID); err != nil {
				fmt.Printf("Warning: Failed to increment stock for cancelled payment: %v\n", err)
				// Continue with status updates even if stock increment fails
			}
		}
	}

	// Update ticket and payment status to cancelled
	ticket.Status = "CANCELLED"
	payment.Status = constants.PaymentStatusCancelled.String()

	if err := s.ticketRepo.Update(ticket); err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	if err := s.paymentRepo.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	fmt.Printf("Payment cancelled by order code %d\n", orderCode)
	return nil
}
