package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TicketTier struct {
	Id          uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2)"`
	Description string          `gorm:"type:varchar(500)"`
	TicketName  string          `gorm:"type:varchar(255)"`
	Stock       int             `gorm:"type:int;check:stock >= 0"`
	IsActive    bool            `gorm:"default:true"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	ModifiedAt  time.Time       `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time      `gorm:"index"`
	IsDeleted   bool            `gorm:"default:false"`
	UserTickets []UserTicket    `gorm:"foreignKey:TicketId"`
}
