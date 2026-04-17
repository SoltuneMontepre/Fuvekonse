package responses

import (
	"general-service/internal/models"
	"time"

	"github.com/google/uuid"
)

// TalentMemberJSON is one participant in API responses.
type TalentMemberJSON struct {
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

// TalentResponse is one talent application.
type TalentResponse struct {
	Id                uuid.UUID           `json:"id"`
	UserId            uuid.UUID           `json:"user_id"`
	Title             string              `json:"title"`
	Nickname          string              `json:"nickname"`
	RepresentativeUrl string              `json:"representative_url"`
	ParticipantCount  int                 `json:"participant_count"`
	PerformanceGenre  string              `json:"performance_genre"`
	Introduction      string              `json:"introduction"`
	DurationMinutes   int                 `json:"duration_minutes"`
	MaterialsDriveUrl string              `json:"materials_drive_url"`
	EquipmentNotes    string              `json:"equipment_notes"`
	Members           []TalentMemberJSON  `json:"members,omitempty"`
	SlotLabel         string              `json:"slot_label,omitempty"`
	ScheduledStartAt  *time.Time          `json:"scheduled_start_at,omitempty"`
	ScheduledEndAt    *time.Time          `json:"scheduled_end_at,omitempty"`
	Status            models.TalentStatus `json:"status"`
	CreatedAt         time.Time           `json:"created_at"`
	ModifiedAt        time.Time           `json:"modified_at"`
}
