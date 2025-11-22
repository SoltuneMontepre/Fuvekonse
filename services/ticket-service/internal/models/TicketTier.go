package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTier struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TicketName  string         `gorm:"size:255;not null" json:"ticket_name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       int64          `gorm:"not null" json:"price"` // Price in VNĐ (smallest unit)
	Stock       int            `gorm:"default:0" json:"stock"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	BannerImage string         `gorm:"size:500" json:"banner_image,omitempty"`
	Tickets     []Ticket       `gorm:"foreignKey:TicketTierID" json:"tickets,omitempty"` // Tickets belonging to this tier
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
