package responses

import (
	"time"

	"github.com/google/uuid"
)

type DealerStaffResponse struct {
	Id         uuid.UUID `json:"id"`
	UserId     uuid.UUID `json:"user_id"`
	UserEmail  string    `json:"user_email"`
	UserName   string    `json:"user_name"`
	IsOwner    bool      `json:"is_owner"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type DealerBoothDetailResponse struct {
	Id          uuid.UUID              `json:"id"`
	BoothName   string                 `json:"booth_name"`
	Description string                 `json:"description"`
	BoothNumber string                 `json:"booth_number"`
	PriceSheet  string                 `json:"price_sheet"`
	IsVerified  bool                   `json:"is_verified"`
	CreatedAt   time.Time              `json:"created_at"`
	ModifiedAt  time.Time              `json:"modified_at"`
	Staffs      []*DealerStaffResponse `json:"staffs,omitempty"`
}
