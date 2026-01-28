package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	loginFailedKeyPrefix = "login:failed:%s"
	otpKeyPrefix         = "otp:%s"
)

// GetLoginFailedAttempts returns the number of failed login attempts for a given email
// Returns 0 if Redis is not available (graceful degradation)
func GetLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string) (int, error) {
	if redisClient == nil {
		// Redis not available, return 0 attempts (allow login)
		return 0, nil
	}
	
	key := fmt.Sprintf(loginFailedKeyPrefix, email)
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	attempts, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return attempts, nil
}

// IncrementLoginFailedAttempts increments the failed login attempts counter
// Returns nil if Redis is not available (graceful degradation)
func IncrementLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string, blockMinutes int) error {
	if redisClient == nil {
		// Redis not available, skip rate limiting
		return nil
	}
	
	key := fmt.Sprintf(loginFailedKeyPrefix, email)
	expiration := time.Duration(blockMinutes) * time.Minute

	// Try to set the key with expiration if it does not exist
	set, err := redisClient.SetNX(ctx, key, 1, expiration).Result()
	if err != nil {
		return err
	}
	if set {
		// Key was created, value is 1, expiration is set
		return nil
	}

	// Key exists, just increment (do not reset expiration)
	return redisClient.Incr(ctx, key).Err()
}

// ResetLoginFailedAttempts removes the failed login attempts counter
// Returns nil if Redis is not available (graceful degradation)
func ResetLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string) error {
	if redisClient == nil {
		// Redis not available, skip reset
		return nil
	}
	
	key := fmt.Sprintf(loginFailedKeyPrefix, email)
	return redisClient.Del(ctx, key).Err()
}

// IsLoginBlocked checks if a user is blocked due to too many failed login attempts
// Returns false (not blocked) if Redis is not available (graceful degradation)
func IsLoginBlocked(ctx context.Context, redisClient *redis.Client, email string, maxFail int) (bool, int, error) {
	if redisClient == nil {
		// Redis not available, allow login (no rate limiting)
		return false, 0, nil
	}
	
	attempts, err := GetLoginFailedAttempts(ctx, redisClient, email)
	if err != nil {
		return false, 0, err
	}

	if attempts >= maxFail {
		// Get remaining TTL
		key := fmt.Sprintf(loginFailedKeyPrefix, email)
		ttl, err := redisClient.TTL(ctx, key).Result()
		if err != nil {
			return false, 0, err
		}
		remainingMinutes := max(int(ttl.Minutes()), 0)
		return true, remainingMinutes, nil
	}

	return false, 0, nil
}

// StoreOTP stores an OTP in Redis with expiration
// Returns an error if Redis is not available (OTP storage is required)
func StoreOTP(ctx context.Context, redisClient *redis.Client, email, otp string, expiration time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not available: cannot store OTP")
	}
	
	key := fmt.Sprintf(otpKeyPrefix, email)
	return redisClient.Set(ctx, key, otp, expiration).Err()
}

// GetOTP retrieves the OTP for a given email from Redis
// Returns empty string and error if Redis is not available
func GetOTP(ctx context.Context, redisClient *redis.Client, email string) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not available: cannot retrieve OTP")
	}
	
	key := fmt.Sprintf(otpKeyPrefix, email)
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // OTP not found or expired
		}
		return "", err
	}
	return val, nil
}

// VerifyAndDeleteOTP verifies the OTP and deletes it from Redis
// Returns false and error if Redis is not available
func VerifyAndDeleteOTP(ctx context.Context, redisClient *redis.Client, email, providedOTP string) (bool, error) {
	if redisClient == nil {
		return false, fmt.Errorf("redis client not available: cannot verify OTP")
	}
	
	storedOTP, err := GetOTP(ctx, redisClient, email)
	if err != nil {
		return false, err
	}

	if storedOTP == "" {
		return false, nil // OTP not found or expired
	}

	if storedOTP != providedOTP {
		return false, nil // OTP mismatch
	}

	// OTP matches, delete it
	key := fmt.Sprintf(otpKeyPrefix, email)
	return true, redisClient.Del(ctx, key).Err()
}

// DeleteOTP removes the OTP for a given email
// Returns nil if Redis is not available (graceful degradation)
func DeleteOTP(ctx context.Context, redisClient *redis.Client, email string) error {
	if redisClient == nil {
		// Redis not available, skip deletion
		return nil
	}
	
	key := fmt.Sprintf(otpKeyPrefix, email)
	return redisClient.Del(ctx, key).Err()
}
