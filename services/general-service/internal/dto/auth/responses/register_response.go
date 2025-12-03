package responses

type RegisterResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}
