package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	TicketTierID uuid.UUID `gorm:"type:uuid;not null;index" json:"ticket_tier_id"`
	OrderCode   int64     `gorm:"uniqueIndex;not null" json:"order_code"` // payOS order code
	Status      string    `gorm:"size:50;default:'pending'" json:"status"` // pending, paid, cancelled, refunded
	Price       int64     `gorm:"not null" json:"price"` // Price in VNĐ
	TicketTier  *TicketTier `gorm:"foreignKey:TicketTierID" json:"ticket_tier,omitempty"`
	Payments    []Payment `gorm:"foreignKey:TicketID" json:"payments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

