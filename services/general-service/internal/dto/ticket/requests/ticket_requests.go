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

// BlacklistUserRequest is the request body for blacklisting a user
type BlacklistUserRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}
