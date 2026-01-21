package repositories

import (
	"general-service/internal/models"
	"time"

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

// FindAllBooths finds all dealer booths with pagination
func (r *DealerRepository) FindAllBooths(page, pageSize int, isVerified *bool) ([]*models.DealerBooth, int64, error) {
	var booths []*models.DealerBooth
	var total int64

	query := r.db.Model(&models.DealerBooth{}).Where("is_deleted = ?", false)

	// Apply filters
	if isVerified != nil {
		query = query.Where("is_verified = ?", *isVerified)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Fetch booths with pagination
	if err := query.
		Preload("Staffs.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&booths).Error; err != nil {
		return nil, 0, err
	}

	return booths, total, nil
}

// FindBoothByIDWithStaffs finds a dealer booth by ID with staff preloaded
func (r *DealerRepository) FindBoothByIDWithStaffs(id string) (*models.DealerBooth, error) {
	var booth models.DealerBooth
	err := r.db.
		Preload("Staffs.User").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&booth).Error
	if err != nil {
		return nil, err
	}
	return &booth, nil
}

// VerifyBooth verifies a dealer booth and assigns a booth number
func (r *DealerRepository) VerifyBooth(id string, boothNumber string) (*models.DealerBooth, error) {
	var booth models.DealerBooth
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&booth).Error
	if err != nil {
		return nil, err
	}

	booth.IsVerified = true
	booth.BoothNumber = boothNumber

	if err := r.db.Save(&booth).Error; err != nil {
		return nil, err
	}

	return &booth, nil
}

// CheckBoothNumberExists checks if a booth number already exists
func (r *DealerRepository) CheckBoothNumberExists(boothNumber string) (bool, error) {
	var count int64
	err := r.db.Model(&models.DealerBooth{}).
		Where("booth_number = ? AND is_deleted = ?", boothNumber, false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindBoothByCode finds a dealer booth by booth code
func (r *DealerRepository) FindBoothByCode(boothCode string) (*models.DealerBooth, error) {
	var booth models.DealerBooth
	err := r.db.Where("booth_number = ? AND is_deleted = ?", boothCode, false).First(&booth).Error
	if err != nil {
		return nil, err
	}
	return &booth, nil
}

// CheckUserIsStaff checks if a user is already a staff member of any booth
func (r *DealerRepository) CheckUserIsStaff(userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserDealerStaff{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindStaffByUserAndBoothID finds a staff record by user ID and booth ID
func (r *DealerRepository) FindStaffByUserAndBoothID(userID string, boothID string) (*models.UserDealerStaff, error) {
	var staff models.UserDealerStaff
	err := r.db.Where("user_id = ? AND booth_id = ? AND is_deleted = ?", userID, boothID, false).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// RemoveStaff soft deletes a staff member from a booth
func (r *DealerRepository) RemoveStaff(staffID string) error {
	var staff models.UserDealerStaff
	err := r.db.Where("id = ? AND is_deleted = ?", staffID, false).First(&staff).Error
	if err != nil {
		return err
	}

	now := time.Now()
	staff.IsDeleted = true
	staff.DeletedAt = &now

	return r.db.Save(&staff).Error
}

// GetBoothByStaffUserID gets the booth that a user is a staff member of
func (r *DealerRepository) GetBoothByStaffUserID(userID string) (*models.DealerBooth, error) {
	var staff models.UserDealerStaff
	err := r.db.
		Preload("Booth").
		Where("user_id = ? AND is_deleted = ?", userID, false).
		First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff.Booth, nil
}
