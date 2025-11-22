package services

import (
	"general-service/internal/models"
	"general-service/internal/repositories"

	"github.com/google/uuid"
)

type UserTicketServiceInterface interface {
	CreateUserTicket(userTicket *models.UserTicket) error
	GetUserTicketByID(id uuid.UUID) (*models.UserTicket, error)
	GetUserTicketsByUserID(userID uuid.UUID) ([]models.UserTicket, error)
	UpdateUserTicket(userTicket *models.UserTicket) error
	DeleteUserTicket(id uuid.UUID) error
}

type UserTicketService struct {
	userTicketRepo *repositories.UserTicketRepository
}

func NewUserTicketService(repos *repositories.Repositories) *UserTicketService {
	return &UserTicketService{
		userTicketRepo: repos.UserTicket,
	}
}

func (s *UserTicketService) CreateUserTicket(userTicket *models.UserTicket) error {
	return s.userTicketRepo.Create(userTicket)
}

func (s *UserTicketService) GetUserTicketByID(id uuid.UUID) (*models.UserTicket, error) {
	return s.userTicketRepo.GetByID(id)
}

func (s *UserTicketService) GetUserTicketsByUserID(userID uuid.UUID) ([]models.UserTicket, error) {
	return s.userTicketRepo.GetByUserID(userID)
}

func (s *UserTicketService) UpdateUserTicket(userTicket *models.UserTicket) error {
	return s.userTicketRepo.Update(userTicket)
}

func (s *UserTicketService) DeleteUserTicket(id uuid.UUID) error {
	return s.userTicketRepo.Delete(id)
}
