package responses

import (
	"general-service/internal/models"
	"time"

	"github.com/google/uuid"
)

// PanelMemberJSON is one participant in API responses.
type PanelMemberJSON struct {
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

// PanelResponse is one performance / panel application.
type PanelResponse struct {
	Id                uuid.UUID          `json:"id"`
	UserId            uuid.UUID          `json:"user_id"`
	Title             string             `json:"title"`
	Nickname          string             `json:"nickname"`
	RepresentativeUrl string             `json:"representative_url"`
	ParticipantCount  int                `json:"participant_count"`
	PerformanceGenre  string             `json:"performance_genre"`
	Introduction      string             `json:"introduction"`
	DurationMinutes   int                `json:"duration_minutes"`
	MaterialsDriveUrl string             `json:"materials_drive_url"`
	EquipmentNotes    string             `json:"equipment_notes"`
	Members           []PanelMemberJSON  `json:"members,omitempty"`
	SlotLabel         string             `json:"slot_label,omitempty"`
	ScheduledStartAt  *time.Time         `json:"scheduled_start_at,omitempty"`
	ScheduledEndAt    *time.Time         `json:"scheduled_end_at,omitempty"` // start + duration when start is set
	Status            models.PanelStatus `json:"status"`
	CreatedAt         time.Time          `json:"created_at"`
	ModifiedAt        time.Time          `json:"modified_at"`
}
