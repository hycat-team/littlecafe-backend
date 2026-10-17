package caching

import (
	"context"
	"fmt"

	shared_caching "github.com/hycat-team/littlecafe-backend/shared/application/interface/caching"

	"github.com/redis/go-redis/v9"
)

// RedisTokenBucketLimiter implements ITokenBucketLimiter with the GCRA
// algorithm in a single Lua script, so a decision is atomic and exact across
// every instance sharing the same Redis.
//
// State per key is the theoretical arrival time (TAT) in seconds. The script
// reads the clock from Redis TIME rather than the caller, so instances with
// skewed clocks agree on ordering.
type RedisTokenBucketLimiter struct {
	client *redis.Client
}

func NewRedisTokenBucketLimiter(client *redis.Client) shared_caching.ITokenBucketLimiter {
	return &RedisTokenBucketLimiter{client: client}
}

var tokenBucketScript = redis.NewScript(`
	local key = KEYS[1]
	local emission = tonumber(ARGV[1])   -- seconds per token
	local burst = tonumber(ARGV[2])      -- bucket size in tokens

	local t = redis.call('TIME')
	local now = tonumber(t[1]) + tonumber(t[2]) / 1000000

	local tat = redis.call('GET', key)
	if tat then
		tat = tonumber(tat)
	else
		tat = now
	end

	local tau = (burst - 1) * emission
	local allow_at = tat - tau
	if now < allow_at then
		return 0
	end

	local new_tat = math.max(now, tat) + emission
	local ttl = math.ceil(tau + emission)
	if ttl < 1 then
		ttl = 1
	end
	redis.call('SET', key, tostring(new_tat), 'EX', ttl)
	return 1
`)

func (l *RedisTokenBucketLimiter) Allow(ctx context.Context, key string, burst int, refillPerSecond float64) (bool, error) {
	if burst < 1 {
		return false, fmt.Errorf("token bucket: burst must be positive (got %d)", burst)
	}
	if refillPerSecond <= 0 {
		return false, fmt.Errorf("token bucket: refill rate must be positive (got %v)", refillPerSecond)
	}

	emission := 1 / refillPerSecond
	res, err := tokenBucketScript.Run(ctx, l.client, []string{key}, emission, burst).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
