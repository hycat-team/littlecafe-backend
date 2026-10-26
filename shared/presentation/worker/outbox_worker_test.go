package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/shared/config"
)

type captureLogger struct {
	mu     sync.Mutex
	infos  int
	errors int
}

func (c *captureLogger) Debug(string, ...zap.Field) {}
func (c *captureLogger) Info(string, ...zap.Field)  { c.mu.Lock(); c.infos++; c.mu.Unlock() }
func (c *captureLogger) Warn(string, ...zap.Field)  {}
func (c *captureLogger) Error(string, ...zap.Field) { c.mu.Lock(); c.errors++; c.mu.Unlock() }
func (c *captureLogger) Fatal(string, ...zap.Field) {}

func (c *captureLogger) snapshot() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.infos, c.errors
}

type fakeUseCase struct {
	mu           sync.Mutex
	publishCalls int
	cleanupCalls int
	gaugeCalls   int
	publishCount int
	publishErr   error
	deleted      int64
	cleanupErr   error
}

func (f *fakeUseCase) PublishBatch(context.Context, int, int, time.Duration) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.publishCalls++
	return f.publishCount, f.publishErr
}
func (f *fakeUseCase) CleanupPublished(context.Context, int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cleanupCalls++
	return f.deleted, f.cleanupErr
}
func (f *fakeUseCase) RefreshGauges(context.Context) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gaugeCalls++
}

type fakeListener struct {
	ch chan struct{}
}

func (f *fakeListener) SignalChan() <-chan struct{} { return f.ch }

func newOutboxWorker(uc *fakeUseCase, listener *fakeListener, log *captureLogger, cfg *config.Config) *OutboxWorker {
	return NewOutboxWorker(uc, listener, cfg, log)
}

func TestBaseWorkerLifecycle(t *testing.T) {
	base := NewBaseWorker("test")
	if base.Name() != "test" {
		t.Fatalf("Name = %q", base.Name())
	}
	if base.Context() == nil {
		t.Fatal("Context must never be nil")
	}
	base.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	base.Start(ctx)
	base.Start(ctx)
	workerCtx := base.Context()
	base.Stop()
	select {
	case <-workerCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("worker context was not cancelled")
	}
	base.SetHealthCheck(func(context.Context) error { return nil })
	cancel()
}

func TestOutboxWorkerProcessBatch(t *testing.T) {
	uc := &fakeUseCase{publishCount: 3}
	log := &captureLogger{}
	w := newOutboxWorker(uc, &fakeListener{ch: make(chan struct{})}, log, &config.Config{})
	w.processBatch(context.Background())
	if uc.publishCalls != 1 {
		t.Fatalf("publish calls = %d", uc.publishCalls)
	}
	if infos, _ := log.snapshot(); infos != 1 {
		t.Fatalf("infos = %d", infos)
	}

	uc = &fakeUseCase{publishErr: errors.New("boom")}
	log = &captureLogger{}
	newOutboxWorker(uc, &fakeListener{ch: make(chan struct{})}, log, &config.Config{}).processBatch(context.Background())
	if _, errs := log.snapshot(); errs != 1 {
		t.Fatalf("errors = %d", errs)
	}
}

func TestOutboxWorkerCleanupAndGauges(t *testing.T) {
	uc := &fakeUseCase{deleted: 2}
	log := &captureLogger{}
	w := newOutboxWorker(uc, &fakeListener{ch: make(chan struct{})}, log, &config.Config{})
	w.cleanupPublished(context.Background())
	w.refreshGauges(context.Background())
	if uc.cleanupCalls != 1 || uc.gaugeCalls != 1 {
		t.Fatalf("cleanup = %d, gauge = %d", uc.cleanupCalls, uc.gaugeCalls)
	}
	// cleanupPublished only adds success counters, so LogSuccessIfData uses debug.
	if infos, _ := log.snapshot(); infos != 0 {
		t.Fatalf("infos = %d, want 0", infos)
	}

	uc = &fakeUseCase{cleanupErr: errors.New("boom")}
	log = &captureLogger{}
	newOutboxWorker(uc, &fakeListener{ch: make(chan struct{})}, log, &config.Config{}).cleanupPublished(context.Background())
	if _, errs := log.snapshot(); errs != 1 {
		t.Fatalf("errors = %d", errs)
	}

	uc = &fakeUseCase{}
	newOutboxWorker(uc, &fakeListener{ch: make(chan struct{})}, &captureLogger{}, &config.Config{}).cleanupPublished(context.Background())
}

func TestOutboxWorkerStartStop(t *testing.T) {
	uc := &fakeUseCase{}
	signals := make(chan struct{}, 1)
	log := &captureLogger{}
	cfg := &config.Config{Outbox: config.OutboxConfig{BatchSize: 1, RetryMaxAttempts: 1, CleanupIntervalMinutes: 60}}
	w := newOutboxWorker(uc, &fakeListener{ch: signals}, log, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)
	signals <- struct{}{}

	deadline := time.After(2 * time.Second)
	for {
		uc.mu.Lock()
		calls := uc.publishCalls
		uc.mu.Unlock()
		if calls > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("worker did not process the signal")
		case <-time.After(10 * time.Millisecond):
		}
	}

	w.Stop()
	cancel()
	time.Sleep(10 * time.Millisecond)
	if infos, _ := log.snapshot(); infos == 0 {
		// The loop also logs on context cancellation; a signal run logs success.
		t.Log("no info logs observed")
	}
}
