package repositories

import "gorm.io/gorm"

type Repositories struct {
	User   *UserRepository
	Ticket *TicketRepository
	Dealer *DealerRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:   NewUserRepository(db),
		Ticket: NewTicketRepository(db),
		Dealer: NewDealerRepository(db),
	}
}
