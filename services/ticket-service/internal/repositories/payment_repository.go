package repositories

import (
	"ticket-service/internal/common/constants"
	"ticket-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepositoryInterface interface {
	Create(payment *models.Payment) error
	GetByID(id uuid.UUID) (*models.Payment, error)
	GetByOrderCode(orderCode int64) (*models.Payment, error)
	GetByTicketID(ticketID uuid.UUID) ([]models.Payment, error)
	GetExpiredReservations() ([]models.Payment, error)
	Update(payment *models.Payment) error
	Delete(id uuid.UUID) error
}

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepositoryInterface {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) GetByID(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Ticket").First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByOrderCode(orderCode int64) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Ticket").
		Where("order_code = ?", orderCode).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByTicketID(ticketID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("ticket_id = ?", ticketID).Find(&payments).Error
	return payments, err
}

func (r *PaymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *PaymentRepository) GetExpiredReservations() ([]models.Payment, error) {
	var payments []models.Payment
	// Get all pending payments - the service will filter by age
	err := r.db.Where("status = ?", constants.PaymentStatusPending.String()).
		Find(&payments).Error
	return payments, err
}

func (r *PaymentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Payment{}, "id = ?", id).Error
}
