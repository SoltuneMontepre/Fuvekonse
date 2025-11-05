package main

import (
	"general-service/internal/common/constants"
	"general-service/internal/config"
	"general-service/internal/database"
	"general-service/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create test user
	testUser := models.User{
		Id:          uuid.New(),
		Email:       "user@example.com",
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

	// Check if user already exists
	var existingUser models.User
	result := db.Where("email = ?", testUser.Email).First(&existingUser)

	if result.Error == nil {
		log.Printf("⚠️  User with email %s already exists. Updating...", testUser.Email)
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
	log.Println("========================\n")

	log.Println("🎉 Seeding completed! You can now login with:")
	log.Println(`{
  "email": "user@example.com",
  "password": "password123"
}`)
}
