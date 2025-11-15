package requests

type ResendOtpRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}
