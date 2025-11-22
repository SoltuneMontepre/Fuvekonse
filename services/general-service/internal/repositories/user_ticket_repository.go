package repositories

import (
	"general-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTicketRepositoryInterface interface {
	Create(userTicket *models.UserTicket) error
	GetByID(id uuid.UUID) (*models.UserTicket, error)
	GetByUserID(userID uuid.UUID) ([]models.UserTicket, error)
	Update(userTicket *models.UserTicket) error
	Delete(id uuid.UUID) error
}

type UserTicketRepository struct {
	db *gorm.DB
}

func NewUserTicketRepository(db *gorm.DB) *UserTicketRepository {
	return &UserTicketRepository{db: db}
}

func (r *UserTicketRepository) Create(userTicket *models.UserTicket) error {
	return r.db.Create(userTicket).Error
}

func (r *UserTicketRepository) GetByID(id uuid.UUID) (*models.UserTicket, error) {
	var userTicket models.UserTicket
	err := r.db.Preload("User").First(&userTicket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &userTicket, nil
}

func (r *UserTicketRepository) GetByUserID(userID uuid.UUID) ([]models.UserTicket, error) {
	var userTickets []models.UserTicket
	err := r.db.Preload("User").Where("user_id = ?", userID).Find(&userTickets).Error
	return userTickets, err
}

func (r *UserTicketRepository) Update(userTicket *models.UserTicket) error {
	return r.db.Save(userTicket).Error
}

func (r *UserTicketRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.UserTicket{}, "id = ?", id).Error
}
