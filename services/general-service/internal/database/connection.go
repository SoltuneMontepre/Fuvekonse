package database

import (
	"fmt"
	"gin/internal/config"

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
	user := config.GetEnvOr("DB_USER", "admin")
	password := config.GetEnvOr("DB_PASSWORD", "password")
	dbname := config.GetEnvOr("DB_NAME", "authorizationdb")
	sslmode := config.GetEnvOr("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return Connect(dsn)
}
