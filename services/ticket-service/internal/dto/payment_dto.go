package dto

import "time"

// CreatePaymentLinkRequest represents the request to create a payment link
type CreatePaymentLinkRequest struct {
	TierID string `json:"tier_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

// CreatePaymentLinkResponse represents the response after creating a payment link
type CreatePaymentLinkResponse struct {
	CheckoutUrl string `json:"checkout_url"`
	OrderCode   int64  `json:"order_code"`
}

// PaymentWebhookRequest represents the webhook request from PayOS
type PaymentWebhookRequest struct {
	Code string `json:"code"` // Webhook status code ("00" = success)
	Desc string `json:"desc"` // Webhook status description
	Data struct {
		OrderCode     int64  `json:"orderCode"`
		Amount        int    `json:"amount"`
		Description   string `json:"description"`
		Transaction   string `json:"transaction,omitempty"`
		Reference     string `json:"reference"`
		Currency      string `json:"currency"`
		PaymentLinkID string `json:"paymentLinkId"`
		Desc          string `json:"desc"` // Transaction description
	} `json:"data,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// PaymentStatusResponse represents the payment status response
type PaymentStatusResponse struct {
	OrderCode     int64      `json:"order_code"`
	Status        string     `json:"status"`
	Amount        int64      `json:"amount"`
	Description   string     `json:"description"`
	PaymentLinkID string     `json:"payment_link_id,omitempty"`
	CheckoutUrl   string     `json:"checkout_url,omitempty"`
	TransactionID string     `json:"transaction_id,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}
