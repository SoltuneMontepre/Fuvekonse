package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	// Try to load .env from multiple locations
	// 1. Current directory (services/general-service/.env)
	// 2. Two levels up (project root: ../../.env)
	envPaths := []string{
		".env",
		"../../.env",
	}

	var lastErr error
	for _, envPath := range envPaths {
		absPath, _ := filepath.Abs(envPath)
		err := godotenv.Load(absPath)
		if err == nil {
			// Successfully loaded
			return nil
		}
		lastErr = err
	}

	// If all paths failed, return the last error
	return fmt.Errorf("error loading .env file from any location: %w", lastErr)
}

func GetEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetLoginMaxFail() int {
	value := GetEnvOr("LOGIN_MAX_FAIL", "5")
	maxFail, err := strconv.Atoi(value)
	if err != nil {
		return 5 // default to 5 if parsing fails
	}
	return maxFail
}

func GetLoginFailBlockMinutes() int {
	value := GetEnvOr("LOGIN_FAIL_BLOCK_MINUTES", "15")
	blockMinutes, err := strconv.Atoi(value)
	if err != nil {
		return 15 // default to 15 if parsing fails
	}
	return blockMinutes
}
