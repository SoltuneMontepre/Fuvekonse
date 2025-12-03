package requests

import role "general-service/internal/common/constants"

// AdminUpdateUserRequest represents the request to update user information by admin
// Using pointers for optional fields allows distinguishing between "not provided" (nil) and "explicitly empty" (*string = "")
// Note: Email cannot be changed by admin for security reasons
type AdminUpdateUserRequest struct {
	FursonaName *string        `json:"fursona_name,omitempty" binding:"omitempty,max=255" example:"FurryUser"`
	FirstName   *string        `json:"first_name,omitempty" binding:"omitempty,max=255" example:"John"`
	LastName    *string        `json:"last_name,omitempty" binding:"omitempty,max=255" example:"Doe"`
	Country     *string        `json:"country,omitempty" binding:"omitempty,max=255" example:"USA"`
	Avatar      *string        `json:"avatar,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
	Role        *role.UserRole `json:"role,omitempty" binding:"omitempty" example:"user"`
	IdCard      *string        `json:"id_card,omitempty" binding:"omitempty,max=255" example:"ID123456"`
	IsVerified  *bool          `json:"is_verified,omitempty" binding:"omitempty" example:"true"`
}
