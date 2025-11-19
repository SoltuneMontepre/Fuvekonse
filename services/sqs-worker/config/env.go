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
		absPath, _ := filepath.Abs(envPath)
		err := godotenv.Load(absPath)
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

func IsLambdaEnv() bool {
	return GetEnvOr("AWS_LAMBDA_FUNCTION_NAME", "") != ""
}
