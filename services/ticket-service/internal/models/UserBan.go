package models

import "time"

type UserBan struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	PermID    uint      `gorm:"not null;index" json:"perm_id"`
	Reason    string    `gorm:"not null" json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Permission Permission `gorm:"foreignKey:PermID;references:PermID"`
}