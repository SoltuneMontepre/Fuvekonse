package database

import (
	"context"
	"fmt"
	"gin/internal/config"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

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
	host := config.GetEnvOr("REDIS_HOST", "localhost")
	port := config.GetEnvOr("REDIS_PORT", "6379")
	password := config.GetEnvOr("REDIS_PASSWORD", "")
	dbStr := config.GetEnvOr("REDIS_DB", "0")
	
	db, err := strconv.Atoi(dbStr)

	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB value: %s", dbStr)
	}
	
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
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

func SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return RedisClient.Get(ctx, key).Result()
}

func Delete(ctx context.Context, key string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return RedisClient.Del(ctx, key).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	result, err := RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}