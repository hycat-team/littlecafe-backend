package contextutils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func TestGetUserId(t *testing.T) {
	c := newContext()
	if _, err := GetUserId(c); err == nil {
		t.Fatal("expected an error when the context has no user")
	}

	want := uuid.New()
	c.Set(CtxUserId, want)
	if got, err := GetUserId(c); err != nil || got != want {
		t.Fatalf("GetUserId = %v, %v", got, err)
	}

	c.Set(CtxUserId, want.String())
	if got, err := GetUserId(c); err != nil || got != want {
		t.Fatalf("GetUserId from string = %v, %v", got, err)
	}

	c.Set(CtxUserId, "not-a-uuid")
	if _, err := GetUserId(c); err == nil {
		t.Fatal("expected an error for an invalid string")
	}

	c.Set(CtxUserId, 42)
	if _, err := GetUserId(c); err == nil {
		t.Fatal("expected an error for an unsupported type")
	}
}

func TestGetUserIdOptional(t *testing.T) {
	c := newContext()
	if got := GetUserIdOptional(c); got != uuid.Nil {
		t.Fatalf("missing = %v", got)
	}

	want := uuid.New()
	c.Set(CtxUserId, want)
	if got := GetUserIdOptional(c); got != want {
		t.Fatalf("uuid = %v", got)
	}

	c.Set(CtxUserId, want.String())
	if got := GetUserIdOptional(c); got != want {
		t.Fatalf("string = %v", got)
	}

	c.Set(CtxUserId, "bad")
	if got := GetUserIdOptional(c); got != uuid.Nil {
		t.Fatalf("invalid string = %v", got)
	}

	c.Set(CtxUserId, 1)
	if got := GetUserIdOptional(c); got != uuid.Nil {
		t.Fatalf("unsupported = %v", got)
	}
}

func TestGetRoleSlug(t *testing.T) {
	c := newContext()
	if _, err := GetRoleSlug(c); err == nil {
		t.Fatal("expected an error when missing")
	}

	c.Set(CtxRoleSlug, "ADMIN")
	if got, err := GetRoleSlug(c); err != nil || string(got) != "ADMIN" {
		t.Fatalf("GetRoleSlug = %q, %v", got, err)
	}

	c.Set(CtxRoleSlug, 1)
	if _, err := GetRoleSlug(c); err == nil {
		t.Fatal("expected an error for an unsupported type")
	}
}

func TestGetAccessToken(t *testing.T) {
	c := newContext()
	if _, err := GetAccessToken(c); err == nil {
		t.Fatal("expected an error when missing")
	}

	c.Set(CtxAccessToken, "token")
	if got, err := GetAccessToken(c); err != nil || got != "token" {
		t.Fatalf("GetAccessToken = %q, %v", got, err)
	}

	c.Set(CtxAccessToken, 1)
	if _, err := GetAccessToken(c); err == nil {
		t.Fatal("expected an error for an unsupported type")
	}
}
