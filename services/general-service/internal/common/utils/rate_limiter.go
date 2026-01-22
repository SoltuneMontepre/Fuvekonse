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
func GetLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string) (int, error) {
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
func IncrementLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string, blockMinutes int) error {
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
func ResetLoginFailedAttempts(ctx context.Context, redisClient *redis.Client, email string) error {
	key := fmt.Sprintf(loginFailedKeyPrefix, email)
	return redisClient.Del(ctx, key).Err()
}

// IsLoginBlocked checks if a user is blocked due to too many failed login attempts
func IsLoginBlocked(ctx context.Context, redisClient *redis.Client, email string, maxFail int) (bool, int, error) {
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
func StoreOTP(ctx context.Context, redisClient *redis.Client, email, otp string, expiration time.Duration) error {
	key := fmt.Sprintf(otpKeyPrefix, email)
	return redisClient.Set(ctx, key, otp, expiration).Err()
}

// GetOTP retrieves the OTP for a given email from Redis
func GetOTP(ctx context.Context, redisClient *redis.Client, email string) (string, error) {
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
func VerifyAndDeleteOTP(ctx context.Context, redisClient *redis.Client, email, providedOTP string) (bool, error) {
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
func DeleteOTP(ctx context.Context, redisClient *redis.Client, email string) error {
	key := fmt.Sprintf(otpKeyPrefix, email)
	return redisClient.Del(ctx, key).Err()
}
