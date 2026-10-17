package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func TestDynamicFactoriesMapStatusAndTitle(t *testing.T) {
	cases := []struct {
		name   string
		err    *Error
		status int
		title  string
	}{
		{"NewBadRequest", NewBadRequest("detail"), http.StatusBadRequest, i18n.ErrorTitleBadRequest},
		{"NewNotFound", NewNotFound("detail"), http.StatusNotFound, i18n.ErrorTitleNotFound},
		{"NewTooManyRequest", NewTooManyRequest("detail"), http.StatusTooManyRequests, i18n.ErrorTitleTooManyRequests},
		{"NewRequestEntityTooLarge", NewRequestEntityTooLarge("detail"), http.StatusRequestEntityTooLarge, i18n.ErrorTitleRequestEntityTooLarge},
		{"NewConflict", NewConflict("detail"), http.StatusConflict, i18n.ErrorTitleConflict},
		{"NewUnauthorized", NewUnauthorized("detail"), http.StatusUnauthorized, i18n.ErrorTitleUnauthorized},
		{"NewForbidden", NewForbidden("detail"), http.StatusForbidden, i18n.ErrorTitleForbidden},
		{"NewInternalError", NewInternalError("detail"), http.StatusInternalServerError, i18n.ErrorTitleInternal},
		{"NewServiceUnavailable", NewServiceUnavailable("detail"), http.StatusServiceUnavailable, i18n.ErrorTitleServiceUnavailable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Status != tc.status {
				t.Fatalf("Status = %d, want %d", tc.err.Status, tc.status)
			}
			if tc.err.Title != msg(tc.title) {
				t.Fatalf("Title = %q, want %q", tc.err.Title, msg(tc.title))
			}
			if tc.err.Detail != "detail" || tc.err.Message != "detail" {
				t.Fatalf("Detail/Message = %q/%q, want detail", tc.err.Detail, tc.err.Message)
			}
			if tc.err.Type != "about:blank" {
				t.Fatalf("Type = %q, want about:blank", tc.err.Type)
			}
		})
	}
}

func TestNewErrorUsesExplicitStatusAndTitle(t *testing.T) {
	err := NewError(http.StatusTeapot, "Trà", "Ấm trà")

	if err.Status != http.StatusTeapot {
		t.Fatalf("Status = %d", err.Status)
	}
	if err.Title != "Trà" || err.Detail != "Ấm trà" {
		t.Fatalf("Title/Detail = %q/%q", err.Title, err.Detail)
	}
}

func TestNewUpstreamUnavailableKeepsCausePrivate(t *testing.T) {
	cause := errors.New("provider timeout: api-key=secret")
	err := NewUpstreamUnavailable(cause)

	if err.Status != http.StatusServiceUnavailable {
		t.Fatalf("Status = %d, want 503", err.Status)
	}
	if err.Message != msg(i18n.ErrorUpstreamUnavailable) {
		t.Fatalf("Message = %q", err.Message)
	}
	if err.Unwrap() != cause {
		t.Fatal("Cause must keep the provider error for logging")
	}
	if err.Error() == cause.Error() {
		t.Fatal("raw provider error must not leak through Error()")
	}
}
