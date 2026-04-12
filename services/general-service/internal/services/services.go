package services

import (
	"general-service/internal/repositories"

	"github.com/redis/go-redis/v9"
)

type Services struct {
	Auth      *AuthService
	User      *UserService
	Mail      *MailService
	Ticket    *TicketService
	Dealer    *DealerService
	Conbook   *ConbookService
	Panel     *PanelService
	Analytics *AnalyticsService
}

func NewServices(repos *repositories.Repositories, redisClient *redis.Client, loginMaxFail int, loginFailBlockMinutes int) *Services {
	mail := NewMailService(repos)
	ticket := NewTicketService(repos, mail)
	return &Services{
		Auth:      NewAuthService(repos, redisClient, loginMaxFail, loginFailBlockMinutes),
		User:      NewUserService(repos),
		Mail:      mail,
		Ticket:    ticket,
		Dealer:    NewDealerService(repos, mail),
		Conbook:   NewConbookService(repos),
		Panel:     NewPanelService(repos),
		Analytics: NewAnalyticsService(repos, ticket),
	}
}
