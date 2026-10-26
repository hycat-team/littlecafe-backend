package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/hycat-team/littlecafe-backend/shared/application/event"
	"github.com/hycat-team/littlecafe-backend/shared/config"
	sharedrepos "github.com/hycat-team/littlecafe-backend/shared/domain/repositories"
)

type fakeOutboxRepo struct {
	claimRecords []sharedrepos.OutboxRecord
	claimErr     error
	countPending int64
	countErr     error
	deleted      int64
	deleteErr    error

	markPublishedErr error
	markFailedErr    error
	markPublishedIDs []uuid.UUID
	markFailedIDs    []uuid.UUID
}

func (f *fakeOutboxRepo) Create(context.Context, *sharedrepos.OutboxRecord) error { return nil }
func (f *fakeOutboxRepo) ClaimToPublish(context.Context, int, int, time.Duration) ([]sharedrepos.OutboxRecord, error) {
	return f.claimRecords, f.claimErr
}
func (f *fakeOutboxRepo) MarkPublished(_ context.Context, id uuid.UUID) error {
	f.markPublishedIDs = append(f.markPublishedIDs, id)
	return f.markPublishedErr
}
func (f *fakeOutboxRepo) MarkFailed(_ context.Context, id uuid.UUID, _ error) error {
	f.markFailedIDs = append(f.markFailedIDs, id)
	return f.markFailedErr
}
func (f *fakeOutboxRepo) DeletePublishedBefore(context.Context, time.Time) (int64, error) {
	return f.deleted, f.deleteErr
}
func (f *fakeOutboxRepo) CountPending(context.Context) (int64, error) {
	return f.countPending, f.countErr
}
func (f *fakeOutboxRepo) SendNotification(context.Context) error { return nil }

type fakePublisher struct {
	mu      sync.Mutex
	err     error
	publish int
	headers []map[string]any
}

func (f *fakePublisher) PublishDirect(_ context.Context, _ string, _ []byte, headers map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.publish++
	f.headers = append(f.headers, headers)
	return f.err
}

func (f *fakePublisher) snapshot() (int, []map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.publish, append([]map[string]any(nil), f.headers...)
}

type fakeGauge struct {
	calls   int
	pending int64
}

func (f *fakeGauge) ProbeGauges(_ context.Context, pending int64) {
	f.calls++
	f.pending = pending
}

type captureLogger struct {
	infos  int
	warns  int
	errors int
}

func (c *captureLogger) Debug(string, ...zap.Field) { c.infos++ }
func (c *captureLogger) Info(string, ...zap.Field)  { c.infos++ }
func (c *captureLogger) Warn(string, ...zap.Field)  { c.warns++ }
func (c *captureLogger) Error(string, ...zap.Field) { c.errors++ }
func (c *captureLogger) Fatal(string, ...zap.Field) {}

func newUseCase(repo *fakeOutboxRepo, pub *fakePublisher, gauge *fakeGauge, log *captureLogger) *OutboxUseCase {
	return NewOutboxUseCase(repo, pub, gauge, &config.Config{}, log)
}

func TestRefreshGauges(t *testing.T) {
	repo := &fakeOutboxRepo{countErr: errors.New("db down")}
	gauge := &fakeGauge{}
	log := &captureLogger{}
	newUseCase(repo, &fakePublisher{}, gauge, log).RefreshGauges(context.Background())
	if gauge.calls != 0 || log.warns != 1 {
		t.Fatalf("gauge calls = %d, warnings = %d", gauge.calls, log.warns)
	}

	repo = &fakeOutboxRepo{countPending: 7}
	gauge = &fakeGauge{}
	newUseCase(repo, &fakePublisher{}, gauge, &captureLogger{}).RefreshGauges(context.Background())
	if gauge.calls != 1 || gauge.pending != 7 {
		t.Fatalf("gauge = %+v", gauge)
	}
}

func TestPublishBatchClaimError(t *testing.T) {
	repo := &fakeOutboxRepo{claimErr: errors.New("claim failed")}
	log := &captureLogger{}
	count, err := newUseCase(repo, &fakePublisher{}, &fakeGauge{}, log).PublishBatch(context.Background(), 10, 3, time.Minute)
	if err == nil || count != 0 || log.errors != 1 {
		t.Fatalf("count = %d, err = %v, errors = %d", count, err, log.errors)
	}
}

func TestPublishBatchEmpty(t *testing.T) {
	count, err := newUseCase(&fakeOutboxRepo{}, &fakePublisher{}, &fakeGauge{}, &captureLogger{}).
		PublishBatch(context.Background(), 10, 3, time.Minute)
	if err != nil || count != 0 {
		t.Fatalf("count = %d, err = %v", count, err)
	}
}

func TestPublishBatchPublishesAndMarksRecords(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	repo := &fakeOutboxRepo{claimRecords: []sharedrepos.OutboxRecord{
		{ID: first, Topic: "topic.a", Payload: []byte(`{"a":1}`), Headers: []byte(`{"correlationId":"corr-1"}`)},
		{ID: second, Topic: "topic.b", Payload: []byte(`{"b":2}`), Headers: []byte(`{}`)},
	}}
	pub := &fakePublisher{}
	count, err := newUseCase(repo, pub, &fakeGauge{}, &captureLogger{}).
		PublishBatch(context.Background(), 10, 3, time.Minute)
	if err != nil || count != 2 {
		t.Fatalf("count = %d, err = %v", count, err)
	}
	published, headers := pub.snapshot()
	if published != 2 || len(repo.markPublishedIDs) != 2 {
		t.Fatalf("publish = %d, marked = %d", published, len(repo.markPublishedIDs))
	}
	found := false
	for _, headers := range headers {
		if headers["x-correlation-id"] == "corr-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("correlation header was not mapped: %v", pub.headers)
	}
}

func TestPublishBatchHandlesFailures(t *testing.T) {
	id := uuid.New()
	repo := &fakeOutboxRepo{
		claimRecords:     []sharedrepos.OutboxRecord{{ID: id, Topic: "t", Payload: []byte("{}"), Headers: []byte("not-json")}},
		markFailedErr:    errors.New("mark failed"),
		markPublishedErr: errors.New("mark published failed"),
	}
	pub := &fakePublisher{err: errors.New("publish failed")}
	log := &captureLogger{}
	count, err := newUseCase(repo, pub, &fakeGauge{}, log).
		PublishBatch(context.Background(), 10, 3, time.Minute)
	if err != nil || count != 0 {
		t.Fatalf("count = %d, err = %v", count, err)
	}
	if len(repo.markFailedIDs) != 1 {
		t.Fatalf("failed records = %v", repo.markFailedIDs)
	}
	if log.errors < 2 {
		t.Fatalf("errors = %d, want at least 2", log.errors)
	}

	repo = &fakeOutboxRepo{
		claimRecords:     []sharedrepos.OutboxRecord{{ID: id, Topic: "t", Payload: []byte("{}")}},
		markPublishedErr: errors.New("mark failed"),
	}
	count, err = newUseCase(repo, &fakePublisher{}, &fakeGauge{}, &captureLogger{}).
		PublishBatch(context.Background(), 10, 3, time.Minute)
	if err != nil || count != 1 {
		t.Fatalf("mark published failure count = %d, err = %v", count, err)
	}
}

func TestCleanupPublished(t *testing.T) {
	repo := &fakeOutboxRepo{deleteErr: errors.New("delete failed")}
	log := &captureLogger{}
	if _, err := newUseCase(repo, &fakePublisher{}, &fakeGauge{}, log).CleanupPublished(context.Background(), 7); err == nil {
		t.Fatal("expected cleanup error")
	}
	if log.errors != 1 {
		t.Fatalf("errors = %d", log.errors)
	}

	repo = &fakeOutboxRepo{deleted: 3}
	log = &captureLogger{}
	deleted, err := newUseCase(repo, &fakePublisher{}, &fakeGauge{}, log).CleanupPublished(context.Background(), 7)
	if err != nil || deleted != 3 || log.infos != 1 {
		t.Fatalf("deleted = %d, err = %v, infos = %d", deleted, err, log.infos)
	}

	repo = &fakeOutboxRepo{}
	log = &captureLogger{}
	if deleted, err := newUseCase(repo, &fakePublisher{}, &fakeGauge{}, log).CleanupPublished(context.Background(), 7); err != nil || deleted != 0 || log.infos != 0 {
		t.Fatalf("empty cleanup = %d, %v, infos = %d", deleted, err, log.infos)
	}
}

var _ event.IRawMessagePublisher = (*fakePublisher)(nil)
var _ event.IOutboxGaugeProbe = (*fakeGauge)(nil)
