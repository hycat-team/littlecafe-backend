package presentation

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
)

const DefaultSSEChunkDeadline = 90 * time.Second

var sseChunkDeadline atomic.Value // time.Duration

func init() {
	ResetSSEDeadlines()
}

// ConfigureSSEDeadlines seeds the per-chunk deadline from config. Zero or
// negative values fall back to the default.
func ConfigureSSEDeadlines(chunk time.Duration) {
	if chunk <= 0 {
		chunk = DefaultSSEChunkDeadline
	}
	sseChunkDeadline.Store(chunk)
}

// ResetSSEDeadlines restores the built-in default (used by tests).
func ResetSSEDeadlines() {
	sseChunkDeadline.Store(DefaultSSEChunkDeadline)
}

func currentChunkDeadline() time.Duration {
	return sseChunkDeadline.Load().(time.Duration)
}

// InitSSE configures the necessary headers for SSE and returns http.Flusher.
func InitSSE(c *gin.Context) (http.Flusher, error) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	ExtendSSEWriteDeadline(c)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, apperror.NewInternalError("Trình duyệt hoặc máy chủ không hỗ trợ kết nối thời gian thực.")
	}
	return flusher, nil
}

// ExtendSSEWriteDeadline pushes the connection write deadline further out so a
// long-running event stream is not cut by the server-wide WriteTimeout.
func ExtendSSEWriteDeadline(c *gin.Context) {
	rc := http.NewResponseController(c.Writer)
	_ = rc.SetWriteDeadline(time.Now().Add(currentChunkDeadline()))
}
