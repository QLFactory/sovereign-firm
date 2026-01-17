package cache

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RateLimiter implements sliding window rate limiting using Redis
type RateLimiter struct {
	cache    *Cache
	requests int           // Max requests per window
	window   time.Duration // Window size
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache *Cache, requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		cache:    cache,
		requests: requests,
		window:   window,
	}
}

// rateLimitKey returns the cache key for rate limiting
func rateLimitKey(identifier string) string {
	return fmt.Sprintf("ratelimit:%s", identifier)
}

// Allow checks if a request is allowed and increments the counter
func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, int, error) {
	key := rateLimitKey(identifier)

	// Increment counter
	count, err := rl.cache.Incr(ctx, key)
	if err != nil {
		return false, 0, err
	}

	// Set expiry on first request in window
	if count == 1 {
		if err := rl.cache.Expire(ctx, key, rl.window); err != nil {
			return false, int(count), err
		}
	}

	remaining := rl.requests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64(rl.requests), remaining, nil
}

// Remaining returns remaining requests for the identifier
func (rl *RateLimiter) Remaining(ctx context.Context, identifier string) (int, error) {
	key := rateLimitKey(identifier)
	var count int64

	err := rl.cache.Get(ctx, key, &count)
	if err != nil {
		// Key doesn't exist = full quota available
		return rl.requests, nil
	}

	remaining := rl.requests - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// Reset resets the rate limit for an identifier
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	return rl.cache.Delete(ctx, rateLimitKey(identifier))
}

// Middleware returns an HTTP middleware that enforces rate limiting
func (rl *RateLimiter) Middleware(keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := keyFunc(r)

			allowed, remaining, err := rl.Allow(r.Context(), identifier)
			if err != nil {
				// On error, allow the request but log
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(rl.window.Seconds()), 10))
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ByIP returns a key function that rate limits by IP address
func ByIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return "ip:" + xff
	}
	return "ip:" + r.RemoteAddr
}

// ByTenant returns a key function that rate limits by tenant ID
func ByTenant(r *http.Request) string {
	// Get tenant from context (set by auth middleware)
	if tenantID := r.Context().Value("tenant_id"); tenantID != nil {
		return "tenant:" + tenantID.(string)
	}
	return ByIP(r)
}

// ByUser returns a key function that rate limits by user ID
func ByUser(r *http.Request) string {
	// Get user from context (set by auth middleware)
	if userID := r.Context().Value("user_id"); userID != nil {
		return "user:" + userID.(string)
	}
	return ByIP(r)
}
