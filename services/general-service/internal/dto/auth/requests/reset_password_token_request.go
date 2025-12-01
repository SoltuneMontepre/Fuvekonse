package requests

type ResetPasswordTokenRequest struct {
	Token             string `json:"token" binding:"required" example:"<reset-token>"`
	NewPassword       string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
	ConfirmedPassword string `json:"confirm_password" binding:"required,min=8" example:"newpassword123"`
}
