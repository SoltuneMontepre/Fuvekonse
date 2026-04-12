package services

import (
	"context"
	"errors"
	"general-service/internal/dto/panel/requests"
	"general-service/internal/dto/panel/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"log"

	"github.com/google/uuid"
)

var (
	ErrPanelNotFound       = repositories.ErrPanelNotFound
	ErrPanelNotEditable    = repositories.ErrPanelNotEditable
	ErrUnauthorizedPanel   = repositories.ErrUnauthorizedPanel
	ErrPanelNotSchedulable = repositories.ErrPanelNotSchedulable
)

type PanelService struct {
	repos *repositories.Repositories
}

func NewPanelService(repos *repositories.Repositories) *PanelService {
	return &PanelService{repos: repos}
}

func (s *PanelService) CreatePanel(ctx context.Context, userIDStr string, req *requests.CreatePanelRequest) (*responses.PanelResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	userTicket, err := s.repos.Ticket.GetUserTicket(ctx, userID)
	if err != nil {
		log.Printf("Error checking user ticket for panel: %v", err)
		return nil, errors.New("failed to check user ticket")
	}
	if userTicket == nil {
		return nil, errors.New("user must have a ticket to submit a panel application")
	}
	if userTicket.Status != models.TicketStatusApproved && userTicket.Status != models.TicketStatusAdminGranted {
		return nil, errors.New("user ticket must be approved to submit a panel application")
	}

	panel := &models.PerformancePanel{
		UserId:                    userID,
		Title:                     req.Title,
		Nickname:                  req.Nickname,
		RepresentativeUrl: req.RepresentativeUrl,
		ParticipantCount:          req.ParticipantCount,
		PerformanceGenre:          req.PerformanceGenre,
		Introduction:              req.Introduction,
		DurationMinutes:           req.DurationMinutes,
		MaterialsDriveUrl:         req.MaterialsDriveUrl,
		EquipmentNotes:            req.EquipmentNotes,
		PanelStatus:               models.PanelStatusPending,
	}

	created, err := s.repos.Panel.CreatePanel(ctx, panel)
	if err != nil {
		log.Printf("Error creating panel: %v", err)
		return nil, err
	}

	resp := mappers.MapPanelToResponse(created)
	return &resp, nil
}

func (s *PanelService) GetUserPanels(ctx context.Context, userIDStr string) ([]responses.PanelResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	panels, err := s.repos.Panel.GetUserPanels(ctx, userID)
	if err != nil {
		log.Printf("Error listing panels: %v", err)
		return nil, err
	}

	return mappers.MapPanelsToResponse(panels), nil
}

func (s *PanelService) GetPanelByID(ctx context.Context, panelIDStr string) (*responses.PanelResponse, error) {
	panelID, err := uuid.Parse(panelIDStr)
	if err != nil {
		return nil, errors.New("invalid panel id")
	}

	panel, err := s.repos.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapPanelToResponse(panel)
	return &resp, nil
}

func (s *PanelService) EditPanel(ctx context.Context, userIDStr string, panelIDStr string, req *requests.UpdatePanelRequest) (*responses.PanelResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	panelID, err := uuid.Parse(panelIDStr)
	if err != nil {
		return nil, errors.New("invalid panel id")
	}

	canEdit, err := s.repos.Panel.CanEditPanel(ctx, userID, panelID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, ErrUnauthorizedPanel
	}

	existing, err := s.repos.Panel.GetPanelByID(ctx, panelID)
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

	updated, err := s.repos.Panel.UpdatePanel(ctx, panelID, existing)
	if err != nil {
		log.Printf("Error updating panel: %v", err)
		return nil, err
	}

	resp := mappers.MapPanelToResponse(updated)
	return &resp, nil
}

func (s *PanelService) DeletePanel(ctx context.Context, userIDStr string, panelIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user id")
	}

	panelID, err := uuid.Parse(panelIDStr)
	if err != nil {
		return errors.New("invalid panel id")
	}

	canDelete, err := s.repos.Panel.CanDeletePanel(ctx, userID, panelID)
	if err != nil {
		return err
	}
	if !canDelete {
		return ErrUnauthorizedPanel
	}

	if err := s.repos.Panel.DeletePanel(ctx, panelID); err != nil {
		log.Printf("Error deleting panel: %v", err)
		return err
	}
	return nil
}

func (s *PanelService) GetPendingPanels(ctx context.Context) ([]responses.PanelResponse, error) {
	panels, err := s.repos.Panel.GetPendingPanels(ctx)
	if err != nil {
		log.Printf("Error listing pending panels: %v", err)
		return nil, err
	}
	return mappers.MapPanelsToResponse(panels), nil
}

func (s *PanelService) GetApprovedPanels(ctx context.Context) ([]responses.PanelResponse, error) {
	panels, err := s.repos.Panel.GetApprovedPanels(ctx)
	if err != nil {
		log.Printf("Error listing approved panels: %v", err)
		return nil, err
	}
	return mappers.MapPanelsToResponse(panels), nil
}

func (s *PanelService) GetDeniedPanels(ctx context.Context) ([]responses.PanelResponse, error) {
	panels, err := s.repos.Panel.GetDeniedPanels(ctx)
	if err != nil {
		log.Printf("Error listing denied panels: %v", err)
		return nil, err
	}
	return mappers.MapPanelsToResponse(panels), nil
}

func (s *PanelService) setPanelStatus(ctx context.Context, panelIDStr string, status models.PanelStatus) (*responses.PanelResponse, error) {
	panelID, err := uuid.Parse(panelIDStr)
	if err != nil {
		return nil, errors.New("invalid panel id")
	}

	panel, err := s.repos.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		return nil, err
	}

	if panel.PanelStatus == status {
		return nil, ErrStatusUnchanged
	}

	if err := s.repos.Panel.SetPanelStatus(ctx, panelID, status); err != nil {
		log.Printf("Error setting panel status: %v", err)
		return nil, err
	}

	updated, err := s.repos.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapPanelToResponse(updated)
	return &resp, nil
}

func (s *PanelService) ApprovePanel(ctx context.Context, panelIDStr string) (*responses.PanelResponse, error) {
	return s.setPanelStatus(ctx, panelIDStr, models.PanelStatusApproved)
}

func (s *PanelService) DenyPanel(ctx context.Context, panelIDStr string) (*responses.PanelResponse, error) {
	return s.setPanelStatus(ctx, panelIDStr, models.PanelStatusDenied)
}

func (s *PanelService) MarkPanelPending(ctx context.Context, panelIDStr string) (*responses.PanelResponse, error) {
	return s.setPanelStatus(ctx, panelIDStr, models.PanelStatusPending)
}

// AssignPanelSchedule sets slot label and start time for an approved panel (admin/staff).
func (s *PanelService) AssignPanelSchedule(ctx context.Context, panelIDStr string, req *requests.AssignPanelScheduleRequest) (*responses.PanelResponse, error) {
	panelID, err := uuid.Parse(panelIDStr)
	if err != nil {
		return nil, errors.New("invalid panel id")
	}

	panel, err := s.repos.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		return nil, err
	}

	if panel.PanelStatus != models.PanelStatusApproved {
		return nil, ErrPanelNotSchedulable
	}

	if err := s.repos.Panel.SetPanelSchedule(ctx, panelID, req.SlotLabel, req.ScheduledStartAt); err != nil {
		log.Printf("Error assigning panel schedule: %v", err)
		return nil, err
	}

	updated, err := s.repos.Panel.GetPanelByID(ctx, panelID)
	if err != nil {
		return nil, err
	}

	resp := mappers.MapPanelToResponse(updated)
	return &resp, nil
}
