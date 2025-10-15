package database

import (
	"fmt"

	"gin/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate base tables: %w", err)
	}

	err = db.AutoMigrate(
		&models.UserBan{},
	)
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
	var count int64
	db.Model(&models.Role{}).Count(&count)

	if count > 0 {
		fmt.Println("Initial data already exists, skipping seed")
		return nil
	}

	permissions := []models.Permission{
		{Name: "create_game_room"},
		{Name: "admin"},
		{Name: "ban_user"},
		{Name: "unban_user"},
		{Name: "view_banned_users"},
		{Name: "assign_roles"},
		{Name: "remove_roles"},
		{Name: "view_roles"},
		{Name: "create_permissions"},
		{Name: "delete_permissions"},
		{Name: "view_permissions"},
		{Name: "assign_permissions"},
		{Name: "remove_permissions"},
	}

	for _, perm := range permissions {
		if err := db.FirstOrCreate(&perm, models.Permission{Name: perm.Name}).Error; err != nil {
			return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
		}
	}

	roles := []models.Role{
		{Name: "player"},
		{Name: "admin"},
	}

	for _, role := range roles {
		if err := db.FirstOrCreate(&role, models.Role{Name: role.Name}).Error; err != nil {
			return fmt.Errorf("failed to create role %s: %w", role.Name, err)
		}
	}

	fmt.Println("Initial data seeded successfully")
	return nil
}