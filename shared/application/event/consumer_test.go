package event

import (
	"context"
	"errors"
	"testing"
)

func TestMessageCorrelationId(t *testing.T) {
	if got := (Message{}).CorrelationId(); got != "" {
		t.Fatalf("nil headers = %q", got)
	}
	if got := (Message{Headers: map[string]any{}}).CorrelationId(); got != "" {
		t.Fatalf("empty headers = %q", got)
	}
	msg := Message{Headers: map[string]any{"x-correlation-id": "first"}}
	if got := msg.CorrelationId(); got != "first" {
		t.Fatalf("x-correlation-id = %q", got)
	}
	msg = Message{Headers: map[string]any{"correlationId": "second"}}
	if got := msg.CorrelationId(); got != "second" {
		t.Fatalf("correlationId = %q", got)
	}
	msg = Message{Headers: map[string]any{"x-correlation-id": 1, "correlationId": "third"}}
	if got := msg.CorrelationId(); got != "third" {
		t.Fatalf("non-string header must be skipped, got %q", got)
	}
}

func TestPermanentAndRetryableErrors(t *testing.T) {
	cause := errors.New("cause")
	permanent := PermanentError{Err: cause}
	if permanent.Error() != "cause" || !errors.Is(permanent, cause) {
		t.Fatalf("permanent = %v", permanent)
	}
	retryable := RetryableError{Err: cause}
	if retryable.Error() != "cause" {
		t.Fatalf("retryable = %v", retryable)
	}
	if !errors.Is(retryable, cause) {
		t.Fatal("RetryableError must unwrap to its cause")
	}
}

func TestCorrelationContext(t *testing.T) {
	if got := CorrelationIdFromContext(context.Background()); got != "" {
		t.Fatalf("empty context = %q", got)
	}
	ctx := ContextWithCorrelationId(context.Background(), Message{Headers: map[string]any{"correlationId": "abc"}})
	if got := CorrelationIdFromContext(ctx); got != "abc" {
		t.Fatalf("correlation = %q", got)
	}
}

type fakeConsumer struct {
	queue    string
	handler  func(ctx context.Context, msg Message) error
	opts     ConsumerOpts
	register int
}

func (f *fakeConsumer) Consume(string) (<-chan Message, error)       { return nil, nil }
func (f *fakeConsumer) RegisterConsumer(string, func(Message) error) {}
func (f *fakeConsumer) RegisterTypedHandler(queue string, handler func(context.Context, Message) error, opts ConsumerOpts) error {
	f.queue = queue
	f.handler = handler
	f.opts = opts
	f.register++
	return nil
}
func (f *fakeConsumer) Close() {}

func TestRegisterTypedConsumer(t *testing.T) {
	consumer := &fakeConsumer{}
	called := false
	err := RegisterTypedConsumer(consumer, "orders", func(_ context.Context, payload struct {
		ID string `json:"id"`
	}) error {
		called = true
		if payload.ID != "42" {
			t.Fatalf("payload = %+v", payload)
		}
		return nil
	}, ConsumerOpts{Prefetch: 5, Concurrency: 2})
	if err != nil {
		t.Fatalf("RegisterTypedConsumer: %v", err)
	}
	if consumer.register != 1 || consumer.queue != "orders" || consumer.opts.Prefetch != 5 {
		t.Fatalf("consumer = %+v", consumer)
	}
	if err := consumer.handler(context.Background(), Message{Body: []byte(`{"id":"42"}`)}); err != nil || !called {
		t.Fatalf("handler = %v, called = %v", err, called)
	}

	err = consumer.handler(context.Background(), Message{Body: []byte("not-json")})
	var permanent PermanentError
	if !errors.As(err, &permanent) {
		t.Fatalf("invalid JSON must produce PermanentError, got %T", err)
	}
}
