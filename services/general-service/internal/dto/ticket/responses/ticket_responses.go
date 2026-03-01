package responses

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TicketTierResponse represents a ticket tier for display
type TicketTierResponse struct {
	ID          uuid.UUID       `json:"id"`
	TierCode    string          `json:"tier_code"`
	TicketName  string          `json:"ticket_name"`
	Description string          `json:"description"`
	Benefits    []string        `json:"benefits"` // Parsed from JSON
	Price       decimal.Decimal `json:"price"`
	Stock       int             `json:"stock"`
	IsActive    bool            `json:"is_active"`
}

// UserTicketResponse represents a user's ticket
type UserTicketResponse struct {
	ID             uuid.UUID  `json:"id"`
	ReferenceCode  string     `json:"reference_code"`
	Status         string     `json:"status"`
	TicketNumber   int        `json:"ticket_number"`
	ConBadgeName   string     `json:"con_badge_name,omitempty"`
	BadgeImage     string     `json:"badge_image,omitempty"`
	IsFursuiter    bool       `json:"is_fursuiter"`
	IsFursuitStaff bool       `json:"is_fursuit_staff"`
	IsCheckedIn    bool       `json:"is_checked_in"`
	DenialReason          string     `json:"denial_reason,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	ApprovedAt            *time.Time `json:"approved_at,omitempty"`
	DeniedAt              *time.Time `json:"denied_at,omitempty"`
	UpgradedFromTierID    *uuid.UUID `json:"upgraded_from_tier_id,omitempty"`
	PreviousReferenceCode string     `json:"previous_reference_code,omitempty"`

	// Tier info
	Tier *TicketTierResponse `json:"tier,omitempty"`

	// User info (for admin view)
	User *TicketUserResponse `json:"user,omitempty"`
}

// UpgradeTicketResponse contains the upgraded ticket plus pricing info for the frontend.
type UpgradeTicketResponse struct {
	Ticket          *UserTicketResponse `json:"ticket"`
	OldTierPrice    decimal.Decimal     `json:"old_tier_price"`
	NewTierPrice    decimal.Decimal     `json:"new_tier_price"`
	PriceDifference decimal.Decimal     `json:"price_difference"`
}

// TicketUserResponse represents user info in ticket context (minimal PII for admin)
type TicketUserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	FursonaName string    `json:"fursona_name"`
	DenialCount int       `json:"denial_count"`
}

// TicketStatisticsResponse represents ticket statistics for admin dashboard
type TicketStatisticsResponse struct {
	TotalTickets       int64                   `json:"total_tickets"`
	PendingCount       int64                   `json:"pending_count"`
	SelfConfirmedCount int64                   `json:"self_confirmed_count"`
	ApprovedCount      int64                   `json:"approved_count"`
	DeniedCount        int64                   `json:"denied_count"`
	PendingOver24Hours int64                   `json:"pending_over_24_hours"`
	TierStats          []TierStatisticsResponse `json:"tier_stats"`
}

// TierStatisticsResponse represents per-tier statistics
type TierStatisticsResponse struct {
	TierID     uuid.UUID `json:"tier_id"`
	TierCode   string    `json:"tier_code"`
	TierName   string    `json:"tier_name"`
	TotalStock int       `json:"total_stock"`
	Sold       int64     `json:"sold"`
	Available  int       `json:"available"`
}

// BlacklistedUserResponse represents a blacklisted user
type BlacklistedUserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	FursonaName     string     `json:"fursona_name"`
	DenialCount     int        `json:"denial_count"`
	BlacklistedAt   *time.Time `json:"blacklisted_at"`
	BlacklistReason string     `json:"blacklist_reason"`
}
