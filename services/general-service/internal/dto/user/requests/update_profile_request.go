package requests

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	FursonaName      string `json:"fursona_name" binding:"omitempty,max=255" example:"FurryUser"`
	FirstName        string `json:"first_name" binding:"omitempty,max=255" example:"John"`
	LastName         string `json:"last_name" binding:"omitempty,max=255" example:"Doe"`
	Country          string `json:"country" binding:"omitempty,max=255" example:"USA"`
	IdentificationId string `json:"identification_id" binding:"omitempty,max=255" example:"ID123456"`
	PassportId       string `json:"passport_id" binding:"omitempty,max=255" example:"PASS123456"`
}

