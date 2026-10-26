package validation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
)

func TestIsUserName(t *testing.T) {
	v := binding.Validator.Engine().(*validator.Validate)
	type payload struct {
		Name string `binding:"username"`
	}
	cases := map[string]bool{
		"alice":                 true,
		"a.b_c1":                true,
		"abc":                   false,
		strings.Repeat("a", 21): false,
		"bad name":              false,
		"a__b":                  false,
		".alice":                false,
		"alice_":                false,
		"":                      false,
	}
	for value, want := range cases {
		err := v.Struct(payload{Name: value})
		if got := err == nil; got != want {
			t.Fatalf("username %q valid = %v, want %v (err %v)", value, got, want, err)
		}
	}
}

func TestIsPasswordComplex(t *testing.T) {
	v := binding.Validator.Engine().(*validator.Validate)
	type payload struct {
		Password string `binding:"password_complexity"`
	}
	cases := map[string]bool{
		"Abcdef1!":  true,
		"abcdef1!":  false,
		"ABCDEF1!":  false,
		"Abcdefgh!": false,
		"Abcdefg1":  false,
		"Ab1!":      false,
	}
	for value, want := range cases {
		err := v.Struct(payload{Password: value})
		if got := err == nil; got != want {
			t.Fatalf("password %q valid = %v, want %v", value, got, want)
		}
	}
}

func TestNotEqualField(t *testing.T) {
	v := binding.Validator.Engine().(*validator.Validate)
	type payload struct {
		Password        string `binding:"required"`
		ConfirmPassword string `binding:"nefield=Password"`
	}
	if err := v.Struct(payload{Password: "a", ConfirmPassword: "b"}); err != nil {
		t.Fatalf("distinct fields must pass: %v", err)
	}
	if err := v.Struct(payload{Password: "a", ConfirmPassword: "a"}); err == nil {
		t.Fatal("equal fields must fail")
	}
}

func TestNFCMax(t *testing.T) {
	v := binding.Validator.Engine().(*validator.Validate)
	type payload struct {
		Value string `binding:"nfcmax=3"`
	}
	if err := v.Struct(payload{Value: "abc"}); err != nil {
		t.Fatalf("3 runes must pass: %v", err)
	}
	if err := v.Struct(payload{Value: "abcd"}); err == nil {
		t.Fatal("4 runes must fail")
	}
	type invalid struct {
		Value string `binding:"nfcmax=-1"`
	}
	if err := v.Struct(invalid{Value: "abc"}); err == nil {
		t.Fatal("invalid limit must fail")
	}
}

func TestBindQueryAndURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/:id", func(c *gin.Context) {
		var query struct {
			Page int `form:"page" binding:"required,gt=0"`
		}
		if err := BindQuery(c, &query); err != nil {
			if _, ok := err.(*apperror.Error); !ok {
				t.Errorf("BindQuery error type = %T", err)
			}
		}
		var uri struct {
			ID string `uri:"id" binding:"required,uuid"`
		}
		if err := BindURI(c, &uri); err != nil {
			if _, ok := err.(*apperror.Error); !ok {
				t.Errorf("BindURI error type = %T", err)
			}
		}
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/not-a-uuid?page=0", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+uuid.NewString()+"?page=1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestParamHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/x/:id/:name/:count", func(c *gin.Context) {
		if _, err := GetUUIDParam(c, "id"); err != nil {
			t.Errorf("GetUUIDParam: %v", err)
		}
		if got, err := GetStringParam(c, "name"); err != nil || got != "hello" {
			t.Errorf("GetStringParam = %q, %v", got, err)
		}
		if got, err := GetInt64Param(c, "count"); err != nil || got != 42 {
			t.Errorf("GetInt64Param = %d, %v", got, err)
		}
		if _, err := GetStringParam(c, "missing"); err == nil {
			t.Error("missing string param must fail")
		}
		if _, err := GetUUIDParam(c, "missing"); err == nil {
			t.Error("missing UUID param must fail")
		}
		if _, err := GetInt64Param(c, "missing"); err == nil {
			t.Error("missing int param must fail")
		}
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x/"+uuid.NewString()+"/hello/42", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}

	router = gin.New()
	router.GET("/x/:count", func(c *gin.Context) {
		if _, err := GetInt64Param(c, "count"); err == nil {
			t.Error("invalid int param must fail")
		}
		c.Status(http.StatusNoContent)
	})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x/not-a-number", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

type messagePayload struct {
	Name     string `json:"name" binding:"required" label:"Full name"`
	Email    string `json:"email" binding:"email"`
	UserName string `json:"userName" binding:"username"`
	Password string `json:"password" binding:"password_complexity"`
	Confirm  string `json:"confirm" binding:"nefield=Password"`
	Code     string `json:"code" binding:"len=2"`
	Score    int    `json:"score" binding:"gt=1,min=2,max=5"`
	Label    string `json:"label" binding:"max=2"`
	Date     string `json:"date" binding:"datetime=2006-01-02"`
	Choice   string `json:"choice" binding:"oneof=a b"`
	Number   string `json:"number" binding:"numeric"`
	Unknown  string `json:"unknown" binding:"startswith=z"`
	Equal    string `json:"equal" binding:"eqfield=Name"`
}

func TestTranslateValidationErrorCoversAllTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	v := binding.Validator.Engine().(*validator.Validate)
	payload := messagePayload{
		Name:     "",
		Email:    "bad",
		UserName: "a",
		Password: "weak",
		Confirm:  "",
		Code:     "x",
		Score:    0,
		Label:    "abc",
		Date:     "bad",
		Choice:   "c",
		Number:   "x",
		Unknown:  "a",
		Equal:    "other",
	}
	err := v.Struct(payload)
	if err == nil {
		t.Fatal("payload must fail validation")
	}
	appErr := TranslateValidationError(i18nContext(), err)
	if appErr == nil || len(appErr.Errors) < 10 {
		t.Fatalf("errors = %+v", appErr)
	}
	for _, item := range appErr.Errors {
		if item.Field == "" || item.Message == "" {
			t.Fatalf("incomplete item: %+v", item)
		}
	}
}

func i18nContext() context.Context {
	return context.Background()
}

func TestTranslateValidationErrorInputs(t *testing.T) {
	ctx := context.Background()

	typeErr := &json.UnmarshalTypeError{Field: "items.quantity", Offset: 1}
	appErr := translateValidationError(ctx, typeErr, []byte(`{"items":[{"quantity":"x"}]}`))
	if len(appErr.Errors) != 1 || appErr.Errors[0].Field != "items[0].quantity" {
		t.Fatalf("type error = %+v", appErr.Errors)
	}

	appErr = translateValidationError(ctx, typeErr, nil)
	if len(appErr.Errors) != 1 || appErr.Errors[0].Field != "items.quantity" {
		t.Fatalf("type error without raw = %+v", appErr.Errors)
	}

	appErr = translateValidationError(ctx, errors.New("cannot unmarshal value"), nil)
	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("status = %d", appErr.Status)
	}

	appErr = translateValidationError(ctx, &http.MaxBytesError{Limit: 2048}, nil)
	if appErr.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d", appErr.Status)
	}

	appErr = translateValidationError(ctx, errors.New("boom"), nil)
	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("status = %d", appErr.Status)
	}
}

func TestLowerCamelCaseAndHumanize(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"UserName":   "userName",
		"user_name":  "userName",
		"user__name": "userName",
		"Already":    "already",
	}
	for in, want := range cases {
		if got := lowerCamelCase(in); got != want {
			t.Fatalf("lowerCamelCase(%q) = %q, want %q", in, got, want)
		}
	}
	if got := humanizeFieldName("UserName"); got != "user name" {
		t.Fatalf("humanizeFieldName = %q", got)
	}
	if got := humanizeFieldName(""); got != "" {
		t.Fatalf("humanizeFieldName empty = %q", got)
	}
}

func TestBindJSONBodyReadFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", func(c *gin.Context) {
		var payload struct{}
		if err := BindJSON(c, &payload); err == nil {
			t.Error("expected a read error")
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Body = errReader{}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errReader) Close() error             { return nil }

func TestFieldLabelUsesLocalizedAndJSONNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	v := binding.Validator.Engine().(*validator.Validate)
	type payload struct {
		DisplayName string `json:"displayName" binding:"required"`
		WithLabel   string `json:"withLabel" binding:"required" label:"Custom label"`
	}
	err := v.Struct(payload{})
	appErr := TranslateValidationError(context.Background(), err)
	if len(appErr.Errors) != 2 {
		t.Fatalf("errors = %+v", appErr.Errors)
	}
	if appErr.Errors[0].Field != "displayName" || appErr.Errors[1].Field != "withLabel" {
		t.Fatalf("fields = %+v", appErr.Errors)
	}
}
