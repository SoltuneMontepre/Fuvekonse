package services

import (
	"fmt"
	"gin/internal/models"
	"gin/internal/repositories"
	"time"
)

type UserBanServiceInterface interface {
	BanUser(userID string, permissionID uint, reason string) (*models.UserBan, error)
	UnbanUser(userID string, permissionID uint) error
	GetUserBan(id uint) (*models.UserBan, error)
	GetUserBans(userID string) ([]models.UserBan, error)
	GetAllUserBans() ([]models.UserBan, error)
	IsUserBanned(userID string, permissionID uint) (bool, error)
	GetActiveUserBans(userID string) ([]models.UserBan, error)
	GetRecentBans(days int, limit int) ([]models.UserBan, error)
	UpdateBanReason(id uint, reason string) error
}

type UserBanService struct {
	userBanRepo    repositories.UserBanRepositoryInterface
	permissionRepo repositories.PermissionRepositoryInterface
}

func NewUserBanService(userBanRepo repositories.UserBanRepositoryInterface, permissionRepo repositories.PermissionRepositoryInterface) UserBanServiceInterface {
	return &UserBanService{
		userBanRepo:    userBanRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *UserBanService) BanUser(userID string, permissionID uint, reason string) (*models.UserBan, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if permissionID == 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	if reason == "" {
		return nil, fmt.Errorf("ban reason cannot be empty")
	}

	_, err := s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	existingBan, err := s.userBanRepo.GetByUserIDAndPermission(userID, permissionID)
	if err == nil && existingBan != nil {
		return nil, fmt.Errorf("user is already banned for this permission")
	}

	userBan := &models.UserBan{
		UserID:    userID,
		PermID:    permissionID,
		Reason:    reason,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userBanRepo.Create(userBan); err != nil {
		return nil, fmt.Errorf("failed to create user ban: %w", err)
	}

	return userBan, nil
}

func (s *UserBanService) UnbanUser(userID string, permissionID uint) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if permissionID == 0 {
		return fmt.Errorf("invalid permission ID")
	}

	existingBan, err := s.userBanRepo.GetByUserIDAndPermission(userID, permissionID)
	if err != nil {
		return fmt.Errorf("ban not found: %w", err)
	}

	return s.userBanRepo.Delete(existingBan.ID)
}

func (s *UserBanService) GetUserBan(id uint) (*models.UserBan, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid ban ID")
	}

	userBan, err := s.userBanRepo.GetWithPermission(id)
	if err != nil {
		return nil, fmt.Errorf("user ban not found: %w", err)
	}

	return userBan, nil
}

func (s *UserBanService) GetUserBans(userID string) ([]models.UserBan, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	return s.userBanRepo.GetByUserID(userID)
}

func (s *UserBanService) GetAllUserBans() ([]models.UserBan, error) {
	return s.userBanRepo.GetAll()
}

func (s *UserBanService) IsUserBanned(userID string, permissionID uint) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	if permissionID == 0 {
		return false, fmt.Errorf("invalid permission ID")
	}

	_, err := s.userBanRepo.GetByUserIDAndPermission(userID, permissionID)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (s *UserBanService) GetActiveUserBans(userID string) ([]models.UserBan, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	return s.userBanRepo.GetByUserID(userID)
}

func (s *UserBanService) GetRecentBans(days int, limit int) ([]models.UserBan, error) {
	if days <= 0 {
		days = 30
	}

	allBans, err := s.userBanRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get bans: %w", err)
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)
	var recentBans []models.UserBan

	for _, ban := range allBans {
		if ban.CreatedAt.After(cutoffDate) {
			recentBans = append(recentBans, ban)
		}
	}

	if limit > 0 && len(recentBans) > limit {
		recentBans = recentBans[:limit]
	}

	return recentBans, nil
}

func (s *UserBanService) UpdateBanReason(id uint, reason string) error {
	if id == 0 {
		return fmt.Errorf("invalid ban ID")
	}

	if reason == "" {
		return fmt.Errorf("ban reason cannot be empty")
	}

	userBan, err := s.userBanRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user ban not found: %w", err)
	}

	userBan.Reason = reason
	userBan.UpdatedAt = time.Now()

	return s.userBanRepo.Update(userBan)
}