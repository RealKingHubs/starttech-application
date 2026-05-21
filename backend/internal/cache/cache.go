package cache

import (
	"context"
	"encoding/json"
	"log"
	// "strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
)

// Cache defines the interface for a caching service.
// This allows for mock/dummy implementations for testing or disabling caching.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetMany(ctx context.Context, data map[string]interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

// RedisCache is the Redis implementation of the Cache interface.
type RedisCache struct {
	client *redis.Client
}

// NewCacheService creates and returns a cache service based on the app config.
// It returns a real Redis client if caching is enabled, otherwise a no-op client.
func NewCacheService(cfg config.Config) Cache {
	if !cfg.EnableCache {
		log.Println("Caching is disabled.")
		return &NoOpCache{}
	}

	log.Println("Caching is enabled. Connecting to Redis...")
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0, // use default DB
	})

	// Ping Redis to check the connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v. Caching will be disabled.", err)
		// Fallback to no-op cache if connection fails
		return &NoOpCache{}
	}

	log.Println("Successfully connected to Redis.")
	return &RedisCache{client: rdb}
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err // redis.Nil if not found
	}
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, p, expiration).Err()
}

// SetMany stores multiple key-value pairs in the cache using a pipeline for efficiency.
func (r *RedisCache) SetMany(ctx context.Context, data map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()
	for key, value := range data {
		p, err := json.Marshal(value)
		if err != nil {
			// Log the error for the specific key but continue with the batch
			log.Printf("Error marshalling value for key %s: %v", key, err)
			continue
		}
		pipe.Set(ctx, key, p, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Ping checks the connection to Redis.
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// --- NoOpCache ---
// A dummy cache implementation that does nothing. Used when caching is disabled.

type NoOpCache struct{}

func (n *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Simulate a cache miss
	return redis.Nil
}

func (n *NoOpCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Do nothing
	return nil
}

func (n *NoOpCache) SetMany(ctx context.Context, data map[string]interface{}, expiration time.Duration) error {
	// Do nothing
	return nil
}

func (n *NoOpCache) Delete(ctx context.Context, key string) error {
	// Do nothing
	return nil
}

// Ping for NoOpCache always succeeds as there is no connection.
func (n *NoOpCache) Ping(ctx context.Context) error {
	return nil
}

