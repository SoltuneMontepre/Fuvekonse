package services

import (
	"fmt"
	"ticket-service/internal/models"
	"ticket-service/internal/repositories"

	"github.com/google/uuid"
)

type TicketServiceInterface interface {
	GetTicketTierByID(id uuid.UUID) (*models.TicketTier, error)
	GetAllTicketTiers() ([]models.TicketTier, error)
	GetActiveTicketTiers() ([]models.TicketTier, error)
	GetTicketByID(id uuid.UUID) (*models.Ticket, error)
	GetTicketsByUserID(userID uuid.UUID) ([]models.Ticket, error)
}

type TicketService struct {
	tierRepo   repositories.TicketTierRepositoryInterface
	ticketRepo repositories.TicketRepositoryInterface
}

func NewTicketService(
	tierRepo repositories.TicketTierRepositoryInterface,
	ticketRepo repositories.TicketRepositoryInterface,
) TicketServiceInterface {
	return &TicketService{
		tierRepo:   tierRepo,
		ticketRepo: ticketRepo,
	}
}

func (s *TicketService) GetTicketTierByID(id uuid.UUID) (*models.TicketTier, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid ticket tier ID")
	}

	tier, err := s.tierRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ticket tier not found: %w", err)
	}

	return tier, nil
}

func (s *TicketService) GetAllTicketTiers() ([]models.TicketTier, error) {
	return s.tierRepo.GetAll()
}

func (s *TicketService) GetActiveTicketTiers() ([]models.TicketTier, error) {
	return s.tierRepo.GetActive()
}

func (s *TicketService) GetTicketByID(id uuid.UUID) (*models.Ticket, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid ticket ID")
	}

	ticket, err := s.ticketRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	return ticket, nil
}

func (s *TicketService) GetTicketsByUserID(userID uuid.UUID) ([]models.Ticket, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.ticketRepo.GetByUserID(userID)
}

