package models

import (
	"time"

	"github.com/google/uuid"
)

// TicketStatus represents the status of a user's ticket
type TicketStatus string

const (
	TicketStatusPending       TicketStatus = "pending"
	TicketStatusSelfConfirmed TicketStatus = "self_confirmed"
	TicketStatusApproved      TicketStatus = "approved"
	TicketStatusDenied        TicketStatus = "denied"
)

type UserTicket struct {
	Id             uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserId         uuid.UUID    `gorm:"type:uuid;index" json:"user_id"`
	TicketId       uuid.UUID    `gorm:"type:uuid;index" json:"ticket_id"`
	TicketNumber   int          `gorm:"type:int" json:"ticket_number"`                                  // Per-tier sequential number
	ReferenceCode  string       `gorm:"type:varchar(50);uniqueIndex" json:"reference_code"`             // e.g., "T1-0042"
	Status         TicketStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`         // pending, self_confirmed, approved, denied
	ConBadgeName   string       `gorm:"type:varchar(255)" json:"con_badge_name"`                        // Filled after approval
	BadgeImage     string       `gorm:"type:varchar(500)" json:"badge_image"`                           // Filled after approval
	IsFursuiter    bool         `gorm:"default:false" json:"is_fursuiter"`                              // Filled after approval
	IsFursuitStaff bool         `gorm:"default:false" json:"is_fursuit_staff"`                          // Filled after approval
	IsCheckedIn    bool         `gorm:"default:false" json:"is_checked_in"`
	DenialReason   string       `gorm:"type:varchar(500)" json:"denial_reason,omitempty"`               // Reason for denial
	ApprovedAt     *time.Time   `gorm:"index" json:"approved_at,omitempty"`
	DeniedAt       *time.Time   `gorm:"index" json:"denied_at,omitempty"`
	ApprovedBy     *uuid.UUID   `gorm:"type:uuid" json:"approved_by,omitempty"`                         // Staff who approved
	DeniedBy              *uuid.UUID   `gorm:"type:uuid" json:"denied_by,omitempty"`                     // Staff who denied
	UpgradedFromTierID    *uuid.UUID   `gorm:"type:uuid" json:"upgraded_from_tier_id,omitempty"`        // Tier ID before upgrade (nil for fresh purchases)
	PreviousReferenceCode string       `gorm:"type:varchar(50)" json:"previous_reference_code,omitempty"` // Reference code before upgrade
	CreatedAt             time.Time    `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt     time.Time    `gorm:"autoUpdateTime" json:"modified_at"`
	DeletedAt      *time.Time   `gorm:"index" json:"deleted_at,omitempty"`
	IsDeleted      bool         `gorm:"default:false" json:"is_deleted"`
	User           User         `gorm:"foreignKey:UserId" json:"user,omitempty"`
	Ticket         TicketTier   `gorm:"foreignKey:TicketId" json:"ticket,omitempty"`
	Payment        Payment      `gorm:"foreignKey:UserTicketId" json:"-"`
}
