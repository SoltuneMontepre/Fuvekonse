package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDealerStaff struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId  uuid.UUID `gorm:"type:uuid;index"`
	BoothId uuid.UUID `gorm:"type:uuid;index"`
	IsOwner bool      `gorm:"default:false"`

	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	ModifiedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
	IsDeleted  bool       `gorm:"default:false"`

	// Relations
	User  User        `gorm:"foreignKey:UserId"`
	Booth DealerBooth `gorm:"foreignKey:BoothId"`
}
