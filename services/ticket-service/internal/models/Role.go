package models

type Role struct {
  RoleID      uint           `gorm:"primaryKey;autoIncrement" json:"role_id"`
  Name        string         `gorm:"size:100;not null;unique" json:"name"`
	Permissions []Permission   `gorm:"many2many:role_permission;" json:"permissions,omitempty"`
}