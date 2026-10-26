package presentation

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/application/sse"
)

func TestWriteSSEEmitsIDEventAndData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	err := WriteSSE(c, rec, sse.Event{
		ID:   "event-1",
		Name: "updated",
		Data: json.RawMessage(`{"status":"ok"}`),
	})
	if err != nil {
		t.Fatalf("WriteSSE: %v", err)
	}

	body := rec.Body.String()
	for _, want := range []string{
		"id: event-1\n",
		"event: updated\n",
		"data: {\"status\":\"ok\"}\n",
		"\n",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %q, missing %q", body, want)
		}
	}
}
