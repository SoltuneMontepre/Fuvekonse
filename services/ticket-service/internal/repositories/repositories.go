package repositories

import "gorm.io/gorm"

type Repositories struct {
	Role       RoleRepositoryInterface
	Permission PermissionRepositoryInterface
	UserBan    UserBanRepositoryInterface
	TicketTier TicketTierRepositoryInterface
	Ticket     TicketRepositoryInterface
	Payment    PaymentRepositoryInterface
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Role:       NewRoleRepository(db),
		Permission: NewPermissionRepository(db),
		UserBan:    NewUserBanRepository(db),
		TicketTier: NewTicketTierRepository(db),
		Ticket:     NewTicketRepository(db),
		Payment:    NewPaymentRepository(db),
	}
}
