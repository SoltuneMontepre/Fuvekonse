package services

import "ticket-service/internal/repositories"

type Services struct {
	Role       RoleServiceInterface
	Permission PermissionServiceInterface
	UserBan    UserBanServiceInterface
	Ticket     TicketServiceInterface
	Payment    PaymentServiceInterface
}

type PaymentServiceConfig struct {
	PayOSClientID    string
	PayOSAPIKey      string
	PayOSChecksumKey string
	FrontendURL      string
	GeneralSvcURL    string
}

func NewServices(repos *repositories.Repositories, paymentConfig PaymentServiceConfig) *Services {
	return &Services{
		Role:       NewRoleService(repos.Role, repos.Permission),
		Permission: NewPermissionService(repos.Permission),
		UserBan:    NewUserBanService(repos.UserBan, repos.Permission),
		Ticket:     NewTicketService(repos.TicketTier, repos.Ticket),
		Payment:    NewPaymentService(repos.Payment, repos.Ticket, repos.TicketTier, paymentConfig.PayOSClientID, paymentConfig.PayOSAPIKey, paymentConfig.PayOSChecksumKey, paymentConfig.FrontendURL, paymentConfig.GeneralSvcURL),
	}
}
