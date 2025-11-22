package dto

import "github.com/google/uuid"

type TicketTierResponse struct {
	ID          uuid.UUID `json:"id"`
	TicketName  string    `json:"ticket_name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Stock       int       `json:"stock"`
	IsActive    bool      `json:"is_active"`
	BannerImage string    `json:"banner_image,omitempty"`
}

type TicketResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	TicketTierID uuid.UUID `json:"ticket_tier_id"`
	OrderCode    int64     `json:"order_code"`
	Status       string    `json:"status"`
	Price        int64     `json:"price"`
}
