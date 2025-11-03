package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Payment struct {
	Id                   uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserTicketId         uuid.UUID       `gorm:"type:uuid;index"`
	PaymentMethod        string          `gorm:"type:varchar(50)"`
	Status               string          `gorm:"type:varchar(50)"`
	InvoicedPrice        decimal.Decimal `gorm:"type:decimal(10,2)"`
	Currency             string          `gorm:"type:varchar(10)"`
	GatewayTransactionId string          `gorm:"type:varchar(255)"`
	Provider             string          `gorm:"type:varchar(255)"`
	RawResponse          string          `gorm:"type:varchar(1000)"`
	UserTicket           *UserTicket     `gorm:"foreignKey:UserTicketId"`
}
