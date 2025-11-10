package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return nil
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
