package main

import (
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/config"
	"general-service/internal/database"
	"general-service/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type seedUser struct {
	Email       string
	Password    string
	FursonaName string
	FirstName   string
	LastName    string
	Country     string
	Role        constants.UserRole
}

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Connect to database
	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Close DB connection when done
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	defer sqlDB.Close()

	log.Println("✅ Database connection established")

	// Define seed users
	seedUsers := []seedUser{
		{
			Email:       "admin@fuve.com",
			Password:    "admin123",
			FursonaName: "AdminFox",
			FirstName:   "Admin",
			LastName:    "FUVE",
			Country:     "Vietnam",
			Role:        constants.RoleAdmin,
		},
		{
			Email:       "user@fuve.com",
			Password:    "user123",
			FursonaName: "UserWolf",
			FirstName:   "User",
			LastName:    "Test",
			Country:     "Vietnam",
			Role:        constants.RoleUser,
		},
		{
			Email:       "dealer@fuve.com",
			Password:    "dealer123",
			FursonaName: "DealerCat",
			FirstName:   "Dealer",
			LastName:    "Test",
			Country:     "Vietnam",
			Role:        constants.RoleDealer,
		},
		// Keep legacy test user for backwards compatibility
		{
			Email:       "user@example.com",
			Password:    "password123",
			FursonaName: "TestFursona",
			FirstName:   "Test",
			LastName:    "User",
			Country:     "Vietnam",
			Role:        constants.RoleUser,
		},
	}

	log.Printf("\n🌱 Seeding %d users...", len(seedUsers))
	log.Println("========================")

	successCount := 0
	for i, su := range seedUsers {
		log.Printf("[%d/%d] Processing: %s", i+1, len(seedUsers), su.Email)
		if err := createOrUpdateUser(db, su); err != nil {
			log.Printf("❌ Failed to seed user %s: %v", su.Email, err)
		} else {
			successCount++
		}
	}

	// Verify by counting users in database
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("❌ Failed to count users in database: %v", err)
	} else {
		log.Printf("📊 Total users in database: %d", count)
	}

	log.Println("\n========================")
	log.Printf("🎉 Seeding completed! %d/%d users processed", successCount, len(seedUsers))
	log.Println("\n=== Seeded Accounts ===")
	log.Println(`
ADMIN:    admin@fuve.com
USER:     user@fuve.com
DEALER:   dealer@fuve.com
LEGACY:   user@example.com

For test credentials, see docs/test-credentials.md or check the seedUsers variable in cmd/seed/main.go
`)
}

func createOrUpdateUser(db *gorm.DB, su seedUser) error {
	// Hash the password
	hashedPassword, err := utils.HashPassword(su.Password)
	if err != nil {
		return err
	}

	user := models.User{
		Id:          uuid.New(),
		Email:       su.Email,
		Password:    string(hashedPassword),
		FursonaName: su.FursonaName,
		FirstName:   su.FirstName,
		LastName:    su.LastName,
		Country:     su.Country,
		Role:        su.Role,
		Avatar:      "https://via.placeholder.com/150",
		IsVerified:  true,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		IsDeleted:   false,
	}

	// Check if user already exists
	var existingUser models.User
	result := db.Where("email = ?", user.Email).First(&existingUser)

	if result.Error == nil {
		// Update existing user
		res := db.Model(&existingUser).Updates(map[string]interface{}{
			"password":     user.Password,
			"fursona_name": user.FursonaName,
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"country":      user.Country,
			"role":         user.Role,
			"avatar":       user.Avatar,
			"is_verified":  user.IsVerified,
			"is_deleted":   user.IsDeleted,
			"modified_at":  time.Now(),
		})
		if res.Error != nil {
			log.Printf("❌ Failed to update: %s (%s): %v", su.Email, su.Role.String(), res.Error)
			return res.Error
		}
		log.Printf("⚠️  Updated: %s (%s)", su.Email, su.Role.String())
	} else {
		// Create new user
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		log.Printf("✅ Created: %s (%s)", su.Email, su.Role.String())
	}

	return nil
}
