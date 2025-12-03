package models

import (
	"time"

	role "general-service/internal/common/constants"

	"github.com/google/uuid"
)

type User struct {
	Id            uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	FursonaName   string        `gorm:"type:varchar(255)" json:"fursona_name"`
	LastName      string        `gorm:"type:varchar(255)" json:"last_name"`
	FirstName     string        `gorm:"type:varchar(255)" json:"first_name"`
	Password      string        `gorm:"type:varchar(255)" json:"-"`
	Country       string        `gorm:"type:varchar(255)" json:"country"`
	Email         string        `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Avatar        string        `gorm:"type:varchar(500)" json:"avatar"` // image url
	Role          role.UserRole `gorm:"type:integer;default:0" json:"role"`
	IdCard        string        `gorm:"type:varchar(255)" json:"id_card"`
	IsVerified    bool          `gorm:"default:false" json:"is_verified"`
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	ModifiedAt    time.Time     `gorm:"autoUpdateTime" json:"modified_at"`
	DeletedAt     *time.Time    `gorm:"index" json:"deleted_at,omitempty"`
	IsDeleted     bool          `gorm:"default:false" json:"is_deleted"`
	Otp           string        `gorm:"type:varchar(6)" json:"otp"`
	OtpExpiryTime *time.Time    `gorm:"index" json:"otp_expiry_time"`
}
