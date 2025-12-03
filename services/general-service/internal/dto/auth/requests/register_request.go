package requests

// RegisterRequest represents the user registration payload
type RegisterRequest struct {
	FullName        string `json:"fullName" binding:"required" example:"John Doe"`
	Nickname        string `json:"nickname" binding:"required" example:"FurryFox"`
	Email           string `json:"email" binding:"required,email" example:"user@example.com"`
	Country         string `json:"country" binding:"required" example:"United States"`
	IdCard          string `json:"idCard" binding:"required" example:"1234567890"`
	Password        string `json:"password" binding:"required,min=6" example:"SecurePass123"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=6" example:"SecurePass123"`
}
