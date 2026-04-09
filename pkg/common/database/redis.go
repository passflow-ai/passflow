package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
	redisErr    error
)

// ConnectRedis establishes a connection to Redis.
func ConnectRedis(url string) error {
	redisOnce.Do(func() {
		opt, err := redis.ParseURL(url)
		if err != nil {
			redisErr = fmt.Errorf("failed to parse Redis URL: %w", err)
			return
		}

		opt.PoolSize = 10
		opt.MinIdleConns = 5
		opt.MaxRetries = 3
		opt.DialTimeout = 5 * time.Second
		opt.ReadTimeout = 3 * time.Second
		opt.WriteTimeout = 3 * time.Second

		client := redis.NewClient(opt)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			redisErr = fmt.Errorf("failed to ping Redis: %w", err)
			return
		}

		redisClient = client
		fmt.Println("Connected to Redis")
	})

	return redisErr
}

// GetRedisClient returns the Redis client.
func GetRedisClient() *redis.Client {
	return redisClient
}

// IsRedisConnected checks if Redis is connected.
func IsRedisConnected() bool {
	if redisClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return redisClient.Ping(ctx).Err() == nil
}

// RedisSet stores a value in Redis with an expiration time.
func RedisSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return redisClient.Set(ctx, key, value, expiration).Err()
}

// RedisGet retrieves a value from Redis.
func RedisGet(ctx context.Context, key string) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return redisClient.Get(ctx, key).Result()
}

// RedisDelete removes a key from Redis.
func RedisDelete(ctx context.Context, keys ...string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return redisClient.Del(ctx, keys...).Err()
}

// RedisExists checks if a key exists in Redis.
func RedisExists(ctx context.Context, key string) (bool, error) {
	if redisClient == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	n, err := redisClient.Exists(ctx, key).Result()
	return n > 0, err
}

// DisconnectRedis closes the Redis connection.
func DisconnectRedis() error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Close()
}
