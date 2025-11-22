package requests

import "github.com/google/uuid"

type CreateUserTicketRequest struct {
	UserId         uuid.UUID `json:"user_id" binding:"required"`
	TicketId       uuid.UUID `json:"ticket_id" binding:"required"`
	ConBadgeName   string    `json:"con_badge_name"`
	BadgeImage     string    `json:"badge_image"`
	IsFursuiter    bool      `json:"is_fursuiter"`
	IsFursuitStaff bool      `json:"is_fursuit_staff"`
	IsCheckedIn    bool      `json:"is_checked_in"`
}
