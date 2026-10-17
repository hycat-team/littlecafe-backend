package presentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func TestLocalizedSuccessResolvesRequestLocale(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), i18n.EN))
		if got := Message(c, i18n.ErrorTitleNotFound); got != "Data not found" {
			t.Fatalf("Message = %q", got)
		}
		LocalizedSuccess(c, i18n.SuccessOK, gin.H{"status": "healthy"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Message != "ok" {
		t.Fatalf("message = %q", body.Message)
	}
}
