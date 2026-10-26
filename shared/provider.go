// Package shared wires the cross-service framework packages into each
// service's dependency graph.
package shared

import (
	"github.com/google/wire"

	app_audit "github.com/hycat-team/littlecafe-backend/shared/application/audit"
	"github.com/hycat-team/littlecafe-backend/shared/application/event"
	sharedUcInterface "github.com/hycat-team/littlecafe-backend/shared/application/interface/usecase"
	sse "github.com/hycat-team/littlecafe-backend/shared/application/sse"
	sharedUsecase "github.com/hycat-team/littlecafe-backend/shared/application/usecase"
	infra_audit "github.com/hycat-team/littlecafe-backend/shared/infrastructure/audit"
	infra_caching "github.com/hycat-team/littlecafe-backend/shared/infrastructure/caching"
	infra_db "github.com/hycat-team/littlecafe-backend/shared/infrastructure/db"
	infra_media "github.com/hycat-team/littlecafe-backend/shared/infrastructure/media"
	"github.com/hycat-team/littlecafe-backend/shared/infrastructure/messaging"
	sharedWorker "github.com/hycat-team/littlecafe-backend/shared/presentation/worker"
)

// ProviderSet binds shared infrastructure to its application ports. Feature
// gated providers appear only when the feature is enabled.
var ProviderSet = wire.NewSet(
	infra_db.NewGormUnitOfWork,
	infra_caching.NewRedisCacheService,
	infra_caching.NewRedisRateLimiter,
	infra_caching.NewRedisTokenBucketLimiter,
	infra_caching.NewRedisCooldownService,
	infra_caching.NewRedisIdempotencyStore,
	messaging.NewOutboxRepository,
	messaging.NewRabbitMQClient,
	messaging.NewOutboxGaugeProbe,
	infra_db.NewOutboxListener,
	sharedUsecase.NewOutboxUseCase,
	sharedWorker.NewOutboxWorker,
	wire.Bind(new(event.IMessageConsumer), new(*messaging.RabbitMQClient)),
	wire.Bind(new(event.IEventPublisher), new(*messaging.RabbitMQClient)),
	wire.Bind(new(event.IRawMessagePublisher), new(*messaging.RabbitMQClient)),
	wire.Bind(new(event.IOutboxGaugeProbe), new(*messaging.OutboxGaugeProbe)),
	wire.Bind(new(sharedWorker.OutboxSignalListener), new(*infra_db.OutboxListener)),
	wire.Bind(new(sharedUcInterface.IOutboxUseCase), new(*sharedUsecase.OutboxUseCase)),
	sse.NewSSEManager,
	wire.Bind(new(sse.ISSEManager), new(*sse.SSEManager)),
	infra_media.NewCloudinaryService,
	infra_audit.NewLogRecorder,
	wire.Bind(new(app_audit.Recorder), new(*infra_audit.LogRecorder)),
)
