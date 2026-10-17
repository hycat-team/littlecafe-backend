package bootstrap

import (
	"context"

	"github.com/hycat-team/littlecafe-backend/shared/presentation/worker"
)

// ModuleWorker is implemented by module event workers. Constructing the
// implementation registers its consumers; the interface only forces Wire to
// build it and carries a name for lifecycle logging.
type ModuleWorker interface {
	Name() string
}

// Workers aggregates background workers owned by this service. Wire builds it
// only when the service runs with messaging enabled.
type Workers struct {
	outbox  *worker.OutboxWorker
	modules []ModuleWorker
}

// NewWorkers builds the worker group.
func NewWorkers(outbox *worker.OutboxWorker, modules []ModuleWorker) *Workers {
	return &Workers{outbox: outbox, modules: modules}
}

// Start launches every worker with the given context. Module workers start
// consuming when they are constructed (consumer registration is a side effect).
func (w *Workers) Start(ctx context.Context) {
	w.outbox.Start(ctx)
}

// Stop signals every worker to stop.
func (w *Workers) Stop() {
	w.outbox.Stop()
}
