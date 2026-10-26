package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	shared_idempotency "github.com/hycat-team/littlecafe-backend/shared/application/interface/idempotency"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/utils/contextutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	defaultIdempotencyTTL         = 24 * time.Hour
	defaultIdempotencyStoreBudget = 5 * time.Second
)

// IdempotencyMiddleware enforces an Idempotency-Key for important mutations.
// Register it after authentication so the key scope can include the user id.
type IdempotencyMiddleware struct {
	store  shared_idempotency.Store
	ttl    time.Duration
	appEnv string
}

func NewIdempotencyMiddleware(store shared_idempotency.Store, ttl time.Duration, appEnv string) *IdempotencyMiddleware {
	if ttl <= 0 {
		ttl = defaultIdempotencyTTL
	}
	return &IdempotencyMiddleware{store: store, ttl: ttl, appEnv: appEnv}
}

func (m *IdempotencyMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if !validIdempotencyKey(key) {
			c.Error(apperror.NewBadRequest(i18n.ErrorIdempotencyRequired))
			c.Abort()
			return
		}

		requestBody, err := c.GetRawData()
		if err != nil {
			c.Error(apperror.NewBadRequest(i18n.ErrorValidationInvalidFormatDetail))
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

		scope := idempotencyScope(c)
		hash := sha256.Sum256(requestBody)
		result, err := m.store.Begin(c.Request.Context(), scope, key, hex.EncodeToString(hash[:]), m.ttl)
		if err != nil {
			c.Error(apperror.NewServiceUnavailable(i18n.ErrorIdempotencyUnavail))
			c.Abort()
			return
		}

		switch result.State {
		case shared_idempotency.StateCompleted:
			m.replay(c, result.Response)
			return
		case shared_idempotency.StateInFlight:
			c.Error(apperror.NewConflict(i18n.ErrorIdempotencyInFlight))
			c.Abort()
			return
		case shared_idempotency.StatePayloadMismatch:
			c.Error(apperror.NewConflict(i18n.ErrorIdempotencyMismatch))
			c.Abort()
			return
		case shared_idempotency.StateNew:
			// Continue below.
		default:
			c.Error(apperror.NewInternalError(apperror.GenericInternalMessage))
			c.Abort()
			return
		}

		claimed := true
		defer func() {
			if claimed {
				ctx, cancel := context.WithTimeout(context.Background(), defaultIdempotencyStoreBudget)
				defer cancel()
				_ = m.store.Abort(ctx, scope, key)
			}
		}()

		original := c.Writer
		capture := &captureWriter{ResponseWriter: original}
		c.Writer = capture
		defer func() {
			c.Writer = original
		}()
		c.Next()

		status := capture.Status()
		body := append([]byte(nil), capture.body.Bytes()...)
		contentType := capture.Header().Get("Content-Type")
		if len(body) == 0 && len(c.Errors) > 0 {
			if problemStatus, problemBody, ok := problemResponse(c, m.appEnv); ok {
				status = problemStatus
				body = problemBody
				contentType = "application/problem+json"
			}
		}
		if status == 0 {
			status = http.StatusOK
		}

		// 5xx is transient: release the key so the client can retry.
		if status >= http.StatusInternalServerError {
			return
		}

		storeCtx, cancelStore := context.WithTimeout(context.Background(), defaultIdempotencyStoreBudget)
		err = m.store.Complete(storeCtx, scope, key, shared_idempotency.StoredResponse{
			Status:      status,
			Body:        body,
			ContentType: contentType,
		}, m.ttl)
		cancelStore()
		if err != nil {
			// The response is already written; keep the key released so the
			// caller can retry and let the global handler log the store error.
			c.Error(err)
		} else {
			claimed = false
		}
	}
}

func (m *IdempotencyMiddleware) replay(c *gin.Context, response *shared_idempotency.StoredResponse) {
	if response == nil {
		c.Status(http.StatusNoContent)
		c.Abort()
		return
	}
	if response.Status == http.StatusNoContent || len(response.Body) == 0 {
		c.Status(response.Status)
		c.Abort()
		return
	}
	contentType := response.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(response.Status, contentType, response.Body)
	c.Abort()
}

func validIdempotencyKey(key string) bool {
	if len(key) < 16 || len(key) > 128 {
		return false
	}
	return key == strings.TrimSpace(key)
}

func idempotencyScope(c *gin.Context) string {
	userID := contextutils.GetUserIdOptional(c)
	subject := "anonymous"
	if userID != uuid.Nil {
		subject = userID.String()
	}
	return subject + ":" + c.Request.Method + ":" + c.Request.URL.Path
}

type captureWriter struct {
	gin.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *captureWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureWriter) WriteHeaderNow() {
	if w.status == 0 {
		w.status = w.ResponseWriter.Status()
	}
	w.ResponseWriter.WriteHeaderNow()
}

func (w *captureWriter) Status() int {
	if w.status != 0 {
		return w.status
	}
	return w.ResponseWriter.Status()
}

func (w *captureWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *captureWriter) WriteString(data string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}
