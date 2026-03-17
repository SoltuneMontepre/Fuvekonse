package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	loginFailedKeyPrefix       = "login:failed:%s"
	otpKeyPrefix               = "otp:%s"
	otpAttemptKeyPrefix        = "otp:attempts:%s"
	passwordResetJTIKeyPrefix  = "pwd_reset_jti:%s"
	otpMaxAttempts             = 5
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

// IncrementOTPAttempts increments the OTP verification attempt counter for an email.
// Returns the new attempt count. The counter TTL matches the OTP TTL so it auto-cleans.
func IncrementOTPAttempts(ctx context.Context, redisClient *redis.Client, email string, otpTTL time.Duration) (int, error) {
	if redisClient == nil {
		return 0, nil
	}
	key := fmt.Sprintf(otpAttemptKeyPrefix, email)
	set, err := redisClient.SetNX(ctx, key, 1, otpTTL).Result()
	if err != nil {
		return 0, err
	}
	if set {
		return 1, nil
	}
	count, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// IsOTPBlocked returns true if the email has exhausted OTP verification attempts.
func IsOTPBlocked(ctx context.Context, redisClient *redis.Client, email string) (bool, error) {
	if redisClient == nil {
		return false, nil
	}
	key := fmt.Sprintf(otpAttemptKeyPrefix, email)
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	attempts, err := strconv.Atoi(val)
	if err != nil {
		return false, err
	}
	return attempts >= otpMaxAttempts, nil
}

// ResetOTPAttempts removes the OTP attempt counter (called on successful verification or new OTP).
func ResetOTPAttempts(ctx context.Context, redisClient *redis.Client, email string) error {
	if redisClient == nil {
		return nil
	}
	key := fmt.Sprintf(otpAttemptKeyPrefix, email)
	return redisClient.Del(ctx, key).Err()
}

// StorePasswordResetJTI marks a password-reset token's JTI as valid in Redis.
// The TTL should match the JWT expiry so the key auto-cleans after the token expires.
func StorePasswordResetJTI(ctx context.Context, redisClient *redis.Client, jti string, expiration time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not available: cannot store password reset JTI")
	}
	key := fmt.Sprintf(passwordResetJTIKeyPrefix, jti)
	return redisClient.Set(ctx, key, "1", expiration).Err()
}

// ConsumePasswordResetJTI atomically checks and deletes a password-reset JTI.
// Returns true if the JTI existed (token is valid for single use), false otherwise.
func ConsumePasswordResetJTI(ctx context.Context, redisClient *redis.Client, jti string) (bool, error) {
	if redisClient == nil {
		return false, fmt.Errorf("redis client not available: cannot consume password reset JTI")
	}
	key := fmt.Sprintf(passwordResetJTIKeyPrefix, jti)
	deleted, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
}
