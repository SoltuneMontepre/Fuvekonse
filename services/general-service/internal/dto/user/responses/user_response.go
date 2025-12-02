package responses

import (
	"time"

	role "general-service/internal/common/constants"

	"github.com/google/uuid"
)

// UserResponse is the public user response DTO without sensitive PII
// This should be used for public APIs where user data is exposed
type UserResponse struct {
	Id          uuid.UUID     `json:"id"`
	FursonaName string        `json:"fursona_name"`
	LastName    string        `json:"last_name"`
	FirstName   string        `json:"first_name"`
	Country     string        `json:"country"`
	Avatar      string        `json:"avatar"`
	Role        role.UserRole `json:"role"`
	IsVerified  bool          `json:"is_verified"`
	CreatedAt   time.Time     `json:"created_at"`
	ModifiedAt  time.Time     `json:"modified_at"`
}

// UserDetailedResponse includes sensitive PII fields
// This should be used only for restricted/internal endpoints where the user
// is accessing their own data or admins are accessing user details
type UserDetailedResponse struct {
	Id          uuid.UUID     `json:"id"`
	FursonaName string        `json:"fursona_name"`
	LastName    string        `json:"last_name"`
	FirstName   string        `json:"first_name"`
	Country     string        `json:"country"`
	Email       string        `json:"email"`
	Avatar      string        `json:"avatar"`
	Role        role.UserRole `json:"role"`
	IdCard      string        `json:"id_card,omitempty"`
	IsVerified  bool          `json:"is_verified"`
	CreatedAt   time.Time     `json:"created_at"`
	ModifiedAt  time.Time     `json:"modified_at"`
}
