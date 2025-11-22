package database

import (
	"fmt"
	"general-service/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// Migrate in dependency order to avoid foreign key constraint errors
	// 1. Tables without foreign keys
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("failed to migrate User: %w", err)
	}
	if err := db.AutoMigrate(&models.DealerBooth{}); err != nil {
		return fmt.Errorf("failed to migrate DealerBooth: %w", err)
	}
	if err := db.AutoMigrate(&models.ConBookArt{}); err != nil {
		return fmt.Errorf("failed to migrate ConBookArt: %w", err)
	}

	// 2. Tables that depend on User
	if err := db.AutoMigrate(&models.UserDealerStaff{}); err != nil {
		return fmt.Errorf("failed to migrate UserDealerStaff: %w", err)
	}
	if err := db.AutoMigrate(&models.UserTicket{}); err != nil {
		return fmt.Errorf("failed to migrate UserTicket: %w", err)
	}

	// 3. Tables that depend on UserTicket
	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		return fmt.Errorf("failed to migrate Payment: %w", err)
	}

	// Note: UserTicket.TicketId now points to Ticket.Id from ticket-service
	// This creates a cross-service relationship handled via API calls

	fmt.Println("Database migration completed successfully")
	return nil
}

func SeedInitialData(db *gorm.DB) error {
	// Note: Ticket tiers are now managed exclusively by ticket-service
	// No seeding needed in general-service

	fmt.Println("Initial data seeded successfully")
	return nil
}
