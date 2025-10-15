package models

type Permission struct {
	PermID uint `gorm:"primaryKey;autoIncrement" json:"perm_id"`
	Name   string `gorm:"size:100;not null;unique" json:"name"`
	Roles  []Role `gorm:"many2many:role_permission;" json:"roles,omitempty"`
}