package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TicketID             *uuid.UUID     `gorm:"type:uuid;index" json:"ticket_id,omitempty"`      // Nullable until ticket is created
	TicketTierID         *uuid.UUID     `gorm:"type:uuid;index" json:"ticket_tier_id,omitempty"` // Store tier ID for webhook
	OrderCode            int64          `gorm:"index;not null" json:"order_code"`                // payOS order code
	Amount               int64          `gorm:"not null" json:"amount"`                          // Amount in VNĐ
	Status               string         `gorm:"size:50;default:'pending'" json:"status"`         // PaymentStatus constants
	PaymentLinkID        string         `gorm:"size:255" json:"payment_link_id,omitempty"`
	CheckoutUrl          string         `gorm:"type:text" json:"checkout_url,omitempty"`
	GatewayTransactionID string         `gorm:"size:255" json:"gateway_transaction_id,omitempty"`
	Provider             string         `gorm:"size:50;default:'payos'" json:"provider"`
	RawResponse          string         `gorm:"type:text" json:"raw_response,omitempty"`
	Ticket               *Ticket        `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
