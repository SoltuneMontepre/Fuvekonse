package models

import (
	"time"

	"github.com/google/uuid"
)

type UserTicket struct {
	Id             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId         uuid.UUID `gorm:"type:uuid;index"`
	TicketId       uuid.UUID `gorm:"type:uuid;index"`
	ConBadgeName   string    `gorm:"type:varchar(255)"`
	BadgeImage     string    `gorm:"type:varchar(500)"` // image url
	IsFursuiter    bool      `gorm:"default:false"`
	IsFursuitStaff bool      `gorm:"default:false"`

	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	ModifiedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
	IsDeleted  bool       `gorm:"default:false"`

	User    User       `gorm:"foreignKey:UserId"`
	Ticket  TicketTier `gorm:"foreignKey:TicketId"`
	Payment Payment    `gorm:"foreignKey:UserTicketId"`
}
