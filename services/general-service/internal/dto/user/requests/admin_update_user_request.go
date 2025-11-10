package requests

import role "general-service/internal/common/constants"

// AdminUpdateUserRequest represents the request to update user information by admin
// Using pointers for optional fields allows distinguishing between "not provided" (nil) and "explicitly empty" (*string = "")
type AdminUpdateUserRequest struct {
	FursonaName      *string       `json:"fursona_name,omitempty" binding:"omitempty,max=255" example:"FurryUser"`
	FirstName        *string       `json:"first_name,omitempty" binding:"omitempty,max=255" example:"John"`
	LastName         *string       `json:"last_name,omitempty" binding:"omitempty,max=255" example:"Doe"`
	Country          *string       `json:"country,omitempty" binding:"omitempty,max=255" example:"USA"`
	Email            *string       `json:"email,omitempty" binding:"omitempty,email" example:"user@example.com"`
	Avatar           *string       `json:"avatar,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
	Role             *role.UserRole `json:"role,omitempty" binding:"omitempty" example:"user"`
	IdentificationId *string       `json:"identification_id,omitempty" binding:"omitempty,max=255" example:"ID123456"`
	PassportId       *string       `json:"passport_id,omitempty" binding:"omitempty,max=255" example:"PASS123456"`
	IsVerified       *bool         `json:"is_verified,omitempty" binding:"omitempty" example:"true"`
}

