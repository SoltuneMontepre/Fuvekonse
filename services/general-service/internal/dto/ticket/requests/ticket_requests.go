package requests

// PurchaseTicketRequest is the request body for purchasing a ticket
type PurchaseTicketRequest struct {
	TierID string `json:"tier_id" binding:"required,uuid"`
}

// ConfirmPaymentRequest is the request body for confirming payment
// No body needed - uses the user's current pending ticket

// DenyTicketRequest is the request body for denying a ticket (admin)
type DenyTicketRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

// UpdateBadgeDetailsRequest is the request body for updating badge details after approval
type UpdateBadgeDetailsRequest struct {
	ConBadgeName   string `json:"con_badge_name" binding:"required,min=1,max=255"`
	BadgeImage     string `json:"badge_image" binding:"omitempty,url,max=500"`
	IsFursuiter    bool   `json:"is_fursuiter"`
	IsFursuitStaff bool   `json:"is_fursuit_staff"`
}

// AdminTicketFilterRequest is the query parameters for admin ticket listing
type AdminTicketFilterRequest struct {
	Status        string `form:"status"`          // pending, self_confirmed, approved, denied
	TierID        string `form:"tier_id"`         // Filter by tier
	Search        string `form:"search"`          // Search reference code, user name, email
	PendingOver24 bool   `form:"pending_over_24"` // Only show > 24h pending
	Page          int    `form:"page,default=1"`
	PageSize      int    `form:"page_size,default=20"`
}

// CreateTicketForAdminRequest is the request body for admin creating a ticket for a user
type CreateTicketForAdminRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	TierID string `json:"tier_id" binding:"required,uuid"`
}

// CreateTicketTierRequest is the request body for admin creating a ticket tier.
// Price and Stock use pointers so that 0 is accepted (validator "required" treats numeric 0 as unset).
type CreateTicketTierRequest struct {
	TicketName  string   `json:"ticket_name" binding:"required,min=1,max=255"`
	Description string   `json:"description" binding:"max=500"`
	Benefits    []string `json:"benefits"`
	Price       *float64 `json:"price" binding:"required,gte=0"`
	Stock       *int     `json:"stock" binding:"required,gte=0"`
	IsActive    bool     `json:"is_active"`
}

// UpdateTicketTierRequest is the request body for admin updating a ticket tier (all optional)
type UpdateTicketTierRequest struct {
	TicketName  *string   `json:"ticket_name" binding:"omitempty,min=1,max=255"`
	Description *string   `json:"description" binding:"omitempty,max=500"`
	Benefits    []string  `json:"benefits"`
	Price       *float64  `json:"price" binding:"omitempty,gte=0"`
	Stock       *int      `json:"stock" binding:"omitempty,gte=0"`
	IsActive    *bool     `json:"is_active"`
}

// UpdateTicketForAdminRequest is the request body for admin updating a ticket (back-door, all fields optional).
type UpdateTicketForAdminRequest struct {
	Status         *string `json:"status" binding:"omitempty,oneof=pending self_confirmed approved denied"`
	TierID         *string `json:"tier_id" binding:"omitempty,uuid"`
	ConBadgeName   *string `json:"con_badge_name" binding:"omitempty,max=255"`
	BadgeImage     *string `json:"badge_image" binding:"omitempty,max=500"`
	IsFursuiter    *bool   `json:"is_fursuiter"`
	IsFursuitStaff *bool   `json:"is_fursuit_staff"`
	IsCheckedIn    *bool   `json:"is_checked_in"`
	DenialReason   *string `json:"denial_reason" binding:"omitempty,max=500"`
}

// UpgradeTicketRequest is the request body for upgrading a ticket to a higher tier
type UpgradeTicketRequest struct {
	NewTierID string `json:"new_tier_id" binding:"required,uuid"`
}

// BlacklistUserRequest is the request body for blacklisting a user
type BlacklistUserRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}
