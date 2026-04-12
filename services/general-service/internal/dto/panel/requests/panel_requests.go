package requests

import "time"

// CreatePanelRequest is the body for submitting a performance / panel application.
type CreatePanelRequest struct {
	Title                     string `json:"title" binding:"required,min=1,max=255"`
	Nickname                  string `json:"nickname" binding:"required,min=1,max=128"`
	RepresentativeUrl string `json:"representative_url" binding:"required,url,max=500"`
	ParticipantCount          int    `json:"participant_count" binding:"required,min=1,max=500"`
	PerformanceGenre          string `json:"performance_genre" binding:"required,min=1,max=128"`
	Introduction              string `json:"introduction" binding:"required,min=1,max=4000"`
	DurationMinutes           int    `json:"duration_minutes" binding:"required,min=5,max=480"`
	MaterialsDriveUrl         string `json:"materials_drive_url" binding:"required,url,max=500"`
	EquipmentNotes            string `json:"equipment_notes" binding:"max=1000"`
}

// UpdatePanelRequest updates a pending application (same fields as create; use pointers for partial updates).
type UpdatePanelRequest struct {
	Title                     *string `json:"title" binding:"omitempty,min=1,max=255"`
	Nickname                  *string `json:"nickname" binding:"omitempty,min=1,max=128"`
	RepresentativeUrl *string `json:"representative_url" binding:"omitempty,url,max=500"`
	ParticipantCount          *int    `json:"participant_count" binding:"omitempty,min=1,max=500"`
	PerformanceGenre          *string `json:"performance_genre" binding:"omitempty,min=1,max=128"`
	Introduction              *string `json:"introduction" binding:"omitempty,min=1,max=4000"`
	DurationMinutes           *int    `json:"duration_minutes" binding:"omitempty,min=5,max=480"`
	MaterialsDriveUrl         *string `json:"materials_drive_url" binding:"omitempty,url,max=500"`
	EquipmentNotes            *string `json:"equipment_notes" binding:"omitempty,max=1000"`
}

// AssignPanelScheduleRequest is set by admin/staff after the panel is approved.
type AssignPanelScheduleRequest struct {
	SlotLabel        string    `json:"slot_label" binding:"required,min=1,max=255"`
	ScheduledStartAt time.Time `json:"scheduled_start_at" binding:"required"`
}
