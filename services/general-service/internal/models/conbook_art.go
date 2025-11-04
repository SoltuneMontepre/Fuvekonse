package models

import (
	"time"

	"github.com/google/uuid"
)

type ConBookArt struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserId      uuid.UUID  `gorm:"type:uuid;index"`
	Title       string     `gorm:"type:varchar(255)"`
	Description string     `gorm:"type:varchar(500)"`
	Handle      string     `gorm:"type:varchar(255)"`
	ImageUrl    string     `gorm:"type:varchar(500)"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	ModifiedAt  time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
	IsDeleted   bool       `gorm:"default:false"`
	User        User       `gorm:"foreignKey:UserId"`
}
