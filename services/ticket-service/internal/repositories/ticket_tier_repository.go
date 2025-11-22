package repositories

import (
	"ticket-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTierRepositoryInterface interface {
	Create(tier *models.TicketTier) error
	GetByID(id uuid.UUID) (*models.TicketTier, error)
	GetAll() ([]models.TicketTier, error)
	GetActive() ([]models.TicketTier, error)
	Update(tier *models.TicketTier) error
	Delete(id uuid.UUID) error
	DecrementStock(id uuid.UUID) error
	IncrementStock(id uuid.UUID) error
}

type TicketTierRepository struct {
	db *gorm.DB
}

func NewTicketTierRepository(db *gorm.DB) TicketTierRepositoryInterface {
	return &TicketTierRepository{db: db}
}

func (r *TicketTierRepository) Create(tier *models.TicketTier) error {
	return r.db.Create(tier).Error
}

func (r *TicketTierRepository) GetByID(id uuid.UUID) (*models.TicketTier, error) {
	var tier models.TicketTier
	err := r.db.First(&tier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *TicketTierRepository) GetAll() ([]models.TicketTier, error) {
	var tiers []models.TicketTier
	err := r.db.Find(&tiers).Error
	return tiers, err
}

func (r *TicketTierRepository) GetActive() ([]models.TicketTier, error) {
	var tiers []models.TicketTier
	err := r.db.Where("is_active = ?", true).Find(&tiers).Error
	return tiers, err
}

func (r *TicketTierRepository) Update(tier *models.TicketTier) error {
	return r.db.Save(tier).Error
}

func (r *TicketTierRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TicketTier{}, "id = ?", id).Error
}

func (r *TicketTierRepository) DecrementStock(id uuid.UUID) error {
	return r.db.Model(&models.TicketTier{}).
		Where("id = ? AND stock > 0", id).
		UpdateColumn("stock", gorm.Expr("stock - 1")).Error
}

func (r *TicketTierRepository) IncrementStock(id uuid.UUID) error {
	return r.db.Model(&models.TicketTier{}).
		Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + 1")).Error
}
