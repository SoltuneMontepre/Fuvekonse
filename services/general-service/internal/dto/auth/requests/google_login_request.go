package requests

// GoogleLoginRequest is the body for Google Sign-In. Credential is required.
// When the user is not yet registered, FullName, Nickname, Country, and IdCard are required (same as normal register).
// Password and ConfirmPassword are optional; if provided and matching, the user can also log in with email/password.
type GoogleLoginRequest struct {
	Credential       string `json:"credential" binding:"required"`
	FullName         string `json:"fullName"`
	Nickname         string `json:"nickname"`
	Country          string `json:"country"`
	IdCard           string `json:"idCard"`
	Password         string `json:"password"`
	ConfirmPassword  string `json:"confirmPassword"`
}
