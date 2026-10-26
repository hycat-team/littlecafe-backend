package validation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

type bindRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type bindNestedItem struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type bindNestedRequest struct {
	Items []bindNestedItem `json:"items" binding:"required,dive"`
}

func TestBindJSONUsesNestedJSONPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var req bindNestedRequest
		err := BindJSON(c, &req)
		if err == nil {
			t.Fatal("expected a validation error")
		}
		appErr, ok := err.(*apperror.Error)
		if !ok {
			t.Fatalf("BindJSON error type = %T", err)
		}
		if len(appErr.Errors) != 1 {
			t.Fatalf("Errors = %v, want one item", appErr.Errors)
		}
		if got, want := appErr.Errors[0].Field, "items[0].quantity"; got != want {
			t.Fatalf("field = %q, want %q", got, want)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"items":[{"quantity":0}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBindJSONLocalizesValidationMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		locale i18n.Locale
		want   string
	}{
		{name: "vi", locale: i18n.VI, want: "Vui lòng nhập email"},
		{name: "en", locale: i18n.EN, want: "Please enter email"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				var req bindRequest
				if err := BindJSON(c, &req); err != nil {
					appErr, ok := err.(*apperror.Error)
					if !ok {
						t.Fatalf("BindJSON error type = %T", err)
					}
					if len(appErr.Errors) != 1 {
						t.Fatalf("Errors = %v, want one item", appErr.Errors)
					}
					c.JSON(http.StatusOK, gin.H{"message": appErr.Errors[0].Message})
					return
				}
				t.Fatal("expected a validation error")
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":""}`))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(i18n.WithLocale(req.Context(), tc.locale))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %q, want it to contain %q", rec.Body.String(), tc.want)
			}
		})
	}
}

func TestBindJSONUnmarshalTypeUsesNestedJSONPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var req bindNestedRequest
		err := BindJSON(c, &req)
		if err == nil {
			t.Fatal("expected a validation error")
		}
		appErr, ok := err.(*apperror.Error)
		if !ok {
			t.Fatalf("BindJSON error type = %T", err)
		}
		if len(appErr.Errors) != 1 {
			t.Fatalf("Errors = %v, want one item", appErr.Errors)
		}
		if got, want := appErr.Errors[0].Field, "items[1].quantity"; got != want {
			t.Fatalf("field = %q, want %q", got, want)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"items":[{"quantity":1},{"quantity":"not-a-number"}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBindJSONIgnoresUnknownFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var req bindRequest
		if err := BindJSON(c, &req); err != nil {
			t.Fatalf("BindJSON: %v", err)
		}
		if req.Email != "user@example.com" {
			t.Fatalf("email = %q", req.Email)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"user@example.com","unknown":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestGetUUIDParamLocalizes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/:id", func(c *gin.Context) {
		_, err := GetUUIDParam(c, "id")
		if err == nil {
			t.Fatal("expected invalid UUID error")
		}
		appErr, ok := err.(*apperror.Error)
		if !ok {
			t.Fatalf("error type = %T", err)
		}
		c.JSON(http.StatusOK, gin.H{"message": appErr.Localize(i18n.EN).Message})
	})

	req := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)
	req = req.WithContext(i18n.WithLocale(req.Context(), i18n.EN))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Invalid parameter id") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
