package db

import (
	"context"
	"fmt"
	"time"

	"fuvekonse/sqs-worker/models"

	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration for all worker models. It creates missing tables
// and adds new columns/indexes when models change; it does not drop existing columns.
func AutoMigrate(gormDB *gorm.DB) error {
	return gormDB.AutoMigrate(
		&models.User{},
		&models.TicketTier{},
		&models.UserTicket{},
	)
}

// ValidateSchema runs minimal queries to ensure the tables and columns used by the worker exist.
// Call once after Connect; if general-service changes schema, this fails at startup with a clear error.
func ValidateSchema(ctx context.Context, gormDB *gorm.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Touch tables and columns we use (same as general-service: users, ticket_tiers, user_tickets).
	var n int64
	if err := gormDB.WithContext(ctx).Model(&models.UserTicket{}).Limit(1).Count(&n).Error; err != nil {
		return fmt.Errorf("schema check user_tickets: %w (ensure worker schema matches general-service)", err)
	}
	if err := gormDB.WithContext(ctx).Model(&models.TicketTier{}).Limit(1).Count(&n).Error; err != nil {
		return fmt.Errorf("schema check ticket_tiers: %w (ensure worker schema matches general-service)", err)
	}
	if err := gormDB.WithContext(ctx).Model(&models.User{}).Limit(1).Count(&n).Error; err != nil {
		return fmt.Errorf("schema check users: %w (ensure worker schema matches general-service)", err)
	}
	return nil
}
