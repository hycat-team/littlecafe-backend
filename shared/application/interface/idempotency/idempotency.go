package idempotency

import (
	"context"
	"time"
)

// State is the outcome of an idempotency begin attempt.
type State string

const (
	// StateNew means the key was claimed by this request.
	StateNew State = "NEW"
	// StateInFlight means another request with the same key is still running.
	StateInFlight State = "IN_FLIGHT"
	// StateCompleted means a stored response can be replayed.
	StateCompleted State = "COMPLETED"
	// StatePayloadMismatch means the key was reused with a different payload.
	StatePayloadMismatch State = "PAYLOAD_MISMATCH"
)

// StoredResponse is the exact response captured for replay.
type StoredResponse struct {
	Status      int
	Body        []byte
	ContentType string
}

// BeginResult tells the middleware what to do with the key.
type BeginResult struct {
	State    State
	Response *StoredResponse
}

// Store persists idempotency records.
type Store interface {
	// Begin atomically claims the key or returns its stored state.
	Begin(ctx context.Context, scope, key, requestHash string, ttl time.Duration) (BeginResult, error)
	// Complete stores the response for later replay.
	Complete(ctx context.Context, scope, key string, response StoredResponse, ttl time.Duration) error
	// Abort releases a claimed key that did not produce a cacheable response.
	Abort(ctx context.Context, scope, key string) error
}
