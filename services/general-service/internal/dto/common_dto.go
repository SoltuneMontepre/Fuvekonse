package dto

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Message string `json:"message" example:"pong"`
	Status  string `json:"status"  example:"healthy"`
}
