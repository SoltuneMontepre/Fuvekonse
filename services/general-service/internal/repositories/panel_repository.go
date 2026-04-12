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
	ErrPanelNotFound       = errors.New("panel not found")
	ErrPanelNotEditable    = errors.New("cannot edit panel unless status is pending")
	ErrUnauthorizedPanel   = errors.New("user is not the owner of this panel")
	ErrPanelNotSchedulable = errors.New("panel must be approved before assigning a schedule")
)

type PanelRepository struct {
	db *gorm.DB
}

func NewPanelRepository(db *gorm.DB) *PanelRepository {
	return &PanelRepository{db: db}
}

func (r *PanelRepository) CreatePanel(ctx context.Context, panel *models.PerformancePanel) (*models.PerformancePanel, error) {
	if panel.Id == uuid.Nil {
		panel.Id = uuid.New()
	}
	panel.CreatedAt = time.Now()
	panel.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(panel).Error; err != nil {
		return nil, err
	}
	return panel, nil
}

func (r *PanelRepository) GetPanelByID(ctx context.Context, id uuid.UUID) (*models.PerformancePanel, error) {
	var panel models.PerformancePanel
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&panel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPanelNotFound
		}
		return nil, err
	}
	return &panel, nil
}

func (r *PanelRepository) GetUserPanels(ctx context.Context, userID uuid.UUID) ([]models.PerformancePanel, error) {
	var panels []models.PerformancePanel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("CASE WHEN scheduled_start_at IS NULL THEN 1 ELSE 0 END ASC, scheduled_start_at ASC, created_at DESC").
		Find(&panels).Error
	if err != nil {
		return nil, err
	}
	return panels, nil
}

func (r *PanelRepository) UpdatePanel(ctx context.Context, id uuid.UUID, panel *models.PerformancePanel) (*models.PerformancePanel, error) {
	existing, err := r.GetPanelByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.PanelStatus != models.PanelStatusPending {
		return nil, ErrPanelNotEditable
	}

	panel.Id = id
	panel.UserId = existing.UserId
	panel.PanelStatus = existing.PanelStatus
	panel.SlotLabel = existing.SlotLabel
	panel.ScheduledStartAt = existing.ScheduledStartAt
	panel.CreatedAt = existing.CreatedAt
	panel.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.PerformancePanel{}).
		Where("id = ?", id).
		Updates(panel).Error; err != nil {
		return nil, err
	}

	return panel, nil
}

func (r *PanelRepository) DeletePanel(ctx context.Context, id uuid.UUID) error {
	existing, err := r.GetPanelByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.PanelStatus == models.PanelStatusApproved {
		return ErrPanelNotEditable
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.PerformancePanel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *PanelRepository) GetPanelsByStatus(ctx context.Context, status models.PanelStatus) ([]models.PerformancePanel, error) {
	var panels []models.PerformancePanel
	err := r.db.WithContext(ctx).
		Where("panel_status = ? AND is_deleted = ?", status, false).
		Preload("User").
		Order("CASE WHEN scheduled_start_at IS NULL THEN 1 ELSE 0 END ASC, scheduled_start_at ASC, created_at ASC").
		Find(&panels).Error
	if err != nil {
		return nil, err
	}
	return panels, nil
}

func (r *PanelRepository) GetPendingPanels(ctx context.Context) ([]models.PerformancePanel, error) {
	return r.GetPanelsByStatus(ctx, models.PanelStatusPending)
}

func (r *PanelRepository) GetApprovedPanels(ctx context.Context) ([]models.PerformancePanel, error) {
	return r.GetPanelsByStatus(ctx, models.PanelStatusApproved)
}

func (r *PanelRepository) GetDeniedPanels(ctx context.Context) ([]models.PerformancePanel, error) {
	return r.GetPanelsByStatus(ctx, models.PanelStatusDenied)
}

func (r *PanelRepository) SetPanelStatus(ctx context.Context, id uuid.UUID, status models.PanelStatus) error {
	if _, err := r.GetPanelByID(ctx, id); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"panel_status": status,
		"modified_at":  time.Now(),
	}
	// Drop schedule when the panel is no longer approved (re-review or deny).
	if status != models.PanelStatusApproved {
		updates["slot_label"] = ""
		updates["scheduled_start_at"] = nil
	}

	return r.db.WithContext(ctx).Model(&models.PerformancePanel{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// SetPanelSchedule assigns venue/track slot label and start time (approved panels only; enforced in service).
func (r *PanelRepository) SetPanelSchedule(ctx context.Context, id uuid.UUID, slotLabel string, startAt time.Time) error {
	if _, err := r.GetPanelByID(ctx, id); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&models.PerformancePanel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"slot_label":         slotLabel,
			"scheduled_start_at": startAt,
			"modified_at":        time.Now(),
		}).Error
}

func (r *PanelRepository) CanEditPanel(ctx context.Context, userID uuid.UUID, panelID uuid.UUID) (bool, error) {
	panel, err := r.GetPanelByID(ctx, panelID)
	if err != nil {
		return false, err
	}

	if panel.UserId != userID || panel.PanelStatus != models.PanelStatusPending {
		return false, nil
	}

	return true, nil
}

func (r *PanelRepository) CanDeletePanel(ctx context.Context, userID uuid.UUID, panelID uuid.UUID) (bool, error) {
	panel, err := r.GetPanelByID(ctx, panelID)
	if err != nil {
		return false, err
	}

	if panel.UserId != userID || panel.PanelStatus == models.PanelStatusApproved {
		return false, nil
	}

	return true, nil
}
