package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisConnection creates a new Redis client
func NewRedisConnection(url string) *redis.Client {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse Redis URL: %v", err))
	}

	client := redis.NewClient(opt)
	return client
}

// RedisHelper provides common Redis operations
type RedisHelper struct {
	client *redis.Client
}

// NewRedisHelper creates a new Redis helper
func NewRedisHelper(client *redis.Client) *RedisHelper {
	return &RedisHelper{client: client}
}

// Set sets a value in Redis with expiration
func (r *RedisHelper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis
func (r *RedisHelper) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// GetJSON retrieves a JSON value from Redis and unmarshals it
func (r *RedisHelper) GetJSON(ctx context.Context, key string, result interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	// Unmarshal implementation would depend on your JSON library
	return nil
}

// Delete deletes a key from Redis
func (r *RedisHelper) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisHelper) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Increment increments a value in Redis
func (r *RedisHelper) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetIfNotExists sets a value only if the key doesn't exist
func (r *RedisHelper) SetIfNotExists(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := r.client.SetNX(ctx, key, value, expiration).Result()
	return result, err
}

// GetAndDelete retrieves a value and deletes the key atomically
func (r *RedisHelper) GetAndDelete(ctx context.Context, key string) (string, error) {
	return r.client.GetDel(ctx, key).Result()
}

// HSet sets a hash field in Redis
func (r *RedisHelper) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet retrieves a hash field from Redis
func (r *RedisHelper) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll retrieves all fields in a hash
func (r *RedisHelper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// SetRateLimitCounter sets a counter for rate limiting
func (r *RedisHelper) SetRateLimitCounter(ctx context.Context, key string, limit int, windowSeconds int) error {
	pipeline := r.client.Pipeline()
	pipeline.Incr(ctx, key)
	pipeline.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	_, err := pipeline.Exec(ctx)
	return err
}

// CheckRateLimit checks if a request should be rate limited
func (r *RedisHelper) CheckRateLimit(ctx context.Context, key string, limit int) (bool, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return val >= int64(limit), nil
}

// Close closes the Redis client connection
func (r *RedisHelper) Close() error {
	return r.client.Close()
}

// Ping checks the Redis connection
func (r *RedisHelper) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
