package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const loginFailedKeyPrefix = "login:failed:%s"

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

	// Increment the counter and set expiration
	pipe := redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	return err
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
		ttl := redisClient.TTL(ctx, key).Val()
		remainingMinutes := max(int(ttl.Minutes()), 0)
		return true, remainingMinutes, nil
	}

	return false, 0, nil
}
