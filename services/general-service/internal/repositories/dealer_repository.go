package repositories

import (
	"general-service/internal/models"

	"gorm.io/gorm"
)

type DealerRepository struct {
	db *gorm.DB
}

func NewDealerRepository(db *gorm.DB) *DealerRepository {
	return &DealerRepository{db: db}
}

// CreateBooth creates a new dealer booth
func (r *DealerRepository) CreateBooth(booth *models.DealerBooth) error {
	return r.db.Create(booth).Error
}

// CreateStaff creates a new dealer staff
func (r *DealerRepository) CreateStaff(staff *models.UserDealerStaff) error {
	return r.db.Create(staff).Error
}

// CreateBoothWithStaff creates a dealer booth and staff in a transaction
func (r *DealerRepository) CreateBoothWithStaff(booth *models.DealerBooth, staff *models.UserDealerStaff) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create booth
		if err := tx.Create(booth).Error; err != nil {
			return err
		}

		// Set the booth ID for the staff
		staff.BoothId = booth.Id

		// Create staff
		if err := tx.Create(staff).Error; err != nil {
			return err
		}

		return nil
	})
}

// FindBoothByID finds a dealer booth by ID
func (r *DealerRepository) FindBoothByID(id string) (*models.DealerBooth, error) {
	var booth models.DealerBooth
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&booth).Error
	if err != nil {
		return nil, err
	}
	return &booth, nil
}

// FindStaffByUserID finds dealer staff by user ID
func (r *DealerRepository) FindStaffByUserID(userID string) (*models.UserDealerStaff, error) {
	var staff models.UserDealerStaff
	err := r.db.Where("user_id = ? AND is_deleted = ?", userID, false).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}
