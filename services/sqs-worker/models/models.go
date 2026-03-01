package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TicketStatus matches general-service schema.
type TicketStatus string

const (
	TicketStatusPending       TicketStatus = "pending"
	TicketStatusSelfConfirmed TicketStatus = "self_confirmed"
	TicketStatusApproved      TicketStatus = "approved"
	TicketStatusDenied        TicketStatus = "denied"
)

// User minimal for blacklist and purchase checks (table: users).
type User struct {
	Id            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	IsBlacklisted bool       `gorm:"default:false;index"`
	BlacklistedAt *time.Time `gorm:"index"`
	BlacklistReason string   `gorm:"type:varchar(500)"`
	DenialCount   int        `gorm:"type:int;default:0"`
	IsDeleted     bool       `gorm:"default:false"`
}

// TicketTier minimal for purchase/upgrade (table: ticket_tiers).
type TicketTier struct {
	Id        uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TierCode  string          `gorm:"type:varchar(10);uniqueIndex"`
	Stock     int             `gorm:"type:int"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2)"`
	IsActive  bool            `gorm:"default:true"`
	IsDeleted bool            `gorm:"default:false"`
}

// UserTicket for job processing (table: user_tickets).
type UserTicket struct {
	Id             uuid.UUID    `gorm:"type:uuid;primaryKey"`
	UserId         uuid.UUID    `gorm:"type:uuid;index"`
	TicketId       uuid.UUID    `gorm:"type:uuid;index"`
	TicketNumber   int          `gorm:"type:int"`
	ReferenceCode  string       `gorm:"type:varchar(50);uniqueIndex"`
	Status         TicketStatus `gorm:"type:varchar(20);default:'pending';index"`
	ConBadgeName   string       `gorm:"type:varchar(255)"`
	BadgeImage     string       `gorm:"type:varchar(500)"`
	IsFursuiter    bool         `gorm:"default:false"`
	IsFursuitStaff bool         `gorm:"default:false"`
	DenialReason   string       `gorm:"type:varchar(500)"`
	ApprovedAt     *time.Time   `gorm:"index"`
	DeniedAt       *time.Time   `gorm:"index"`
	ApprovedBy     *uuid.UUID   `gorm:"type:uuid"`
	DeniedBy       *uuid.UUID   `gorm:"type:uuid"`
	IsDeleted      bool         `gorm:"default:false"`
	CreatedAt      time.Time    `gorm:"autoCreateTime"`
	ModifiedAt     time.Time    `gorm:"autoUpdateTime"`
}
