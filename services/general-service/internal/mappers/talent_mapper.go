package mappers

import (
	"general-service/internal/dto/talent/responses"
	"general-service/internal/models"
	"time"
)

func MapTalentToResponse(talent *models.PerformanceTalent) responses.TalentResponse {
	var scheduledEnd *time.Time
	if talent.ScheduledStartAt != nil && talent.DurationMinutes > 0 {
		t := talent.ScheduledStartAt.Add(time.Duration(talent.DurationMinutes) * time.Minute)
		scheduledEnd = &t
	}

	members := make([]responses.TalentMemberJSON, 0, len(talent.MembersInfo))
	for i := range talent.MembersInfo {
		members = append(members, responses.TalentMemberJSON{
			Name:   talent.MembersInfo[i].Name,
			Detail: talent.MembersInfo[i].Detail,
		})
	}

	return responses.TalentResponse{
		Id:                talent.Id,
		UserId:            talent.UserId,
		Title:             talent.Title,
		Nickname:          talent.Nickname,
		RepresentativeUrl: talent.RepresentativeUrl,
		ParticipantCount:  talent.ParticipantCount,
		PerformanceGenre:  talent.PerformanceGenre,
		Introduction:      talent.Introduction,
		DurationMinutes:   talent.DurationMinutes,
		MaterialsDriveUrl: talent.MaterialsDriveUrl,
		EquipmentNotes:    talent.EquipmentNotes,
		Members:           members,
		SlotLabel:         talent.SlotLabel,
		ScheduledStartAt:  talent.ScheduledStartAt,
		ScheduledEndAt:    scheduledEnd,
		Status:            talent.TalentStatus,
		CreatedAt:         talent.CreatedAt,
		ModifiedAt:        talent.ModifiedAt,
	}
}

func MapTalentsToResponse(talents []models.PerformanceTalent) []responses.TalentResponse {
	out := make([]responses.TalentResponse, len(talents))
	for i := range talents {
		out[i] = MapTalentToResponse(&talents[i])
	}
	return out
}
