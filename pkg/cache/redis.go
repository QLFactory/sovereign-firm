package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides Redis caching operations
type Cache struct {
	client *redis.Client
	prefix string
}

// Config holds Redis configuration
type Config struct {
	URL      string
	Password string
	DB       int
	Prefix   string
}

// DefaultConfig returns default Redis configuration
func DefaultConfig() Config {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "localhost:6379"
	}
	return Config{
		URL:      url,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		Prefix:   "sf:",
	}
}

// New creates a new Redis cache client
func New(ctx context.Context, cfg Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.URL,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Cache{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	return c.client.Close()
}

// Client returns the underlying Redis client
func (c *Cache) Client() *redis.Client {
	return c.client
}

// key prefixes the key with the cache prefix
func (c *Cache) key(k string) string {
	return c.prefix + k
}

// Set stores a value with optional TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.client.Set(ctx, c.key(key), data, ttl).Err()
}

// Get retrieves a value
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.key(k)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

// Exists checks if a key exists
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.key(key)).Result()
	return n > 0, err
}

// SetNX sets a value only if it doesn't exist (for locks)
func (c *Cache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("marshal: %w", err)
	}
	return c.client.SetNX(ctx, c.key(key), data, ttl).Result()
}

// Incr increments a counter
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, c.key(key)).Result()
}

// Expire sets TTL on an existing key
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.key(key), ttl).Err()
}

// TTL gets remaining TTL on a key
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.key(key)).Result()
}

// Health checks Redis health
func (c *Cache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
