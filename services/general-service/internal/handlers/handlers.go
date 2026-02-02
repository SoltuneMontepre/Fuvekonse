package handlers

import (
	"general-service/internal/queue"
	"general-service/internal/services"
)

type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Ticket *TicketHandler
	Dealer *DealerHandler
}

func NewHandlers(services *services.Services, queuePublisher queue.Publisher) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(services),
		User:   NewUserHandler(services),
		Ticket: NewTicketHandler(services, queuePublisher),
		Dealer: NewDealerHandler(services),
	}
}
