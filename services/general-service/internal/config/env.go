package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	// In AWS Lambda, environment variables are already set by the Lambda runtime
	// Check if we're running in Lambda by looking for AWS_LAMBDA_FUNCTION_NAME
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda - environment variables are already set, no need to load .env
		return nil
	}

	// Running locally - try to load .env file
	envPaths := []string{
		".env",
		"../../.env",
		"../../../.env",
		"../../../../.env",
	}

	var lastErr error
	for _, envPath := range envPaths {
		absPath, _ := filepath.Abs(envPath)
		err := godotenv.Load(absPath)
		if err == nil {
			// Successfully loaded
			fmt.Printf("Loaded .env from: %s\n", absPath)
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
