package handlers

import "ticket-service/internal/services"

type Handlers struct {
	Role       *RoleHandler
	Permission *PermissionHandler
	UserBan    *UserBanHandler
	Ticket     *TicketHandler
	Payment    *PaymentHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Role:       NewRoleHandler(services.Role),
		Permission: NewPermissionHandler(services.Permission),
		UserBan:    NewUserBanHandler(services.UserBan),
		Ticket:     NewTicketHandler(services.Ticket),
		Payment:    NewPaymentHandler(services.Payment),
	}
}
