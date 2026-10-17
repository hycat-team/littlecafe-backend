package caching

import (
	"context"
	"fmt"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	"github.com/hycat-team/littlecafe-backend/shared/infrastructure/resilience"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(cfg *config.Config, l logger.Interface) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	var client *redis.Client

	err := resilience.RunStartupRetry(cfg, l, "redis", func(timeout time.Duration) error {
		rdb := redis.NewClient(redisOptions(cfg, timeout))

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			_ = rdb.Close()
			return fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
		}

		client = rdb
		return nil
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
