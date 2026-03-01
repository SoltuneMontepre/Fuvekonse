package responses

import (
	"time"

	"github.com/google/uuid"
)

// ConbookResponse is the response for a single conbook
type ConbookResponse struct {
	Id          uuid.UUID `json:"id"`
	UserId      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Handle      string    `json:"handle"`
	ImageUrl    string    `json:"image_url"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

// ConbookListResponse is the response for a list of conbooks
type ConbookListResponse struct {
	Conbooks []ConbookResponse `json:"conbooks"`
	Count    int64             `json:"count"`
}
