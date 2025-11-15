package requests

type VerifyOtpRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	Otp   string `json:"otp" binding:"required,len=6" example:"123456"`
}
