package caching

import "context"

// ITokenBucketLimiter is the smooth (token-bucket / GCRA) limiter used for the
// coarse per-IP shield. It is deliberately separate from IRateLimiter
// (fixed window), which serves exact per-subject business quotas and login
// policies. Choose by semantics, not preference: flood shaping → this port;
// "N times per period" → IRateLimiter.
type ITokenBucketLimiter interface {
	// Allow reports whether one token is available for key. burst is the
	// maximum bucket size and refillPerSecond is the sustained refill rate.
	Allow(ctx context.Context, key string, burst int, refillPerSecond float64) (bool, error)
}
