package mappers

import (
	"general-service/internal/dto/panel/responses"
	"general-service/internal/models"
	"time"
)

func MapPanelToResponse(panel *models.PerformancePanel) responses.PanelResponse {
	var scheduledEnd *time.Time
	if panel.ScheduledStartAt != nil && panel.DurationMinutes > 0 {
		t := panel.ScheduledStartAt.Add(time.Duration(panel.DurationMinutes) * time.Minute)
		scheduledEnd = &t
	}

	return responses.PanelResponse{
		Id:                        panel.Id,
		UserId:                    panel.UserId,
		Title:                     panel.Title,
		Nickname:                  panel.Nickname,
		RepresentativeUrl: panel.RepresentativeUrl,
		ParticipantCount:          panel.ParticipantCount,
		PerformanceGenre:          panel.PerformanceGenre,
		Introduction:              panel.Introduction,
		DurationMinutes:           panel.DurationMinutes,
		MaterialsDriveUrl:         panel.MaterialsDriveUrl,
		EquipmentNotes:            panel.EquipmentNotes,
		SlotLabel:                 panel.SlotLabel,
		ScheduledStartAt:          panel.ScheduledStartAt,
		ScheduledEndAt:            scheduledEnd,
		Status:                    panel.PanelStatus,
		CreatedAt:                 panel.CreatedAt,
		ModifiedAt:                panel.ModifiedAt,
	}
}

func MapPanelsToResponse(panels []models.PerformancePanel) []responses.PanelResponse {
	out := make([]responses.PanelResponse, len(panels))
	for i := range panels {
		out[i] = MapPanelToResponse(&panels[i])
	}
	return out
}
