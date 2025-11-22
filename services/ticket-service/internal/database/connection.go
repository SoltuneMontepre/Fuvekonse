package database

import (
	"fmt"
	"time"

	"ticket-service/internal/config"
	"ticket-service/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(connectionString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return db, nil
}

func ConnectWithEnv() (*gorm.DB, error) {
	host := config.GetEnvOr("DB_HOST", "localhost")
	port := config.GetEnvOr("DB_PORT", "5432")
	user := config.GetEnvOr("DB_USER", "root")
	password := config.GetEnvOr("DB_PASSWORD", "root")
	dbname := config.GetEnvOr("DB_NAME", "fuvekon")
	sslmode := config.GetEnvOr("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return Connect(dsn)
}

func AutoMigrate(db *gorm.DB) error {
	// Migrate in dependency order to avoid foreign key constraint errors
	// 1. Base tables without foreign keys
	if err := db.AutoMigrate(&models.Permission{}); err != nil {
		return fmt.Errorf("failed to migrate Permission: %w", err)
	}
	if err := db.AutoMigrate(&models.Role{}); err != nil {
		return fmt.Errorf("failed to migrate Role: %w", err)
	}

	// 2. Tables that depend on Permission
	if err := db.AutoMigrate(&models.UserBan{}); err != nil {
		return fmt.Errorf("failed to migrate UserBan: %w", err)
	}

	// 3. Ticket-related tables
	if err := db.AutoMigrate(&models.TicketTier{}); err != nil {
		return fmt.Errorf("failed to migrate TicketTier: %w", err)
	}
	if err := db.AutoMigrate(&models.Ticket{}); err != nil {
		return fmt.Errorf("failed to migrate Ticket: %w", err)
	}
	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		return fmt.Errorf("failed to migrate Payment: %w", err)
	}

	return nil
}

func SeedInitialData(db *gorm.DB) error {
	// Seed ticket tiers
	ticketTiers := []models.TicketTier{
		{
			ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			TicketName:  "Tier 1",
			Description: "Basic tier ticket",
			Price:       1000000, // 1,000,000 VNĐ
			Stock:       100,
			IsActive:    true,
			BannerImage: "/images/ticket/tier1-banner.webp",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TicketName:  "Tier 2",
			Description: "Standard tier ticket",
			Price:       5000000, // 5,000,000 VNĐ
			Stock:       50,
			IsActive:    true,
			BannerImage: "/images/ticket/tier2-banner.webp",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			TicketName:  "Tier 3",
			Description: "Premium tier ticket",
			Price:       15000000, // 15,000,000 VNĐ
			Stock:       20,
			IsActive:    true,
			BannerImage: "/images/ticket/tier3-banner.webp",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, tier := range ticketTiers {
		var existingTier models.TicketTier
		result := db.Where("id = ?", tier.ID).First(&existingTier)

		if result.Error == nil {
			fmt.Printf("⚠️  Ticket tier with ID %s already exists. Skipping...\n", tier.ID)
			continue
		}

		// Create new tier
		if err := db.Create(&tier).Error; err != nil {
			return fmt.Errorf("failed to create ticket tier %s: %w", tier.TicketName, err)
		}
		fmt.Printf("✅ Ticket tier %s created successfully!\n", tier.TicketName)
	}

	fmt.Println("Initial ticket-service data seeded successfully")
	return nil
}
