package messaging

import (
	"context"

	"github.com/hycat-team/littlecafe-backend/shared/application/sse"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisRealtimePublisher publishes SSE events to a Redis Pub/Sub channel so the
// relay running in the HTTP process can deliver them to clients in both
// topologies (single-process APP_MODE=all and split http/worker). Publishing is
// best-effort: a Redis failure is logged and swallowed so it never blocks or
// fails the business flow.
type RedisRealtimePublisher struct {
	client *redis.Client
	log    logger.Interface
}

// NewRedisRealtimePublisher builds the Redis-backed realtime publisher.
func NewRedisRealtimePublisher(client *redis.Client, log logger.Interface) *RedisRealtimePublisher {
	return &RedisRealtimePublisher{client: client, log: log}
}

// Publish sends the event bytes to the stream's channel. It always returns nil
// to keep the caller's business flow uninterrupted; transport errors are
// logged.
func (p *RedisRealtimePublisher) Publish(ctx context.Context, streamId string, data []byte) error {
	if p.client == nil {
		return nil
	}
	sse.IncRealtimePublishTotal()
	if err := p.client.Publish(ctx, sse.RealtimeChannel(streamId), data).Err(); err != nil {
		sse.IncRealtimePublishFailed()
		p.log.Warn("realtime publish failed",
			zap.String("stream_id", streamId),
			zap.Error(err),
		)
	}
	return nil
}
