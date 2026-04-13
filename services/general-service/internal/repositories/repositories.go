package repositories

import "gorm.io/gorm"

type Repositories struct {
	User    *UserRepository
	Ticket  *TicketRepository
	Dealer  *DealerRepository
	Conbook *ConbookRepository
	Panel   *PanelRepository
	Talent  *TalentRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:    NewUserRepository(db),
		Ticket:  NewTicketRepository(db),
		Dealer:  NewDealerRepository(db),
		Conbook: NewConbookRepository(db),
		Panel:   NewPanelRepository(db),
		Talent:  NewTalentRepository(db),
	}
}
