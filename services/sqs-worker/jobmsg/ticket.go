package jobmsg

// Action is the type of ticket operation (must match general-service queue payload).
type Action string

const (
	ActionPurchaseTicket  Action = "purchase"
	ActionConfirmPayment  Action = "confirm_payment"
	ActionCancelTicket    Action = "cancel"
	ActionUpdateBadge     Action = "update_badge"
	ActionApproveTicket   Action = "approve"
	ActionDenyTicket      Action = "deny"
	ActionUpgradeTicket   Action = "upgrade_ticket"
	ActionBlacklistUser   Action = "blacklist_user"
	ActionUnblacklistUser Action = "unblacklist_user"
)

// TicketJobMessage is the SQS body (same shape as general-service queue.TicketJobMessage).
type TicketJobMessage struct {
	Action        Action `json:"action"`
	UserID        string `json:"user_id,omitempty"`
	StaffID       string `json:"staff_id,omitempty"`
	TicketID      string `json:"ticket_id,omitempty"`
	TargetUserID  string `json:"target_user_id,omitempty"`
	TierID        string `json:"tier_id,omitempty"`
	Reason        string `json:"reason,omitempty"`
	ConBadgeName  string `json:"con_badge_name,omitempty"`
	BadgeImage    string `json:"badge_image,omitempty"`
	IsFursuiter   bool   `json:"is_fursuiter"`
	IsFursuitStaff bool  `json:"is_fursuit_staff"`
}
