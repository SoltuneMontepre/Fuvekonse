package models

import (
	"time"

	"github.com/google/uuid"
)

type DealerBooth struct {
	Id              uuid.UUID         `gorm:"type:uuid;primaryKey"`
	BoothName       string            `gorm:"type:varchar(255)"`
	Description     string            `gorm:"type:varchar(500)"`
	BoothNumber     string            `gorm:"type:varchar(100)"`
	PriceSheet      string            `gorm:"type:varchar(500)"` // image url
	IsVerified      bool              `gorm:"default:false"`
	PaymentVerified bool              `gorm:"default:false"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"`
	ModifiedAt      time.Time         `gorm:"autoUpdateTime"`
	DeletedAt       *time.Time        `gorm:"index"`
	IsDeleted       bool              `gorm:"default:false"`
	Staffs          []UserDealerStaff `gorm:"foreignKey:BoothId"`
}
