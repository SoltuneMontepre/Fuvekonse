package responses

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users []UserDetailedResponse `json:"users"`
}

