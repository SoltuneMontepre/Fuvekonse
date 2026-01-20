package repositories

import (
	"context"
	"errors"
	"fmt"
	"general-service/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTicketTierNotFound     = errors.New("ticket tier not found")
	ErrTicketNotFound         = errors.New("ticket not found")
	ErrOutOfStock             = errors.New("ticket tier is out of stock")
	ErrUserAlreadyHasTicket   = errors.New("user already has a ticket")
	ErrUserBlacklisted        = errors.New("user is blacklisted from purchasing tickets")
	ErrInvalidTicketStatus    = errors.New("invalid ticket status for this operation")
)

type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// ========== Ticket Tier Operations ==========

// GetAllActiveTiers returns all active (non-deleted) ticket tiers
func (r *TicketRepository) GetAllActiveTiers(ctx context.Context) ([]models.TicketTier, error) {
	var tiers []models.TicketTier
	err := r.db.WithContext(ctx).
		Where("is_deleted = ? AND is_active = ?", false, true).
		Order("price ASC").
		Find(&tiers).Error
	if err != nil {
		return nil, err
	}
	return tiers, nil
}

// GetTierByID returns a ticket tier by ID
func (r *TicketRepository) GetTierByID(ctx context.Context, id uuid.UUID) (*models.TicketTier, error) {
	var tier models.TicketTier
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&tier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketTierNotFound
		}
		return nil, err
	}
	return &tier, nil
}

// GetTierByCode returns a ticket tier by tier code
func (r *TicketRepository) GetTierByCode(ctx context.Context, tierCode string) (*models.TicketTier, error) {
	var tier models.TicketTier
	err := r.db.WithContext(ctx).
		Where("tier_code = ? AND is_deleted = ?", tierCode, false).
		First(&tier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketTierNotFound
		}
		return nil, err
	}
	return &tier, nil
}

// ========== User Ticket Operations ==========

// GetUserTicket returns the user's current ticket (if any)
// Excludes denied tickets since they are no longer considered "active"
func (r *TicketRepository) GetUserTicket(ctx context.Context, userID uuid.UUID) (*models.UserTicket, error) {
	var ticket models.UserTicket
	err := r.db.WithContext(ctx).
		Preload("Ticket").
		Preload("User").
		Where("user_id = ? AND is_deleted = ? AND status != ?", userID, false, models.TicketStatusDenied).
		First(&ticket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No ticket found - this is valid
		}
		return nil, err
	}
	return &ticket, nil
}

// GetUserTicketByID returns a ticket by ID
func (r *TicketRepository) GetUserTicketByID(ctx context.Context, id uuid.UUID) (*models.UserTicket, error) {
	var ticket models.UserTicket
	err := r.db.WithContext(ctx).
		Preload("Ticket").
		Preload("User").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&ticket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

// GetUserTicketByReference returns a ticket by reference code
func (r *TicketRepository) GetUserTicketByReference(ctx context.Context, referenceCode string) (*models.UserTicket, error) {
	var ticket models.UserTicket
	err := r.db.WithContext(ctx).
		Preload("Ticket").
		Preload("User").
		Where("reference_code = ? AND is_deleted = ?", referenceCode, false).
		First(&ticket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

// GetNextTicketNumber returns the next ticket number for a tier (thread-safe within transaction)
func (r *TicketRepository) GetNextTicketNumber(ctx context.Context, tx *gorm.DB, tierID uuid.UUID) (int, error) {
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

// PurchaseTicket creates a new ticket with stock decrement (atomic operation)
// This uses row-level locking to prevent race conditions
func (r *TicketRepository) PurchaseTicket(ctx context.Context, userID, tierID uuid.UUID) (*models.UserTicket, error) {
	var ticket *models.UserTicket

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Check if user is blacklisted
		var user models.User
		if err := tx.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error; err != nil {
			return err
		}
		if user.IsBlacklisted {
			return ErrUserBlacklisted
		}

		// 2. Check if user already has a non-denied ticket
		var existingTicket models.UserTicket
		err := tx.Where("user_id = ? AND is_deleted = ? AND status != ?", userID, false, models.TicketStatusDenied).
			First(&existingTicket).Error
		if err == nil {
			return ErrUserAlreadyHasTicket
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 3. Lock the tier row for update and check stock
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

		// 4. Get next ticket number for this tier
		ticketNumber, err := r.GetNextTicketNumber(ctx, tx, tierID)
		if err != nil {
			return err
		}

		// 5. Decrement stock
		if err := tx.Model(&tier).Update("stock", tier.Stock-1).Error; err != nil {
			return err
		}

		// 6. Create the ticket
		referenceCode := fmt.Sprintf("%s-%04d", tier.TierCode, ticketNumber)
		ticket = &models.UserTicket{
			Id:            uuid.New(),
			UserId:        userID,
			TicketId:      tierID,
			TicketNumber:  ticketNumber,
			ReferenceCode: referenceCode,
			Status:        models.TicketStatusPending,
		}

		if err := tx.Create(ticket).Error; err != nil {
			return err
		}

		// 7. Load related data
		if err := tx.Preload("Ticket").Preload("User").First(ticket, "id = ?", ticket.Id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// ConfirmPayment updates ticket status to self_confirmed
func (r *TicketRepository) ConfirmPayment(ctx context.Context, ticketID, userID uuid.UUID) (*models.UserTicket, error) {
	var ticket models.UserTicket

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find and lock the ticket
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		// Verify status
		if ticket.Status != models.TicketStatusPending {
			return ErrInvalidTicketStatus
		}

		// Update status
		ticket.Status = models.TicketStatusSelfConfirmed
		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load related data
	if err := r.db.WithContext(ctx).Preload("Ticket").Preload("User").First(&ticket, "id = ?", ticket.Id).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

// ApproveTicket approves a ticket (staff action)
func (r *TicketRepository) ApproveTicket(ctx context.Context, ticketID, staffID uuid.UUID) (*models.UserTicket, error) {
	var ticket models.UserTicket

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find and lock the ticket
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("User").
			Where("id = ? AND is_deleted = ?", ticketID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		// Verify status - can only approve pending or self_confirmed
		if ticket.Status != models.TicketStatusPending && ticket.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}

		// Update ticket
		now := time.Now()
		ticket.Status = models.TicketStatusApproved
		ticket.ApprovedAt = &now
		ticket.ApprovedBy = &staffID

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load related data
	if err := r.db.WithContext(ctx).Preload("Ticket").Preload("User").First(&ticket, "id = ?", ticket.Id).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

// DenyTicket denies a ticket and re-increments stock (staff action)
func (r *TicketRepository) DenyTicket(ctx context.Context, ticketID, staffID uuid.UUID, reason string) (*models.UserTicket, error) {
	var ticket models.UserTicket

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find and lock the ticket
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("User").
			Where("id = ? AND is_deleted = ?", ticketID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		// Verify status - can only deny pending or self_confirmed
		if ticket.Status != models.TicketStatusPending && ticket.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}

		// Lock and update the tier (re-increment stock)
		var tier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", ticket.TicketId).
			First(&tier).Error; err != nil {
			return err
		}
		if err := tx.Model(&tier).Update("stock", tier.Stock+1).Error; err != nil {
			return err
		}

		// Update ticket
		now := time.Now()
		ticket.Status = models.TicketStatusDenied
		ticket.DeniedAt = &now
		ticket.DeniedBy = &staffID
		ticket.DenialReason = reason

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		// Update user's denial count and check for blacklist (only for non-deleted users)
		var user models.User
		if err := tx.Where("id = ? AND is_deleted = ?", ticket.UserId, false).First(&user).Error; err != nil {
			return err
		}

		user.DenialCount++
		if user.DenialCount >= 3 {
			user.IsBlacklisted = true
			user.BlacklistedAt = &now
			user.BlacklistReason = "Automatically blacklisted after 3 ticket denials"
		}

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load related data
	if err := r.db.WithContext(ctx).Preload("Ticket").Preload("User").First(&ticket, "id = ?", ticket.Id).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

// CancelTicket cancels a pending ticket and re-increments stock
func (r *TicketRepository) CancelTicket(ctx context.Context, ticketID, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find and lock the ticket
		var ticket models.UserTicket
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		// Can only cancel pending or self_confirmed tickets
		if ticket.Status != models.TicketStatusPending && ticket.Status != models.TicketStatusSelfConfirmed {
			return ErrInvalidTicketStatus
		}

		// Lock and update the tier (re-increment stock)
		var tier models.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", ticket.TicketId).
			First(&tier).Error; err != nil {
			return err
		}
		if err := tx.Model(&tier).Update("stock", tier.Stock+1).Error; err != nil {
			return err
		}

		// Soft delete the ticket
		now := time.Now()
		ticket.IsDeleted = true
		ticket.DeletedAt = &now

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// UpdateBadgeDetails updates badge details after approval
func (r *TicketRepository) UpdateBadgeDetails(ctx context.Context, ticketID, userID uuid.UUID, badgeName, badgeImage string, isFursuiter, isFursuitStaff bool) (*models.UserTicket, error) {
	var ticket models.UserTicket

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find the ticket
		if err := tx.Where("id = ? AND user_id = ? AND is_deleted = ?", ticketID, userID, false).
			First(&ticket).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTicketNotFound
			}
			return err
		}

		// Verify status - can only update badge details after approval
		if ticket.Status != models.TicketStatusApproved {
			return ErrInvalidTicketStatus
		}

		// Update badge details
		ticket.ConBadgeName = badgeName
		ticket.BadgeImage = badgeImage
		ticket.IsFursuiter = isFursuiter
		ticket.IsFursuitStaff = isFursuitStaff

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load related data
	if err := r.db.WithContext(ctx).Preload("Ticket").Preload("User").First(&ticket, "id = ?", ticket.Id).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

// ========== Admin Operations ==========

// AdminTicketFilter defines filters for admin ticket listing
type AdminTicketFilter struct {
	Status        *models.TicketStatus
	TierID        *uuid.UUID
	Search        string // Search by reference code, user name, or email
	PendingOver24 bool   // Only show tickets pending > 24 hours
	Page          int
	PageSize      int
}

// GetTicketsForAdmin returns tickets with filters for admin view
func (r *TicketRepository) GetTicketsForAdmin(ctx context.Context, filter AdminTicketFilter) ([]models.UserTicket, int64, error) {
	var tickets []models.UserTicket
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.UserTicket{}).
		Preload("Ticket").
		Preload("User").
		Where("user_tickets.is_deleted = ?", false)

	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.TierID != nil {
		query = query.Where("ticket_id = ?", *filter.TierID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Joins("LEFT JOIN users ON users.id = user_tickets.user_id AND users.is_deleted = false").
			Where("reference_code ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ? OR users.fursona_name ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if filter.PendingOver24 {
		twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
		query = query.Where("created_at < ? AND status IN (?, ?)",
			twentyFourHoursAgo, models.TicketStatusPending, models.TicketStatusSelfConfirmed)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Fetch with ordering (self_confirmed first, then by created_at)
	if err := query.
		Order("CASE WHEN status = 'self_confirmed' THEN 0 WHEN status = 'pending' THEN 1 ELSE 2 END").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// GetTicketStatistics returns ticket statistics for admin dashboard
type TicketStatistics struct {
	TotalTickets        int64
	PendingCount        int64
	SelfConfirmedCount  int64
	ApprovedCount       int64
	DeniedCount         int64
	PendingOver24Hours  int64
	TierStats           []TierStatistics
}

type TierStatistics struct {
	TierID      uuid.UUID
	TierCode    string
	TierName    string
	TotalStock  int
	Sold        int64
	Available   int
}

func (r *TicketRepository) GetTicketStatistics(ctx context.Context) (*TicketStatistics, error) {
	stats := &TicketStatistics{}

	// Get overall counts
	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ?", false).
		Count(&stats.TotalTickets)

	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ? AND status = ?", false, models.TicketStatusPending).
		Count(&stats.PendingCount)

	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ? AND status = ?", false, models.TicketStatusSelfConfirmed).
		Count(&stats.SelfConfirmedCount)

	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ? AND status = ?", false, models.TicketStatusApproved).
		Count(&stats.ApprovedCount)

	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ? AND status = ?", false, models.TicketStatusDenied).
		Count(&stats.DeniedCount)

	// Pending over 24 hours
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.UserTicket{}).
		Where("is_deleted = ? AND created_at < ? AND status IN (?, ?)",
			false, twentyFourHoursAgo, models.TicketStatusPending, models.TicketStatusSelfConfirmed).
		Count(&stats.PendingOver24Hours)

	// Get tier statistics
	var tiers []models.TicketTier
	if err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Order("price ASC").
		Find(&tiers).Error; err != nil {
		return nil, err
	}

	for _, tier := range tiers {
		var sold int64
		r.db.WithContext(ctx).Model(&models.UserTicket{}).
			Where("ticket_id = ? AND is_deleted = ? AND status != ?", tier.Id, false, models.TicketStatusDenied).
			Count(&sold)

		stats.TierStats = append(stats.TierStats, TierStatistics{
			TierID:     tier.Id,
			TierCode:   tier.TierCode,
			TierName:   tier.TicketName,
			TotalStock: tier.Stock + int(sold), // Original stock = current + sold
			Sold:       sold,
			Available:  tier.Stock,
		})
	}

	return stats, nil
}

// ========== Blacklist Operations ==========

// GetBlacklistedUsers returns all blacklisted users
func (r *TicketRepository) GetBlacklistedUsers(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("is_blacklisted = ? AND is_deleted = ?", true, false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.
		Order("blacklisted_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// BlacklistUser manually blacklists a user (excludes soft-deleted users)
func (r *TicketRepository) BlacklistUser(ctx context.Context, userID uuid.UUID, reason string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = false", userID).
		Updates(map[string]interface{}{
			"is_blacklisted":   true,
			"blacklisted_at":   now,
			"blacklist_reason": reason,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UnblacklistUser removes a user from blacklist (excludes soft-deleted users)
func (r *TicketRepository) UnblacklistUser(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = false", userID).
		Updates(map[string]interface{}{
			"is_blacklisted":   false,
			"blacklisted_at":   nil,
			"blacklist_reason": "",
			"denial_count":     0, // Reset denial count
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
