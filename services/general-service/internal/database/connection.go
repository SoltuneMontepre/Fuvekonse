package database

import (
	"fmt"
	"general-service/internal/config"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(connectionString string) (*gorm.DB, error) {
	// Ignore "record not found" in logs (expected when checking e.g. user has no ticket)
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	AutoGenerateUUID(db)

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
