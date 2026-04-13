package handlers

import (
	"general-service/internal/queue"
	"general-service/internal/services"
)

type Handlers struct {
	Auth      *AuthHandler
	User      *UserHandler
	Ticket    *TicketHandler
	Dealer    *DealerHandler
	Conbook   *ConbookHandler
	Panel     *PanelHandler
	Talent    *TalentHandler
	Analytics *AnalyticsHandler
	DevMail   *DevMailHandler
}

func NewHandlers(services *services.Services, queuePublisher queue.Publisher) *Handlers {
	return &Handlers{
		Auth:      NewAuthHandler(services),
		User:      NewUserHandler(services),
		Ticket:    NewTicketHandler(services, queuePublisher),
		Dealer:    NewDealerHandler(services),
		Conbook:   NewConbookHandler(services),
		Panel:     NewPanelHandler(services),
		Talent:    NewTalentHandler(services),
		Analytics: NewAnalyticsHandler(services),
		DevMail:   NewDevMailHandler(services),
	}
}
