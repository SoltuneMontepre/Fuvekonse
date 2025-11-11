package requests

// UpdateProfileRequest represents the request to update user profile
// Using pointers for optional fields allows distinguishing between "not provided" (nil) and "explicitly empty" (*string = "")
type UpdateProfileRequest struct {
	FursonaName      *string `json:"fursona_name,omitempty" binding:"omitempty,max=255" example:"FurryUser"`
	FirstName        *string `json:"first_name,omitempty" binding:"omitempty,max=255" example:"John"`
	LastName         *string `json:"last_name,omitempty" binding:"omitempty,max=255" example:"Doe"`
	Country          *string `json:"country,omitempty" binding:"omitempty,max=255" example:"USA"`
	IdentificationId *string `json:"identification_id,omitempty" binding:"omitempty,max=255" example:"ID123456"`
	PassportId       *string `json:"passport_id,omitempty" binding:"omitempty,max=255" example:"PASS123456"`
}

