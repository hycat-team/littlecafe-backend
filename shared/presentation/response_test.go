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

func TestSuccessAlwaysIncludesMessageAndData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		Success(c, "", nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got, want := rec.Body.String(), `{"message":"","data":null}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestCreatedWithLocationSetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		CreatedWithLocation(c, "/api/v1/order/orders/1", "created", gin.H{"id": "1"})
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/api/v1/order/orders/1" {
		t.Fatalf("location = %q", got)
	}
}

func TestNoContentHasNoBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.DELETE("/", func(c *gin.Context) {
		NoContent(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
}
