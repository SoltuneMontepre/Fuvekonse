package models

import (
	"time"

	"github.com/google/uuid"
)

type PanelStatus string

const (
	PanelStatusPending  PanelStatus = "pending"
	PanelStatusApproved PanelStatus = "approved"
	PanelStatusDenied   PanelStatus = "denied"
)

// PerformancePanel stores a performance / panel application (schedule slot proposal).
type PerformancePanel struct {
	Id                uuid.UUID               `gorm:"type:uuid;primaryKey"`
	UserId            uuid.UUID               `gorm:"type:uuid;index"`
	Title             string                  `gorm:"type:varchar(255)"`
	Nickname          string                  `gorm:"type:varchar(128);not null;default:''"`
	RepresentativeUrl string                  `gorm:"type:varchar(500);column:representative_facebook_url"`
	ParticipantCount  int                     `gorm:"type:int;not null"`
	PerformanceGenre  string                  `gorm:"type:varchar(128)"`
	Introduction      string                  `gorm:"type:text"`
	DurationMinutes   int                     `gorm:"type:int;not null"`
	MaterialsDriveUrl string                  `gorm:"type:varchar(500)"`
	EquipmentNotes    string                  `gorm:"type:varchar(1000)"`
	MembersInfo       []PerformanceMemberInfo `gorm:"serializer:json;type:text"`
	SlotLabel         string                  `gorm:"type:varchar(255)"` // set by staff after approval (room/track + slot name)
	ScheduledStartAt  *time.Time              `gorm:"index"`             // wall-clock start when slotted
	PanelStatus       PanelStatus             `gorm:"type:varchar(20);default:'pending';index"`
	CreatedAt         time.Time               `gorm:"autoCreateTime"`
	ModifiedAt        time.Time               `gorm:"autoUpdateTime"`
	DeletedAt         *time.Time              `gorm:"index"`
	IsDeleted         bool                    `gorm:"default:false"`
	User              User                    `gorm:"foreignKey:UserId"`
}
