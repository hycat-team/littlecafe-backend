package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	shared_idempotency "github.com/hycat-team/littlecafe-backend/shared/application/interface/idempotency"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/logger"

	"github.com/gin-gonic/gin"
)

type fakeIdempotencyStore struct {
	beginResult  shared_idempotency.BeginResult
	beginErr     error
	completeErr  error
	completed    *shared_idempotency.StoredResponse
	aborted      bool
	beginCalls   int
	completeCall int
}

func (f *fakeIdempotencyStore) Begin(context.Context, string, string, string, time.Duration) (shared_idempotency.BeginResult, error) {
	f.beginCalls++
	return f.beginResult, f.beginErr
}

func (f *fakeIdempotencyStore) Complete(_ context.Context, _, _ string, response shared_idempotency.StoredResponse, _ time.Duration) error {
	f.completeCall++
	f.completed = &response
	return f.completeErr
}

func (f *fakeIdempotencyStore) Abort(context.Context, string, string) error {
	f.aborted = true
	return nil
}

func newIdempotencyTestRouter(store shared_idempotency.Store, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	mw := NewIdempotencyMiddleware(store, time.Hour, "production")
	r := gin.New()
	r.Use(GlobalErrorHandler(logger.New("test", "", "error", false), "production"))
	r.Use(mw.Handle())
	r.POST("/mutate", handler)
	return r
}

func idempotencyRequest(key string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(`{"amount":100}`))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	return req
}

func TestIdempotencyRequiresKey(t *testing.T) {
	store := &fakeIdempotencyStore{}
	called := false
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		called = true
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("short"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if called {
		t.Fatal("handler must not run without a valid key")
	}
	if store.beginCalls != 0 {
		t.Fatalf("begin calls = %d, want 0", store.beginCalls)
	}
}

func TestIdempotencyReplaysCompletedResponse(t *testing.T) {
	store := &fakeIdempotencyStore{beginResult: shared_idempotency.BeginResult{
		State: shared_idempotency.StateCompleted,
		Response: &shared_idempotency.StoredResponse{
			Status:      http.StatusCreated,
			Body:        []byte(`{"message":"ok","data":{"id":"1"}}`),
			ContentType: "application/json",
		},
	}}
	called := false
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		called = true
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if called {
		t.Fatal("handler must not run when replaying")
	}
	if rec.Body.String() != `{"message":"ok","data":{"id":"1"}}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestIdempotencyRejectsInFlightAndMismatch(t *testing.T) {
	tests := []struct {
		name  string
		state shared_idempotency.State
	}{
		{name: "in flight", state: shared_idempotency.StateInFlight},
		{name: "payload mismatch", state: shared_idempotency.StatePayloadMismatch},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeIdempotencyStore{beginResult: shared_idempotency.BeginResult{State: tc.state}}
			called := false
			r := newIdempotencyTestRouter(store, func(c *gin.Context) {
				called = true
			})

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409", rec.Code)
			}
			if called {
				t.Fatal("handler must not run on a conflict")
			}
		})
	}
}

func TestIdempotencyStoresSuccessfulResponse(t *testing.T) {
	store := &fakeIdempotencyStore{beginResult: shared_idempotency.BeginResult{State: shared_idempotency.StateNew}}
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"message": "ok", "data": gin.H{"id": "1"}})
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if store.completeCall != 1 {
		t.Fatalf("complete calls = %d, want 1", store.completeCall)
	}
	if store.completed == nil || store.completed.Status != http.StatusCreated {
		t.Fatalf("stored response = %+v", store.completed)
	}
	if store.aborted {
		t.Fatal("key must not be aborted after a stored response")
	}
	var body map[string]any
	if err := json.Unmarshal(store.completed.Body, &body); err != nil {
		t.Fatalf("stored body: %v", err)
	}
	if body["message"] != "ok" {
		t.Fatalf("stored body = %s", store.completed.Body)
	}
}

func TestIdempotencyStoresProblemDetails(t *testing.T) {
	store := &fakeIdempotencyStore{beginResult: shared_idempotency.BeginResult{State: shared_idempotency.StateNew}}
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		c.Error(apperror.NewBadRequest(i18n.ErrorValidationInvalidFormatDetail))
		c.Abort()
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if store.completed == nil || store.completed.Status != http.StatusBadRequest {
		t.Fatalf("stored response = %+v", store.completed)
	}
	if store.completed.ContentType != "application/problem+json" {
		t.Fatalf("content type = %q", store.completed.ContentType)
	}
	var problem map[string]any
	if err := json.Unmarshal(store.completed.Body, &problem); err != nil {
		t.Fatalf("stored problem: %v", err)
	}
	if problem["status"] != float64(http.StatusBadRequest) {
		t.Fatalf("stored problem = %s", store.completed.Body)
	}
}

func TestIdempotencyAbortsKeyOnServerError(t *testing.T) {
	store := &fakeIdempotencyStore{beginResult: shared_idempotency.BeginResult{State: shared_idempotency.StateNew}}
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		c.Error(apperror.NewInternalError(apperror.GenericInternalMessage))
		c.Abort()
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !store.aborted {
		t.Fatal("key must be aborted after a 5xx")
	}
	if store.completeCall != 0 {
		t.Fatalf("complete calls = %d, want 0", store.completeCall)
	}
}

func TestIdempotencyFailsClosedWhenStoreUnavailable(t *testing.T) {
	store := &fakeIdempotencyStore{beginErr: context.DeadlineExceeded}
	called := false
	r := newIdempotencyTestRouter(store, func(c *gin.Context) {
		called = true
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, idempotencyRequest("idempotency-key-123"))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if called {
		t.Fatal("handler must not run when the store is unavailable")
	}
}
