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
)

// Note: TicketTier is now managed exclusively by ticket-service

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
	log.Println("✅ Database connection established")

	// Hash the password using centralized utility
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create test user with hardcoded ID
	hardcodedUserID := uuid.MustParse("b3c8e5f4-0a86-4d73-a8b9-3daafe0f6a20")
	testUser := models.User{
		Id:          hardcodedUserID,
		Email:       "admin@fuve.com",
		Password:    string(hashedPassword),
		FursonaName: "TestFursona",
		FirstName:   "Test",
		LastName:    "User",
		Country:     "Vietnam",
		Role:        constants.RoleUser,
		Avatar:      "https://via.placeholder.com/150",
		IsVerified:  true,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		IsDeleted:   false,
	}

	// Check if user already exists by ID
	var existingUser models.User
	result := db.Where("id = ?", hardcodedUserID).First(&existingUser)

	if result.Error == nil {
		log.Printf("⚠️  User with ID %s already exists. Updating...", hardcodedUserID)
		// Update existing user
		db.Model(&existingUser).Updates(map[string]interface{}{
			"password":     testUser.Password,
			"fursona_name": testUser.FursonaName,
			"first_name":   testUser.FirstName,
			"last_name":    testUser.LastName,
			"country":      testUser.Country,
			"role":         testUser.Role,
			"avatar":       testUser.Avatar,
			"is_verified":  testUser.IsVerified,
			"is_deleted":   testUser.IsDeleted,
			"modified_at":  time.Now(),
		})
		log.Println("✅ User updated successfully!")
	} else {
		// Create new user
		if err := db.Create(&testUser).Error; err != nil {
			log.Fatal("Failed to create test user:", err)
		}
		log.Println("✅ Test user created successfully!")
	}

	// Print user details
	log.Println("\n=== Test User Details ===")
	log.Printf("ID:          %s", testUser.Id)
	log.Printf("Email:       %s", testUser.Email)
	log.Printf("Password:    password123")
	log.Printf("Fursona:     %s", testUser.FursonaName)
	log.Printf("Name:        %s %s", testUser.FirstName, testUser.LastName)
	log.Printf("Role:        %s", testUser.Role)
	log.Printf("Is Verified: %v", testUser.IsVerified)
	log.Println("========================")

	// Note: Ticket tiers are now seeded exclusively by ticket-service

	log.Println("🎉 Seeding completed! You can now login with:")
	log.Println(`{
  "email": "user@example.com",
  "password": "password123"
}`)
}
