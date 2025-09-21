package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient          *redis.Client
	limit                int
	window               time.Duration
	ctx                  context.Context
	currentNumberRequest int
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration, ctx context.Context) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
		ctx:         ctx,
	}
}

func (r *RateLimiter) IsAllow(key string) (bool, error) {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	script := redis.NewScript(`
	local current
	current = redis.call("INCR", KEYS[1])
	if tonumber(current) == 1	then 
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end
	return current
	`)
	result, err := script.Run(r.ctx, r.redisClient, []string{redisKey}, int(r.window.Milliseconds())).Result()
	if err != nil {
		return false, err
	}
	count := int(result.(int64))
	r.currentNumberRequest = count
	if count > r.limit {
		return false, nil
	}
	return true, nil
}

func (r *RateLimiter) GetLimit() int {
	return r.limit
}

func (r *RateLimiter) GetCurrentNumberRequest() int {
	return r.currentNumberRequest
}
