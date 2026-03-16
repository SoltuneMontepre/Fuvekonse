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

// SetVerified sets the is_verified flag for a user by ID (used after OTP verification).
func (r *UserRepository) SetVerified(userID string, verified bool) error {
	return r.db.Model(&models.User{}).Where("id = ? AND is_deleted = ?", userID, false).
		Update("is_verified", verified).Error
}

// Count returns the total number of non-deleted users (for analytics).
func (r *UserRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&models.User{}).Where("is_deleted = ?", false).Count(&total).Error
	return total, err
}

// FindAll finds all users with pagination and optional search (email, first_name, last_name, fursona_name)
func (r *UserRepository) FindAll(page, pageSize int, search string) ([]*models.User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, errors.New("page must be >= 1")
	}
	if pageSize <= 0 {
		return nil, 0, errors.New("pageSize must be > 0")
	}

	// Use a fresh session for count so no Limit/Offset from other chains can affect it
	countDB := r.db.Session(&gorm.Session{}).Model(&models.User{}).Where("is_deleted = ?", false)
	if search != "" {
		pattern := "%" + search + "%"
		countDB = countDB.Where(
			"email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR fursona_name ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var users []*models.User
	query := r.db.Session(&gorm.Session{}).Where("is_deleted = ?", false)
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where(
			"email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR fursona_name ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if err := query.Order("created_at DESC").
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

// CountByCountryResult holds country code and count for aggregation
type CountByCountryResult struct {
	Country string `gorm:"column:country"`
	Count   int64  `gorm:"column:count"`
}

// CountByCountry returns counts of non-deleted users grouped by country.
// Empty or NULL country is returned as empty string.
func (r *UserRepository) CountByCountry() ([]CountByCountryResult, error) {
	var results []CountByCountryResult
	err := r.db.Model(&models.User{}).
		Select("COALESCE(country, '') AS country, COUNT(*) AS count").
		Where("is_deleted = ?", false).
		Group("COALESCE(country, '')").
		Order("count DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
