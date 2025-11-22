package main

import (
	"log"
	"ticket-service/internal/config"
	"ticket-service/internal/database"
)

func main() {
	config.LoadEnv()

	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	defer sqlDB.Close()

	log.Println("🌱 Starting ticket-service data seeding...")

	if err := database.SeedInitialData(db); err != nil {
		log.Fatal("❌ Failed to seed data:", err)
	}

	log.Println("✅ Data seeding completed successfully!")
}
