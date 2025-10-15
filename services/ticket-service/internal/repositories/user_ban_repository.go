package repositories

import (
	"gin/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserBanRepositoryInterface interface {
	Create(userBan *models.UserBan) error
	GetByID(id uint) (*models.UserBan, error)
	GetByUserID(userID string) ([]models.UserBan, error)
	GetByUserIDAndPermission(userID string, permID uint) (*models.UserBan, error)
	GetAll() ([]models.UserBan, error)
	Update(userBan *models.UserBan) error
	Delete(id uint) error
	GetWithPermission(id uint) (*models.UserBan, error)
	GetActiveUserBans(userID string) ([]models.UserBan, error)
	IsUserBanned(userID string, permID uint) (bool, error)
	BanUser(userID string, permID uint, reason string) error
	UnbanUser(userID string, permID uint) error
}

type UserBanRepository struct {
	db *gorm.DB
}

func NewUserBanRepository(db *gorm.DB) UserBanRepositoryInterface {
	return &UserBanRepository{db: db}
}

func (u *UserBanRepository) Create(userBan *models.UserBan) error {
	return u.db.Create(userBan).Error
}

func (u *UserBanRepository) GetByID(id uint) (*models.UserBan, error) {
	var userBan models.UserBan
	err := u.db.First(&userBan, id).Error
	if err != nil {
		return nil, err
	}
	return &userBan, nil
}

func (u *UserBanRepository) GetByUserID(userID string) ([]models.UserBan, error) {
	var userBans []models.UserBan
	err := u.db.Where("user_id = ?", userID).Find(&userBans).Error
	return userBans, err
}

func (u *UserBanRepository) GetByUserIDAndPermission(userID string, permID uint) (*models.UserBan, error) {
	var userBan models.UserBan
	err := u.db.Where("user_id = ? AND perm_id = ?", userID, permID).First(&userBan).Error
	if err != nil {
		return nil, err
	}
	return &userBan, nil
}

func (u *UserBanRepository) GetAll() ([]models.UserBan, error) {
	var userBans []models.UserBan
	err := u.db.Find(&userBans).Error
	return userBans, err
}

func (u *UserBanRepository) Update(userBan *models.UserBan) error {
	return u.db.Save(userBan).Error
}

func (u *UserBanRepository) Delete(id uint) error {
	return u.db.Delete(&models.UserBan{}, id).Error
}

func (u *UserBanRepository) GetWithPermission(id uint) (*models.UserBan, error) {
	var userBan models.UserBan
	err := u.db.Preload("Permission").First(&userBan, id).Error
	if err != nil {
		return nil, err
	}
	return &userBan, nil
}

func (u *UserBanRepository) GetActiveUserBans(userID string) ([]models.UserBan, error) {
	var userBans []models.UserBan
	err := u.db.Where("user_id = ?", userID).Preload("Permission").Find(&userBans).Error
	return userBans, err
}

func (u *UserBanRepository) IsUserBanned(userID string, permID uint) (bool, error) {
	var count int64
	err := u.db.Model(&models.UserBan{}).Where("user_id = ? AND perm_id = ?", userID, permID).Count(&count).Error
	return count > 0, err
}

func (u *UserBanRepository) BanUser(userID string, permID uint, reason string) error {
	userBan := &models.UserBan{
		UserID:    userID,
		PermID:    permID,
		Reason:    reason,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return u.Create(userBan)
}

func (u *UserBanRepository) UnbanUser(userID string, permID uint) error {
	return u.db.Where("user_id = ? AND perm_id = ?", userID, permID).Delete(&models.UserBan{}).Error
}

func (u *UserBanRepository) GetRecentBans(limit int) ([]models.UserBan, error) {
	var userBans []models.UserBan
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	query := u.db.Where("created_at > ?", thirtyDaysAgo).
		Preload("Permission").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&userBans).Error
	return userBans, err
}