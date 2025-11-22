package responses

import (
	"time"

	"github.com/google/uuid"
)

type UserTicketResponse struct {
	Id             uuid.UUID `json:"id"`
	UserId         uuid.UUID `json:"user_id"`
	TicketId       uuid.UUID `json:"ticket_id"`
	ConBadgeName   string    `json:"con_badge_name"`
	BadgeImage     string    `json:"badge_image"`
	IsFursuiter    bool      `json:"is_fursuiter"`
	IsFursuitStaff bool      `json:"is_fursuit_staff"`
	IsCheckedIn    bool      `json:"is_checked_in"`
	CreatedAt      time.Time `json:"created_at"`
}
