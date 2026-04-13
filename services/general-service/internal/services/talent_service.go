package services

import (
	"context"
	"errors"
	"general-service/internal/dto/talent/requests"
	"general-service/internal/dto/talent/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"log"

	"github.com/google/uuid"
)

var (
	ErrTalentNotFound       = repositories.ErrTalentNotFound
	ErrTalentNotEditable    = repositories.ErrTalentNotEditable
	ErrUnauthorizedTalent   = repositories.ErrUnauthorizedTalent
	ErrTalentNotSchedulable = repositories.ErrTalentNotSchedulable
)

type TalentService struct {
	repos *repositories.Repositories
}

func NewTalentService(repos *repositories.Repositories) *TalentService {
	return &TalentService{repos: repos}
}

func (s *TalentService) CreateTalent(ctx context.Context, userIDStr string, req *requests.CreateTalentRequest) (*responses.TalentResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	userTicket, err := s.repos.Ticket.GetUserTicket(ctx, userID)
	if err != nil {
		log.Printf("Error checking user ticket for talent: %v", err)
		return nil, errors.New("failed to check user ticket")
	}
	if userTicket == nil {
		return nil, errors.New("user must have a ticket to submit a talent application")
	}
	if userTicket.Status != models.TicketStatusApproved && userTicket.Status != models.TicketStatusAdminGranted {
		return nil, errors.New("user ticket must be approved to submit a talent application")
	}

	talent := &models.PerformanceTalent{
		UserId:            userID,
		Title:             req.Title,
		Nickname:          req.Nickname,
		RepresentativeUrl: req.RepresentativeUrl,
		ParticipantCount:  req.ParticipantCount,
		PerformanceGenre:  req.PerformanceGenre,
		Introduction:      req.Introduction,
		DurationMinutes:   req.DurationMinutes,
		MaterialsDriveUrl: req.MaterialsDriveUrl,
		EquipmentNotes:    req.EquipmentNotes,
		TalentStatus:      models.TalentStatusPending,
	}

	created, err := s.repos.Talent.CreateTalent(ctx, talent)
	if err != nil {
		log.Printf("Error creating talent: %v", err)
		return nil, err
	}

	resp := mappers.MapTalentToResponse(created)
	return &resp, nil
}

func (s *TalentService) GetUserTalents(ctx context.Context, userIDStr string) ([]responses.TalentResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	talents, err := s.repos.Talent.GetUserTalents(ctx, userID)
	if err != nil {
		log.Printf("Error listing talents: %v", err)
		return nil, err
	}

	return mappers.MapTalentsToResponse(talents), nil
}

func (s *TalentService) GetTalentByID(ctx context.Context, talentIDStr string) (*responses.TalentResponse, error) {
	talentID, err := uuid.Parse(talentIDStr)
	if err != nil {
		return nil, errors.New("invalid talent id")
	}

	talent, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapTalentToResponse(talent)
	return &resp, nil
}

func (s *TalentService) EditTalent(ctx context.Context, userIDStr string, talentIDStr string, req *requests.UpdateTalentRequest) (*responses.TalentResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	talentID, err := uuid.Parse(talentIDStr)
	if err != nil {
		return nil, errors.New("invalid talent id")
	}

	canEdit, err := s.repos.Talent.CanEditTalent(ctx, userID, talentID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, ErrUnauthorizedTalent
	}

	existing, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Nickname != nil {
		existing.Nickname = *req.Nickname
	}
	if req.RepresentativeUrl != nil {
		existing.RepresentativeUrl = *req.RepresentativeUrl
	}
	if req.ParticipantCount != nil {
		existing.ParticipantCount = *req.ParticipantCount
	}
	if req.PerformanceGenre != nil {
		existing.PerformanceGenre = *req.PerformanceGenre
	}
	if req.Introduction != nil {
		existing.Introduction = *req.Introduction
	}
	if req.DurationMinutes != nil {
		existing.DurationMinutes = *req.DurationMinutes
	}
	if req.MaterialsDriveUrl != nil {
		existing.MaterialsDriveUrl = *req.MaterialsDriveUrl
	}
	if req.EquipmentNotes != nil {
		existing.EquipmentNotes = *req.EquipmentNotes
	}

	updated, err := s.repos.Talent.UpdateTalent(ctx, talentID, existing)
	if err != nil {
		log.Printf("Error updating talent: %v", err)
		return nil, err
	}

	resp := mappers.MapTalentToResponse(updated)
	return &resp, nil
}

func (s *TalentService) DeleteTalent(ctx context.Context, userIDStr string, talentIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user id")
	}

	talentID, err := uuid.Parse(talentIDStr)
	if err != nil {
		return errors.New("invalid talent id")
	}

	canDelete, err := s.repos.Talent.CanDeleteTalent(ctx, userID, talentID)
	if err != nil {
		return err
	}
	if !canDelete {
		return ErrUnauthorizedTalent
	}

	if err := s.repos.Talent.DeleteTalent(ctx, talentID); err != nil {
		log.Printf("Error deleting talent: %v", err)
		return err
	}
	return nil
}

func (s *TalentService) GetPendingTalents(ctx context.Context) ([]responses.TalentResponse, error) {
	talents, err := s.repos.Talent.GetPendingTalents(ctx)
	if err != nil {
		log.Printf("Error listing pending talents: %v", err)
		return nil, err
	}
	return mappers.MapTalentsToResponse(talents), nil
}

func (s *TalentService) GetApprovedTalents(ctx context.Context) ([]responses.TalentResponse, error) {
	talents, err := s.repos.Talent.GetApprovedTalents(ctx)
	if err != nil {
		log.Printf("Error listing approved talents: %v", err)
		return nil, err
	}
	return mappers.MapTalentsToResponse(talents), nil
}

func (s *TalentService) GetDeniedTalents(ctx context.Context) ([]responses.TalentResponse, error) {
	talents, err := s.repos.Talent.GetDeniedTalents(ctx)
	if err != nil {
		log.Printf("Error listing denied talents: %v", err)
		return nil, err
	}
	return mappers.MapTalentsToResponse(talents), nil
}

func (s *TalentService) setTalentStatus(ctx context.Context, talentIDStr string, status models.TalentStatus) (*responses.TalentResponse, error) {
	talentID, err := uuid.Parse(talentIDStr)
	if err != nil {
		return nil, errors.New("invalid talent id")
	}

	talent, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	if talent.TalentStatus == status {
		return nil, ErrStatusUnchanged
	}

	if err := s.repos.Talent.SetTalentStatus(ctx, talentID, status); err != nil {
		log.Printf("Error setting talent status: %v", err)
		return nil, err
	}

	updated, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapTalentToResponse(updated)
	return &resp, nil
}

func (s *TalentService) ApproveTalent(ctx context.Context, talentIDStr string) (*responses.TalentResponse, error) {
	return s.setTalentStatus(ctx, talentIDStr, models.TalentStatusApproved)
}

func (s *TalentService) DenyTalent(ctx context.Context, talentIDStr string) (*responses.TalentResponse, error) {
	return s.setTalentStatus(ctx, talentIDStr, models.TalentStatusDenied)
}

func (s *TalentService) MarkTalentPending(ctx context.Context, talentIDStr string) (*responses.TalentResponse, error) {
	return s.setTalentStatus(ctx, talentIDStr, models.TalentStatusPending)
}

func (s *TalentService) AssignTalentSchedule(ctx context.Context, talentIDStr string, req *requests.AssignTalentScheduleRequest) (*responses.TalentResponse, error) {
	talentID, err := uuid.Parse(talentIDStr)
	if err != nil {
		return nil, errors.New("invalid talent id")
	}

	talent, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	if talent.TalentStatus != models.TalentStatusApproved {
		return nil, ErrTalentNotSchedulable
	}

	if err := s.repos.Talent.SetTalentSchedule(ctx, talentID, req.SlotLabel, req.ScheduledStartAt); err != nil {
		log.Printf("Error assigning talent schedule: %v", err)
		return nil, err
	}

	updated, err := s.repos.Talent.GetTalentByID(ctx, talentID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapTalentToResponse(updated)
	return &resp, nil
}

