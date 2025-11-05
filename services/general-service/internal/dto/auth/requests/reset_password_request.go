package requests

type ResetPasswordRequest struct {
	NewPassword       string `json:"new_password" binding:"required,min=8" example:"password123"`
	ConfirmedPassword string `json:"confirm_password" binding:"required,min=8" example:"password123"`
}
