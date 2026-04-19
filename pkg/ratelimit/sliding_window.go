package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	redispkg "github.com/goconnect/pkg/redis"
)

// SlidingWindowLimiter implements Redis-based sliding window rate limiting
type SlidingWindowLimiter struct {
	redis interface {
		ZAdd(ctx context.Context, key string, members ...*redis.Z) error
		ZRemRangeByScore(ctx context.Context, key string, min, max string) error
		ZCard(ctx context.Context, key string) (int64, error)
		Expire(ctx context.Context, key string, expiration time.Duration) error
		Get(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, keys ...string) error
		Increment(ctx context.Context, key string) (int64, error)
		TTL(ctx context.Context, key string) (time.Duration, error)
	}
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(redisClient *redispkg.RedisClient) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		redis: redisClient,
	}
}

// CheckIPRegistrationLimit checks if an IP has exceeded registration rate limit
// Returns: allowed (bool), resetIn (time.Duration), error
func (s *SlidingWindowLimiter) CheckIPRegistrationLimit(ctx context.Context, ip string, maxAttempts int) (bool, time.Duration, error) {
	key := fmt.Sprintf("rate:register:ip:%s", ip)
	return s.checkLimit(ctx, key, maxAttempts, 1*time.Hour)
}

// CheckEmailRegistrationLimit checks if an email has exceeded OTP request rate limit
// Returns: allowed (bool), resetIn (time.Duration), error
func (s *SlidingWindowLimiter) CheckEmailRegistrationLimit(ctx context.Context, email string, maxAttempts int) (bool, time.Duration, error) {
	key := fmt.Sprintf("rate:register:email:%s", email)
	return s.checkLimit(ctx, key, maxAttempts, 1*time.Hour)
}

// CheckUsernameCheckLimit checks if an IP has exceeded username availability check rate limit
// Returns: allowed (bool), resetIn (time.Duration), error
func (s *SlidingWindowLimiter) CheckUsernameCheckLimit(ctx context.Context, ip string, maxAttempts int) (bool, time.Duration, error) {
	key := fmt.Sprintf("rate:username:ip:%s", ip)
	return s.checkLimit(ctx, key, maxAttempts, 1*time.Minute)
}

// IncrementCounter increments the counter for a given key with TTL
func (s *SlidingWindowLimiter) IncrementCounter(ctx context.Context, key string, window time.Duration) error {
	_, err := s.redis.Increment(ctx, key)
	if err != nil {
		return err
	}
	return s.redis.Expire(ctx, key, window)
}

// checkLimit is the internal method that checks rate limit for any key
func (s *SlidingWindowLimiter) checkLimit(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, time.Duration, error) {
	// Get current count
	valStr, err := s.redis.Get(ctx, key)
	count := 0
	if err == nil {
		count, _ = strconv.Atoi(valStr)
	}

	// If count is less than max, allow
	if count < maxAttempts {
		return true, 0, nil
	}

	// Exceeded limit, get TTL
	ttl, err := s.redis.TTL(ctx, key)
	if err != nil {
		return false, 0, err
	}

	return false, ttl, nil
}

// IncrementIPRegistration increments IP registration counter
func (s *SlidingWindowLimiter) IncrementIPRegistration(ctx context.Context, ip string) error {
	key := fmt.Sprintf("rate:register:ip:%s", ip)
	return s.IncrementCounter(ctx, key, 1*time.Hour)
}

// IncrementEmailRegistration increments email registration counter
func (s *SlidingWindowLimiter) IncrementEmailRegistration(ctx context.Context, email string) error {
	key := fmt.Sprintf("rate:register:email:%s", email)
	return s.IncrementCounter(ctx, key, 1*time.Hour)
}

// IncrementUsernameCheck increments username check counter
func (s *SlidingWindowLimiter) IncrementUsernameCheck(ctx context.Context, ip string) error {
	key := fmt.Sprintf("rate:username:ip:%s", ip)
	return s.IncrementCounter(ctx, key, 1*time.Minute)
}

// GetRemainingAttempts returns the number of remaining attempts for a given key
func (s *SlidingWindowLimiter) GetRemainingAttempts(ctx context.Context, key string, maxAttempts int) (int, error) {
	valStr, err := s.redis.Get(ctx, key)
	count := 0
	if err == nil {
		count, _ = strconv.Atoi(valStr)
	}

	remaining := maxAttempts - count
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// Reset clears the rate limit for a given key
func (s *SlidingWindowLimiter) Reset(ctx context.Context, key string) error {
	return s.redis.Delete(ctx, key)
}

// GetCount returns the current count for a given key
func (s *SlidingWindowLimiter) GetCount(ctx context.Context, key string) (int, error) {
	val, err := s.redis.Get(ctx, key)
	if err != nil {
		return 0, nil
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}
