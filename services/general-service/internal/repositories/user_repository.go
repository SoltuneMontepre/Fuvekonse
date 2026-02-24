package repositories

import (
	"errors"
	"general-service/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds a user by email (case-insensitive)
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = LOWER(?) AND is_deleted = ?", email, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleId finds a user by Google OAuth subject ID
func (r *UserRepository) FindByGoogleId(googleId string) (*models.User, error) {
	if googleId == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var user models.User
	err := r.db.Where("google_id = ? AND is_deleted = ?", googleId, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUserProfile(user *models.User) error {
	return r.db.Save(user).Error
}

// FindAll finds all users with pagination support
func (r *UserRepository) FindAll(page, pageSize int) ([]*models.User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, errors.New("page must be >= 1")
	}
	if pageSize <= 0 {
		return nil, 0, errors.New("pageSize must be > 0")
	}

	var users []*models.User
	var total int64

	// Count total records
	if err := r.db.Model(&models.User{}).Where("is_deleted = ?", false).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch users with pagination
	if err := r.db.Where("is_deleted = ?", false).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindByIDForAdmin finds a user by ID (includes deleted users for admin)
func (r *UserRepository) FindByIDForAdmin(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser soft deletes a user
func (r *UserRepository) DeleteUser(user *models.User) error {
	now := time.Now()
	user.IsDeleted = true
	user.DeletedAt = &now
	return r.db.Save(user).Error
}

// UpdateUser updates user information (admin use)
func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
