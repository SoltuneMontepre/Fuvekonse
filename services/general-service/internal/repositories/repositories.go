package repositories

import "gorm.io/gorm"

type Repositories struct {
	User       *UserRepository
	UserTicket *UserTicketRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:       NewUserRepository(db),
		UserTicket: NewUserTicketRepository(db),
	}
}
