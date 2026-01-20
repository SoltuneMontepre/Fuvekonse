package config

import (
	"fmt"
	"os"

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
