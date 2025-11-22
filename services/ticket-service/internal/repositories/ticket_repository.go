package repositories

import (
	"ticket-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketRepositoryInterface interface {
	Create(ticket *models.Ticket) error
	GetByID(id uuid.UUID) (*models.Ticket, error)
	GetByOrderCode(orderCode int64) (*models.Ticket, error)
	GetByUserID(userID uuid.UUID) ([]models.Ticket, error)
	Update(ticket *models.Ticket) error
	Delete(id uuid.UUID) error
}

type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepositoryInterface {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *TicketRepository) GetByID(id uuid.UUID) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("TicketTier").Preload("Payments").
		First(&ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) GetByOrderCode(orderCode int64) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("TicketTier").Preload("Payments").
		Where("order_code = ?", orderCode).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) GetByUserID(userID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := r.db.Preload("TicketTier").
		Where("user_id = ?", userID).Find(&tickets).Error
	return tickets, err
}

func (r *TicketRepository) Update(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *TicketRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Ticket{}, "id = ?", id).Error
}

