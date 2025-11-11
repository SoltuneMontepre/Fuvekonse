package requests

// UpdateAvatarRequest represents the request to update user avatar
type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required,url,max=500" example:"https://example.com/avatar.jpg"`
}

