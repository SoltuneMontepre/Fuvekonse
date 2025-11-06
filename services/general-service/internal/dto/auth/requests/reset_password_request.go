package requests

type ResetPasswordRequest struct {
	// Current password
    CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`

    // New password
    NewPassword string `json:"new_password" binding:"required,min=8" example:"newpassword123"`

    // Confirm new password
    ConfirmedPassword string `json:"confirm_password" binding:"required,min=8" example:"newpassword123"`
}
