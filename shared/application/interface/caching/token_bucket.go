package caching

import (
	"context"
	"time"
)

// TokenBucketState describes the outcome of a token-bucket decision.
type TokenBucketState struct {
	// Allowed reports whether the request may proceed.
	Allowed bool
	// Remaining is the number of tokens left after the decision.
	Remaining int
	// Reset is when the next token becomes available.
	Reset time.Time
	// RetryAfter is how long the caller must wait before retrying when denied.
	RetryAfter time.Duration
}

// ITokenBucketLimiter is the smooth (token-bucket / GCRA) limiter used for the
// coarse per-IP shield. It is deliberately separate from IRateLimiter
// (fixed window), which serves exact per-subject business quotas and login
// policies. Choose by semantics, not preference: flood shaping -> this port;
// "N times per period" -> IRateLimiter.
type ITokenBucketLimiter interface {
	// Allow consumes at most one token for key and returns the resulting state.
	// burst is the maximum bucket size and refillPerSecond is the sustained
	// refill rate.
	Allow(ctx context.Context, key string, burst int, refillPerSecond float64) (TokenBucketState, error)
}
