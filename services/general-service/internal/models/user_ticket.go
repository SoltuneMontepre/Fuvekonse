package models

import (
	"time"

	"github.com/google/uuid"
)

type UserTicket struct {
	Id             uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserId         uuid.UUID   `gorm:"type:uuid;"`
	TicketId       uuid.UUID   `gorm:"type:uuid;"` // Now points to Ticket.Id, not TicketTier.Id
	ConBadgeName   string      `gorm:"type:varchar(255)"`
	BadgeImage     string      `gorm:"type:varchar(500)"` // image url
	IsFursuiter    bool        `gorm:"default:false"`
	IsFursuitStaff bool        `gorm:"default:false"`
	IsCheckedIn    bool        `gorm:"default:false"`
	CreatedAt      time.Time   `gorm:"autoCreateTime"`
	ModifiedAt     time.Time   `gorm:"autoUpdateTime"`
	DeletedAt      *time.Time  `gorm:"index"`
	IsDeleted      bool        `gorm:"default:false"`
	User           *User       `gorm:"foreignKey:UserId"`
	Ticket         interface{} `gorm:"-"` // Ticket info fetched via API from ticket-service
	Payment        Payment     `gorm:"foreignKey:UserTicketId"`
}
