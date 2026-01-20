package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TicketTier struct {
	Id          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	TierCode    string          `gorm:"type:varchar(10);uniqueIndex" json:"tier_code"` // e.g., "T1", "T2", "T3"
	Price       decimal.Decimal `gorm:"type:decimal(10,2)" json:"price"`
	Description string          `gorm:"type:varchar(500)" json:"description"`
	TicketName  string          `gorm:"type:varchar(255)" json:"ticket_name"`
	Benefits    string          `gorm:"type:text" json:"benefits"` // JSON array of benefit strings
	Stock       int             `gorm:"type:int;check:stock >= 0" json:"stock"`
	IsActive    bool            `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt  time.Time       `gorm:"autoUpdateTime" json:"modified_at"`
	DeletedAt   *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
	IsDeleted   bool            `gorm:"default:false" json:"is_deleted"`
	UserTickets []UserTicket    `gorm:"foreignKey:TicketId" json:"-"`
}
