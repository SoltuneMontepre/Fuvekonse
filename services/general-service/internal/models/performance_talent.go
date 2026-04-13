package models

import (
	"time"

	"github.com/google/uuid"
)

type TalentStatus string

const (
	TalentStatusPending  TalentStatus = "pending"
	TalentStatusApproved TalentStatus = "approved"
	TalentStatusDenied   TalentStatus = "denied"
)

// PerformanceTalent stores a talent application (schedule slot proposal).
type PerformanceTalent struct {
	Id                uuid.UUID    `gorm:"type:uuid;primaryKey"`
	UserId            uuid.UUID    `gorm:"type:uuid;index"`
	Title             string       `gorm:"type:varchar(255)"`
	Nickname          string       `gorm:"type:varchar(128);not null;default:''"`
	RepresentativeUrl string       `gorm:"type:varchar(500);column:representative_url"`
	ParticipantCount  int          `gorm:"type:int;not null"`
	PerformanceGenre  string       `gorm:"type:varchar(128)"`
	Introduction      string       `gorm:"type:text"`
	DurationMinutes   int          `gorm:"type:int;not null"`
	MaterialsDriveUrl string       `gorm:"type:varchar(500)"`
	EquipmentNotes    string       `gorm:"type:varchar(1000)"`
	SlotLabel         string       `gorm:"type:varchar(255)"`
	ScheduledStartAt  *time.Time   `gorm:"index"`
	TalentStatus      TalentStatus `gorm:"type:varchar(20);default:'pending';index"`
	CreatedAt         time.Time    `gorm:"autoCreateTime"`
	ModifiedAt        time.Time    `gorm:"autoUpdateTime"`
	DeletedAt         *time.Time   `gorm:"index"`
	IsDeleted         bool         `gorm:"default:false"`
	User              User         `gorm:"foreignKey:UserId"`
}

