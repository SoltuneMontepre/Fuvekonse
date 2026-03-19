package repositories

import (
	"context"
	"errors"
	"general-service/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrConbookNotFound     = errors.New("conbook not found")
	ErrConbookLimit        = errors.New("maximum conbook uploads (10) reached")
	ErrConbookNotEditable  = errors.New("cannot edit conbook unless status is pending")
	ErrUnauthorizedConbook = errors.New("user is not the owner of this conbook")
)

type ConbookRepository struct {
	db *gorm.DB
}

func NewConbookRepository(db *gorm.DB) *ConbookRepository {
	return &ConbookRepository{db: db}
}

// CreateConbook creates a new conbook
func (r *ConbookRepository) CreateConbook(ctx context.Context, conbook *models.ConBookArt) (*models.ConBookArt, error) {
	if conbook.Id == uuid.Nil {
		conbook.Id = uuid.New()
	}
	conbook.CreatedAt = time.Now()
	conbook.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(conbook).Error; err != nil {
		return nil, err
	}
	return conbook, nil
}

// GetConbookByID retrieves a conbook by ID (excludes deleted)
func (r *ConbookRepository) GetConbookByID(ctx context.Context, id uuid.UUID) (*models.ConBookArt, error) {
	var conbook models.ConBookArt
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&conbook).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConbookNotFound
		}
		return nil, err
	}
	return &conbook, nil
}

// GetUserConbooks retrieves all conbooks for a user (excludes deleted)
func (r *ConbookRepository) GetUserConbooks(ctx context.Context, userID uuid.UUID) ([]models.ConBookArt, error) {
	var conbooks []models.ConBookArt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&conbooks).Error
	if err != nil {
		return nil, err
	}
	return conbooks, nil
}

// GetUserConbookCount returns the count of non-deleted conbooks for a user
func (r *ConbookRepository) GetUserConbookCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ConBookArt{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	return count, err
}

// UpdateConbook updates a conbook (only if status is pending)
func (r *ConbookRepository) UpdateConbook(ctx context.Context, id uuid.UUID, conbook *models.ConBookArt) (*models.ConBookArt, error) {
	// Check if exists and is still editable.
	existing, err := r.GetConbookByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.ConBookArtStatus != models.ConbookStatusPending {
		return nil, ErrConbookNotEditable
	}

	conbook.Id = id
	conbook.UserId = existing.UserId
	conbook.CreatedAt = existing.CreatedAt
	conbook.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.ConBookArt{}).
		Where("id = ?", id).
		Updates(conbook).Error; err != nil {
		return nil, err
	}

	return conbook, nil
}

// DeleteConbook soft deletes a conbook (only if status is pending or denied)
func (r *ConbookRepository) DeleteConbook(ctx context.Context, id uuid.UUID) error {
	existing, err := r.GetConbookByID(ctx, id)
	if err != nil {
		return err
	}

	// Allow deletion only if status is pending or denied, but not approved
	if existing.ConBookArtStatus == models.ConbookStatusApproved {
		return ErrConbookNotEditable
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.ConBookArt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

// GetConbooksByStatus retrieves all conbooks by status.
func (r *ConbookRepository) GetConbooksByStatus(ctx context.Context, status models.ConbookStatus) ([]models.ConBookArt, error) {
	var conbooks []models.ConBookArt
	err := r.db.WithContext(ctx).
		Where("con_book_art_status = ? AND is_deleted = ?", status, false).
		Preload("User").
		Order("created_at ASC").
		Find(&conbooks).Error
	if err != nil {
		return nil, err
	}
	return conbooks, nil
}

// GetPendingConbooks retrieves all pending conbooks.
func (r *ConbookRepository) GetPendingConbooks(ctx context.Context) ([]models.ConBookArt, error) {
	return r.GetConbooksByStatus(ctx, models.ConbookStatusPending)
}

// GetApprovedConbooks retrieves all approved conbooks.
func (r *ConbookRepository) GetApprovedConbooks(ctx context.Context) ([]models.ConBookArt, error) {
	return r.GetConbooksByStatus(ctx, models.ConbookStatusApproved)
}

// GetDeniedConbooks retrieves all denied conbooks.
func (r *ConbookRepository) GetDeniedConbooks(ctx context.Context) ([]models.ConBookArt, error) {
	return r.GetConbooksByStatus(ctx, models.ConbookStatusDenied)
}

// SetConbookStatus updates conbook status (staff only).
func (r *ConbookRepository) SetConbookStatus(ctx context.Context, id uuid.UUID, status models.ConbookStatus) error {
	if _, err := r.GetConbookByID(ctx, id); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&models.ConBookArt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"con_book_art_status": status,
			"modified_at":         time.Now(),
		}).Error
}

// CanEditConbook checks if a user can edit a conbook
func (r *ConbookRepository) CanEditConbook(ctx context.Context, userID uuid.UUID, conbookID uuid.UUID) (bool, error) {
	conbook, err := r.GetConbookByID(ctx, conbookID)
	if err != nil {
		return false, err
	}

	// User must be the owner and conbook must still be pending.
	if conbook.UserId != userID || conbook.ConBookArtStatus != models.ConbookStatusPending {
		return false, nil
	}

	return true, nil
}

// CanDeleteConbook checks if a user can delete a conbook
func (r *ConbookRepository) CanDeleteConbook(ctx context.Context, userID uuid.UUID, conbookID uuid.UUID) (bool, error) {
	conbook, err := r.GetConbookByID(ctx, conbookID)
	if err != nil {
		return false, err
	}

	// User must be the owner and status must be pending or denied (not approved)
	if conbook.UserId != userID || conbook.ConBookArtStatus == models.ConbookStatusApproved {
		return false, nil
	}

	return true, nil
}
