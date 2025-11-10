package requests

import role "general-service/internal/common/constants"

// AdminUpdateUserRequest represents the request to update user information by admin
type AdminUpdateUserRequest struct {
	FursonaName      string        `json:"fursona_name" binding:"omitempty,max=255" example:"FurryUser"`
	FirstName        string        `json:"first_name" binding:"omitempty,max=255" example:"John"`
	LastName         string        `json:"last_name" binding:"omitempty,max=255" example:"Doe"`
	Country          string        `json:"country" binding:"omitempty,max=255" example:"USA"`
	Email            string        `json:"email" binding:"omitempty,email" example:"user@example.com"`
	Avatar           string        `json:"avatar" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
	Role             *role.UserRole `json:"role" binding:"omitempty" example:"user"`
	IdentificationId string        `json:"identification_id" binding:"omitempty,max=255" example:"ID123456"`
	PassportId       string        `json:"passport_id" binding:"omitempty,max=255" example:"PASS123456"`
	IsVerified       *bool         `json:"is_verified" binding:"omitempty" example:"true"`
}

