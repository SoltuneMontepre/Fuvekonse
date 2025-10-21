package main

import (
	"log"

	"github.com/SoltuneMontepre/Fuvekonse/services/general-service/internal/config"
	"github.com/SoltuneMontepre/Fuvekonse/services/general-service/internal/database"
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

	log.Println("🔄 Starting database migration...")

	if err := database.MigrateAndSeed(db); err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database migration completed successfully!")
}
