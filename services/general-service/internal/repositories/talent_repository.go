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
	ErrTalentNotFound       = errors.New("talent not found")
	ErrTalentNotEditable    = errors.New("cannot edit talent unless status is pending")
	ErrUnauthorizedTalent   = errors.New("user is not the owner of this talent")
	ErrTalentNotSchedulable = errors.New("talent must be approved before assigning a schedule")
)

type TalentRepository struct {
	db *gorm.DB
}

func NewTalentRepository(db *gorm.DB) *TalentRepository {
	return &TalentRepository{db: db}
}

func (r *TalentRepository) CreateTalent(ctx context.Context, talent *models.PerformanceTalent) (*models.PerformanceTalent, error) {
	if talent.Id == uuid.Nil {
		talent.Id = uuid.New()
	}
	talent.CreatedAt = time.Now()
	talent.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(talent).Error; err != nil {
		return nil, err
	}
	return talent, nil
}

func (r *TalentRepository) GetTalentByID(ctx context.Context, id uuid.UUID) (*models.PerformanceTalent, error) {
	var talent models.PerformanceTalent
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&talent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTalentNotFound
		}
		return nil, err
	}
	return &talent, nil
}

func (r *TalentRepository) GetUserTalents(ctx context.Context, userID uuid.UUID) ([]models.PerformanceTalent, error) {
	var talents []models.PerformanceTalent
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("CASE WHEN scheduled_start_at IS NULL THEN 1 ELSE 0 END ASC, scheduled_start_at ASC, created_at DESC").
		Find(&talents).Error
	if err != nil {
		return nil, err
	}
	return talents, nil
}

func (r *TalentRepository) UpdateTalent(ctx context.Context, id uuid.UUID, talent *models.PerformanceTalent) (*models.PerformanceTalent, error) {
	existing, err := r.GetTalentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.TalentStatus != models.TalentStatusPending {
		return nil, ErrTalentNotEditable
	}

	talent.Id = id
	talent.UserId = existing.UserId
	talent.TalentStatus = existing.TalentStatus
	talent.SlotLabel = existing.SlotLabel
	talent.ScheduledStartAt = existing.ScheduledStartAt
	talent.CreatedAt = existing.CreatedAt
	talent.ModifiedAt = time.Now()

	if err := r.db.WithContext(ctx).Model(&models.PerformanceTalent{}).
		Where("id = ?", id).
		Updates(talent).Error; err != nil {
		return nil, err
	}

	return talent, nil
}

func (r *TalentRepository) DeleteTalent(ctx context.Context, id uuid.UUID) error {
	existing, err := r.GetTalentByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.TalentStatus == models.TalentStatusApproved {
		return ErrTalentNotEditable
	}

	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.PerformanceTalent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *TalentRepository) GetTalentsByStatus(ctx context.Context, status models.TalentStatus) ([]models.PerformanceTalent, error) {
	var talents []models.PerformanceTalent
	err := r.db.WithContext(ctx).
		Where("talent_status = ? AND is_deleted = ?", status, false).
		Preload("User").
		Order("CASE WHEN scheduled_start_at IS NULL THEN 1 ELSE 0 END ASC, scheduled_start_at ASC, created_at ASC").
		Find(&talents).Error
	if err != nil {
		return nil, err
	}
	return talents, nil
}

func (r *TalentRepository) GetPendingTalents(ctx context.Context) ([]models.PerformanceTalent, error) {
	return r.GetTalentsByStatus(ctx, models.TalentStatusPending)
}

func (r *TalentRepository) GetApprovedTalents(ctx context.Context) ([]models.PerformanceTalent, error) {
	return r.GetTalentsByStatus(ctx, models.TalentStatusApproved)
}

func (r *TalentRepository) GetDeniedTalents(ctx context.Context) ([]models.PerformanceTalent, error) {
	return r.GetTalentsByStatus(ctx, models.TalentStatusDenied)
}

func (r *TalentRepository) SetTalentStatus(ctx context.Context, id uuid.UUID, status models.TalentStatus) error {
	if _, err := r.GetTalentByID(ctx, id); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"talent_status": status,
		"modified_at":   time.Now(),
	}
	if status != models.TalentStatusApproved {
		updates["slot_label"] = ""
		updates["scheduled_start_at"] = nil
	}

	return r.db.WithContext(ctx).Model(&models.PerformanceTalent{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *TalentRepository) SetTalentSchedule(ctx context.Context, id uuid.UUID, slotLabel string, startAt time.Time) error {
	if _, err := r.GetTalentByID(ctx, id); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&models.PerformanceTalent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"slot_label":         slotLabel,
			"scheduled_start_at": startAt,
			"modified_at":        time.Now(),
		}).Error
}

func (r *TalentRepository) CanEditTalent(ctx context.Context, userID uuid.UUID, talentID uuid.UUID) (bool, error) {
	talent, err := r.GetTalentByID(ctx, talentID)
	if err != nil {
		return false, err
	}

	if talent.UserId != userID || talent.TalentStatus != models.TalentStatusPending {
		return false, nil
	}

	return true, nil
}

func (r *TalentRepository) CanDeleteTalent(ctx context.Context, userID uuid.UUID, talentID uuid.UUID) (bool, error) {
	talent, err := r.GetTalentByID(ctx, talentID)
	if err != nil {
		return false, err
	}

	if talent.UserId != userID || talent.TalentStatus == models.TalentStatusApproved {
		return false, nil
	}

	return true, nil
}

