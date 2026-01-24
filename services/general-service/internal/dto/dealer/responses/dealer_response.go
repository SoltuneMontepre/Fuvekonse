package responses

import (
	"time"

	"github.com/google/uuid"
)

type DealerBoothResponse struct {
	Id          uuid.UUID `json:"id"`
	BoothName   string    `json:"booth_name"`
	Description string    `json:"description"`
	BoothNumber string    `json:"booth_number"`
	PriceSheet  string    `json:"price_sheet"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}
