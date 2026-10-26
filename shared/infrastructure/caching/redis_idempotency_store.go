package caching

import (
	"context"
	"fmt"
	"strconv"
	"time"

	shared_idempotency "github.com/hycat-team/littlecafe-backend/shared/application/interface/idempotency"

	"github.com/redis/go-redis/v9"
)

// RedisIdempotencyStore persists idempotency records in Redis.
type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(client *redis.Client) shared_idempotency.Store {
	return &RedisIdempotencyStore{client: client}
}

var idempotencyBeginScript = redis.NewScript(`
	local key = KEYS[1]
	local requestHash = ARGV[1]
	local ttl = tonumber(ARGV[2])

	local state = redis.call('HGET', key, 'state')
	if not state then
		redis.call('HSET', key, 'state', 'IN_FLIGHT', 'hash', requestHash)
		redis.call('EXPIRE', key, ttl)
		return {'NEW'}
	end

	local storedHash = redis.call('HGET', key, 'hash')
	if storedHash ~= requestHash then
		return {'PAYLOAD_MISMATCH'}
	end
	if state == 'IN_FLIGHT' then
		return {'IN_FLIGHT'}
	end

	local status = redis.call('HGET', key, 'status')
	local body = redis.call('HGET', key, 'body')
	local contentType = redis.call('HGET', key, 'contentType')
	return {'COMPLETED', status or '', body or '', contentType or ''}
`)

func idempotencyKey(scope, key string) string {
	return "idempotency:" + scope + ":" + key
}

func (s *RedisIdempotencyStore) Begin(ctx context.Context, scope, key, requestHash string, ttl time.Duration) (shared_idempotency.BeginResult, error) {
	if ttl <= 0 {
		return shared_idempotency.BeginResult{}, fmt.Errorf("idempotency: ttl must be positive (got %v)", ttl)
	}

	raw, err := idempotencyBeginScript.Run(
		ctx,
		s.client,
		[]string{idempotencyKey(scope, key)},
		requestHash,
		int64(ttl.Seconds()),
	).Slice()
	if err != nil {
		return shared_idempotency.BeginResult{}, err
	}
	if len(raw) == 0 {
		return shared_idempotency.BeginResult{}, fmt.Errorf("idempotency: empty script result")
	}

	switch luaString(raw[0]) {
	case string(shared_idempotency.StateNew):
		return shared_idempotency.BeginResult{State: shared_idempotency.StateNew}, nil
	case string(shared_idempotency.StateInFlight):
		return shared_idempotency.BeginResult{State: shared_idempotency.StateInFlight}, nil
	case string(shared_idempotency.StatePayloadMismatch):
		return shared_idempotency.BeginResult{State: shared_idempotency.StatePayloadMismatch}, nil
	case string(shared_idempotency.StateCompleted):
		if len(raw) < 4 {
			return shared_idempotency.BeginResult{}, fmt.Errorf("idempotency: incomplete completed result")
		}
		status, parseErr := strconv.Atoi(luaString(raw[1]))
		if parseErr != nil {
			return shared_idempotency.BeginResult{}, fmt.Errorf("idempotency: invalid stored status: %w", parseErr)
		}
		return shared_idempotency.BeginResult{
			State: shared_idempotency.StateCompleted,
			Response: &shared_idempotency.StoredResponse{
				Status:      status,
				Body:        []byte(luaString(raw[2])),
				ContentType: luaString(raw[3]),
			},
		}, nil
	default:
		return shared_idempotency.BeginResult{}, fmt.Errorf("idempotency: unknown state %q", luaString(raw[0]))
	}
}

func (s *RedisIdempotencyStore) Complete(ctx context.Context, scope, key string, response shared_idempotency.StoredResponse, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("idempotency: ttl must be positive (got %v)", ttl)
	}
	redisKey := idempotencyKey(scope, key)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, redisKey,
		"state", string(shared_idempotency.StateCompleted),
		"status", strconv.Itoa(response.Status),
		"body", response.Body,
		"contentType", response.ContentType,
	)
	pipe.Expire(ctx, redisKey, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisIdempotencyStore) Abort(ctx context.Context, scope, key string) error {
	return s.client.Del(ctx, idempotencyKey(scope, key)).Err()
}

func luaString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
