// Package config provides utilities for loading environment variables and environment helpers
// used by the sqs-worker service.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	envPaths := []string{
		".env",
		"../../.env",
		"../.env",
	}

	var lastErr error
	for _, envPath := range envPaths {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			lastErr = fmt.Errorf("failed to resolve absolute path for %s: %w", envPath, err)
			continue
		}
		err = godotenv.Load(absPath)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("error loading .env file from any location: %w", lastErr)
}

func GetEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RequiredDBVars is the list of env vars required for the worker to connect to the database.
// Use the same values as general-service (same DB).
var RequiredDBVars = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}

// ValidateDBEnv returns an error if any required DB_* env var is missing.
// Call at startup (Lambda or local) to fail fast with a clear message.
func ValidateDBEnv() error {
	var missing []string
	for _, key := range RequiredDBVars {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required DB env (use same DB as general-service): %v", missing)
	}
	return nil
}

// DatabaseDSN returns a PostgreSQL DSN from DB_* env vars (same as general-service).
func DatabaseDSN() string {
	host := GetEnvOr("DB_HOST", "localhost")
	port := GetEnvOr("DB_PORT", "5432")
	user := GetEnvOr("DB_USER", "root")
	password := GetEnvOr("DB_PASSWORD", "root")
	dbname := GetEnvOr("DB_NAME", "fuvekon")
	sslmode := GetEnvOr("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func IsLambdaEnv() bool {
	return GetEnvOr("AWS_LAMBDA_FUNCTION_NAME", "") != ""
}
