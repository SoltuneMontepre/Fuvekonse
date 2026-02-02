package queue

// TicketJobAction is the type of ticket operation to perform.
type TicketJobAction string

const (
	ActionPurchaseTicket   TicketJobAction = "purchase"
	ActionConfirmPayment   TicketJobAction = "confirm_payment"
	ActionCancelTicket     TicketJobAction = "cancel"
	ActionUpdateBadge      TicketJobAction = "update_badge"
	ActionApproveTicket    TicketJobAction = "approve"
	ActionDenyTicket       TicketJobAction = "deny"
	ActionBlacklistUser    TicketJobAction = "blacklist_user"
	ActionUnblacklistUser  TicketJobAction = "unblacklist_user"
)

// TicketJobMessage is the payload sent to SQS for ticket-related work.
type TicketJobMessage struct {
	Action  TicketJobAction `json:"action"`
	UserID  string          `json:"user_id,omitempty"`   // For user actions (purchase, confirm, cancel, update_badge)
	StaffID string          `json:"staff_id,omitempty"`  // For admin actions (approve, deny, blacklist)
	TicketID string         `json:"ticket_id,omitempty"` // For approve/deny
	TargetUserID string     `json:"target_user_id,omitempty"` // For blacklist/unblacklist
	// Request body payloads (JSON-marshalled)
	TierID     string `json:"tier_id,omitempty"`     // For purchase
	Reason     string `json:"reason,omitempty"`      // For deny, blacklist
	ConBadgeName   string `json:"con_badge_name,omitempty"`
	BadgeImage     string `json:"badge_image,omitempty"`
	IsFursuiter    bool   `json:"is_fursuiter"`
	IsFursuitStaff bool   `json:"is_fursuit_staff"`
}
