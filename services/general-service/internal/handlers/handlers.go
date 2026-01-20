package handlers

import "general-service/internal/services"

type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Ticket *TicketHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(services),
		User:   NewUserHandler(services),
		Ticket: NewTicketHandler(services),
	}
}
