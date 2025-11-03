package database

import (
	"fmt"
	"general-service/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.DealerBooth{},
		&models.UserDealerStaff{},
		&models.TicketTier{},
		&models.UserTicket{},
		&models.ConBookArt{},
		&models.Payment{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate base tables: %w", err)
	}

	err = db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to auto-migrate dependent tables: %w", err)
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

func MigrateAndSeed(db *gorm.DB) error {
	if err := AutoMigrate(db); err != nil {
		return err
	}

	if err := seedInitialData(db); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

func seedInitialData(db *gorm.DB) error {

	fmt.Println("Initial data seeded successfully")
	return nil
}
