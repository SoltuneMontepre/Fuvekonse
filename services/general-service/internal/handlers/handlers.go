package handlers

import "general-service/internal/services"

type Handlers struct {
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{}
}
