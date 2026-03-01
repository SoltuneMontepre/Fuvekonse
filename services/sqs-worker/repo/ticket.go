package repo

import (
	"context"
	"errors"
	"fmt"
	"fuvekonse/sqs-worker/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTicketTierNotFound   = errors.New("ticket tier not found")
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrOutOfStock           = errors.New("ticket tier is out of stock")
	ErrUserAlreadyHasTicket = errors.New("user already has a ticket")
	ErrUserBlacklisted      = errors.New("user is blacklisted from purchasing tickets")
	ErrInvalidTicketStatus  = errors.New("invalid ticket status for this operation")
	ErrCannotDowngrade      = errors.New("cannot downgrade: new tier price must be higher than current tier price")
	ErrTicketDenied         = errors.New("cannot upgrade a denied ticket")
)

type TicketRepo struct {
	db *gorm.DB
}

func NewTicketRepo(db *gorm.DB) *TicketRepo {
	return &TicketRepo{db: db}
}

func (r *TicketRepo) GetUserTicket(ctx context.Context, userID uuid.UUID) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ? AND status != ?", userID, false, models.TicketStatusDenied).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) GetUserTicketByID(ctx context.Context, id uuid.UUID) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &t, nil
}

func getNextTicketNumber(ctx context.Context, tx *gorm.DB, tierID uuid.UUID) (int, error) {
	var maxNumber int
	err := tx.WithContext(ctx).
		Model(&models.UserTicket{}).
		Where("ticket_id = ?", tierID).
		Select("COALESCE(MAX(ticket_number), 0)").
		Scan(&maxNumber).Error
	if err != nil {
		return 0, err
	}
	return maxNumber + 1, nil
}

func (r *TicketRepo) PurchaseTicket(ctx context.Context, userID, tierID uuid.UUID) (*models.UserTicket, error) {
	var ticket *models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error; err != nil {
			return err
		}
		if user.IsBlacklisted {
			return ErrUserBlacklisted
		}
		var existing models.UserTicket
		if err := tx.Where("user_id = ? AND is_deleted = ? AND status != ?", userID, false, models.TicketStatusDenied).First(&existing).Error; err == nil {
			// Idempotent: already have a ticket for this tier -> success
			if existing.TicketId == tierID {
				ticket = &existing
				return nil
			}
			return ErrUserAlreadyHasTicket
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var tier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND is_deleted = ? AND is_active = ?", tierID, false, true).
			First(&tier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketTierNotFound
			}
			return err
		}
		if tier.Stock <= 0 {
			return ErrOutOfStock
		}
		num, err := getNextTicketNumber(ctx, tx, tierID)
		if err != nil {
			return err
		}
		if err := tx.Model(&tier).Update("stock", tier.Stock-1).Error; err != nil {
			return err
		}
		ref := fmt.Sprintf("%s-%04d", tier.TierCode, num)
		ticket = &models.UserTicket{
			Id:            uuid.New(),
			UserId:        userID,
			TicketId:      tierID,
			TicketNumber:  num,
			ReferenceCode: ref,
			Status:        models.TicketStatusPending,
		}
		return tx.Create(ticket).Error
	})
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (r *TicketRepo) ConfirmPayment(ctx context.Context, ticketID, userID uuid.UUID) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).
			First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		// Idempotent: already confirmed or approved -> success
		if t.Status == models.TicketStatusSelfConfirmed || t.Status == models.TicketStatusApproved {
			return nil
		}
		if t.Status != models.TicketStatusPending {
			return ErrInvalidTicketStatus
		}
		t.Status = models.TicketStatusSelfConfirmed
		return tx.Save(&t).Error
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) CancelTicket(ctx context.Context, ticketID, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var t models.UserTicket
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).
			First(&t).Error; err != nil {
			// Idempotent: already deleted/cancelled -> success
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if t.Status != models.TicketStatusPending && t.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}
		var tier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", t.TicketId).First(&tier).Error; err != nil {
			return err
		}
		if err := tx.Model(&tier).Update("stock", tier.Stock+1).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&t).Error
	})
	return err
}

func (r *TicketRepo) UpdateBadgeDetails(ctx context.Context, ticketID, userID uuid.UUID, badgeName, badgeImage string, isFursuiter, isFursuitStaff bool) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		if t.Status != models.TicketStatusApproved {
			return ErrInvalidTicketStatus
		}
		t.ConBadgeName = badgeName
		t.BadgeImage = badgeImage
		t.IsFursuiter = isFursuiter
		t.IsFursuitStaff = isFursuitStaff
		return tx.Save(&t).Error
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) ApproveTicket(ctx context.Context, ticketID, staffID uuid.UUID) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND is_deleted = ?", ticketID, false).
			First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		// Idempotent: already approved -> success
		if t.Status == models.TicketStatusApproved {
			return nil
		}
		if t.Status != models.TicketStatusPending && t.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}
		now := time.Now()
		t.Status = models.TicketStatusApproved
		t.ApprovedAt = &now
		t.ApprovedBy = &staffID
		return tx.Save(&t).Error
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) DenyTicket(ctx context.Context, ticketID, staffID uuid.UUID, reason string) (*models.UserTicket, error) {
	var t models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND is_deleted = ?", ticketID, false).
			First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		// Idempotent: already denied -> success (no stock/user update)
		if t.Status == models.TicketStatusDenied {
			return nil
		}
		if t.Status != models.TicketStatusPending && t.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}
		var tier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", t.TicketId).First(&tier).Error; err != nil {
			return err
		}
		if err := tx.Model(&tier).Update("stock", tier.Stock+1).Error; err != nil {
			return err
		}
		now := time.Now()
		t.Status = models.TicketStatusDenied
		t.DeniedAt = &now
		t.DeniedBy = &staffID
		t.DenialReason = reason
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
		var user models.User
		if err := tx.Where("id = ? AND is_deleted = ?", t.UserId, false).First(&user).Error; err != nil {
			return err
		}
		user.DenialCount++
		if user.DenialCount >= 3 {
			user.IsBlacklisted = true
			user.BlacklistedAt = &now
			user.BlacklistReason = "Automatically blacklisted after 3 ticket denials"
		}
		return tx.Model(&models.User{}).Where("id = ?", user.Id).Updates(map[string]interface{}{
			"denial_count":     user.DenialCount,
			"is_blacklisted":   user.IsBlacklisted,
			"blacklisted_at":   user.BlacklistedAt,
			"blacklist_reason": user.BlacklistReason,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) UpgradeTicketTier(ctx context.Context, userID, newTierID uuid.UUID) (*models.UserTicket, error) {
	var ticket models.UserTicket
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND is_deleted = ?", userID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}
		if ticket.Status == models.TicketStatusDenied {
			return ErrTicketDenied
		}
		var oldTier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", ticket.TicketId).First(&oldTier).Error; err != nil {
			return err
		}
		// Idempotent: already on this tier -> success
		if oldTier.Id == newTierID {
			return nil
		}
		var newTier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND is_deleted = ? AND is_active = ?", newTierID, false, true).
			First(&newTier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketTierNotFound
			}
			return err
		}
		if newTier.Stock <= 0 {
			return ErrOutOfStock
		}
		if newTier.Price.LessThanOrEqual(oldTier.Price) {
			return ErrCannotDowngrade
		}
		if err := tx.Model(&oldTier).Update("stock", oldTier.Stock+1).Error; err != nil {
			return err
		}
		if err := tx.Model(&newTier).Update("stock", newTier.Stock-1).Error; err != nil {
			return err
		}
		num, err := getNextTicketNumber(ctx, tx, newTierID)
		if err != nil {
			return err
		}
		ref := fmt.Sprintf("%s-%04d", newTier.TierCode, num)
		prevRef := ticket.ReferenceCode
		oldID := ticket.TicketId
		return tx.Model(&ticket).Updates(map[string]interface{}{
			"ticket_id":               newTierID,
			"ticket_number":           num,
			"reference_code":          ref,
			"status":                  models.TicketStatusPending,
			"upgraded_from_tier_id":   oldID,
			"previous_reference_code": prevRef,
			"approved_at":            nil,
			"approved_by":             nil,
			"denied_at":               nil,
			"denied_by":               nil,
			"denial_reason":           "",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepo) BlacklistUser(ctx context.Context, userID uuid.UUID, reason string) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = false", userID).
		Updates(map[string]interface{}{
			"is_blacklisted":   true,
			"blacklisted_at":   time.Now(),
			"blacklist_reason": reason,
		})
	if res.Error != nil {
		return res.Error
	}
	// 0 rows = user not found (invalid id or deleted). Already blacklisted still updates 1 row (idempotent).
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *TicketRepo) UnblacklistUser(ctx context.Context, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = false", userID).
		Updates(map[string]interface{}{
			"is_blacklisted":   false,
			"blacklisted_at":   nil,
			"blacklist_reason": "",
			"denial_count":     0,
		})
	if res.Error != nil {
		return res.Error
	}
	// 0 rows = user not found. Already not blacklisted still updates 1 row (idempotent).
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func isDuplicateKey(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505"))
}
