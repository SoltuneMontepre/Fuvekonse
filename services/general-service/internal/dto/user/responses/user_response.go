package responses

import (
	"time"

	role "general-service/internal/common/constants"

	"github.com/google/uuid"
)

type UserResponse struct {
	Id               uuid.UUID     `json:"id"`
	FursonaName      string        `json:"fursona_name"`
	LastName         string        `json:"last_name"`
	FirstName        string        `json:"first_name"`
	Country          string        `json:"country"`
	Email            string        `json:"email"`
	Avatar           string        `json:"avatar"`
	Role             role.UserRole `json:"role"`
	IdentificationId string        `json:"identification_id,omitempty"`
	PassportId       string        `json:"passport_id,omitempty"`
	IsVerified       bool          `json:"is_verified"`
	CreatedAt        time.Time     `json:"created_at"`
	ModifiedAt       time.Time     `json:"modified_at"`
}
