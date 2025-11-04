package handlers

import "general-service/internal/services"

type Handlers struct {
	Auth *AuthHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Auth: NewAuthHandler(services),
	}
}
