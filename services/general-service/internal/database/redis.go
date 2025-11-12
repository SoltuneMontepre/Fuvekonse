package database

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"general-service/internal/config"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

const errRedisClientNotInitialized = "redis client not initialized"

func ConnectRedis(opts *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return client, nil
}

func ConnectRedisWithEnv() (*redis.Client, error) {
	redisURL := config.GetEnvOr("REDIS_URL", "redis://localhost:6379/0")

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL value: %s: %w", redisURL, err)
	}

	// Enable TLS if running in AWS Lambda or if REDIS_TLS is explicitly set to "true"
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" || config.GetEnvOr("REDIS_TLS", "false") == "true" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return ConnectRedis(opts)
}

func InitRedis() error {
	client, err := ConnectRedisWithEnv()
	if err != nil {
		return err
	}

	RedisClient = client
	fmt.Println("Redis connection established successfully")
	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

func SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	if RedisClient == nil {
		return errors.New(errRedisClientNotInitialized)
	}
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", errors.New(errRedisClientNotInitialized)
	}
	return RedisClient.Get(ctx, key).Result()
}

func Delete(ctx context.Context, key string) error {
	if RedisClient == nil {
		return errors.New(errRedisClientNotInitialized)
	}
	return RedisClient.Del(ctx, key).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	if RedisClient == nil {
		return false, errors.New(errRedisClientNotInitialized)
	}
	result, err := RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}
