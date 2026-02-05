package requests

// UpdateAvatarRequest represents the request to update user avatar.
// Use empty string to clear the avatar.
type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
}

