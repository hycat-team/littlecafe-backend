package caching

import (
	"context"
	"fmt"
	"time"

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
	local allowed = 0
	local remaining = 0
	local reset_at = 0
	local retry_after = 0

	if now < allow_at then
		reset_at = math.ceil(allow_at)
		retry_after = math.ceil(allow_at - now)
	else
		allowed = 1
		local new_tat = math.max(now, tat) + emission
		local ttl = math.ceil(tau + emission)
		if ttl < 1 then
			ttl = 1
		end
		redis.call('SET', key, tostring(new_tat), 'EX', ttl)

		remaining = math.floor((now + tau - new_tat) / emission) + 1
		if remaining < 0 then
			remaining = 0
		end
		reset_at = math.ceil(math.max(now, new_tat - tau))
	end

	return {allowed, remaining, reset_at, retry_after}
`)

func (l *RedisTokenBucketLimiter) Allow(ctx context.Context, key string, burst int, refillPerSecond float64) (shared_caching.TokenBucketState, error) {
	if burst < 1 {
		return shared_caching.TokenBucketState{}, fmt.Errorf("token bucket: burst must be positive (got %d)", burst)
	}
	if refillPerSecond <= 0 {
		return shared_caching.TokenBucketState{}, fmt.Errorf("token bucket: refill rate must be positive (got %v)", refillPerSecond)
	}

	emission := 1 / refillPerSecond
	raw, err := tokenBucketScript.Run(ctx, l.client, []string{key}, emission, burst).Slice()
	if err != nil {
		return shared_caching.TokenBucketState{}, err
	}
	if len(raw) < 4 {
		return shared_caching.TokenBucketState{}, fmt.Errorf("token bucket: unexpected script result length %d", len(raw))
	}

	resetSeconds := luaInt64(raw[2])
	retrySeconds := luaInt64(raw[3])
	return shared_caching.TokenBucketState{
		Allowed:    luaInt64(raw[0]) == 1,
		Remaining:  int(luaInt64(raw[1])),
		Reset:      time.Unix(resetSeconds, 0),
		RetryAfter: time.Duration(retrySeconds) * time.Second,
	}, nil
}

func luaInt64(value any) int64 {
	switch n := value.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		var parsed int64
		_, _ = fmt.Sscan(n, &parsed)
		return parsed
	default:
		return 0
	}
}
