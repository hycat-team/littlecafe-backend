package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewSetsRFC9457Defaults(t *testing.T) {
	err := New(http.StatusBadRequest, "Thao tác không thành công", "Dữ liệu không hợp lệ")

	if err.Type != "about:blank" {
		t.Fatalf("Type = %q, want about:blank", err.Type)
	}
	if err.Status != http.StatusBadRequest {
		t.Fatalf("Status = %d, want %d", err.Status, http.StatusBadRequest)
	}
	if err.Title != "Thao tác không thành công" {
		t.Fatalf("Title = %q", err.Title)
	}
	if err.Detail != "Dữ liệu không hợp lệ" {
		t.Fatalf("Detail = %q", err.Detail)
	}
	if err.Message != "Dữ liệu không hợp lệ" {
		t.Fatalf("Message = %q", err.Message)
	}
	if err.Stack() == "" {
		t.Fatal("Stack() must capture the creation stack")
	}
}

func TestErrorStringPrefersMessageThenCause(t *testing.T) {
	withMessage := New(http.StatusBadRequest, "title", "message")
	if withMessage.Error() != "message" {
		t.Fatalf("Error() = %q, want message", withMessage.Error())
	}

	cause := errors.New("raw cause")
	withCause := &Error{Status: http.StatusInternalServerError, Cause: cause}
	if withCause.Error() != "raw cause" {
		t.Fatalf("Error() = %q, want raw cause", withCause.Error())
	}

	empty := &Error{Status: http.StatusBadRequest}
	if empty.Error() != http.StatusText(http.StatusBadRequest) {
		t.Fatalf("Error() = %q, want status text", empty.Error())
	}
}

func TestWrapReturnsExistingAppErrorUnchanged(t *testing.T) {
	original := NewNotFound("không thấy")
	wrapped := Wrap(original, "fallback")

	var appErr *Error
	if !errors.As(wrapped, &appErr) {
		t.Fatal("Wrap must return an *Error")
	}
	if appErr != original {
		t.Fatal("Wrap must return the existing apperror unchanged")
	}
}

func TestWrapHidesUnknownErrorBehindGenericMessage(t *testing.T) {
	cause := errors.New("pq: duplicate key value violates unique constraint")
	wrapped := Wrap(cause)

	var appErr *Error
	if !errors.As(wrapped, &appErr) {
		t.Fatal("Wrap must return an *Error")
	}
	if appErr.Status != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want 500", appErr.Status)
	}
	if appErr.Message != GenericInternalMessage {
		t.Fatalf("Message = %q, want generic message", appErr.Message)
	}
	if appErr.Unwrap() != cause {
		t.Fatal("Cause must keep the original error for server-side logging")
	}
	if appErr.Error() == cause.Error() {
		t.Fatal("raw error text must not leak through Error()")
	}
}

func TestWrapUsesFallbackMessageWhenProvided(t *testing.T) {
	wrapped := Wrap(errors.New("db down"), "Không thể xử lý yêu cầu")

	var appErr *Error
	if !errors.As(wrapped, &appErr) {
		t.Fatal("Wrap must return an *Error")
	}
	if appErr.Message != "Không thể xử lý yêu cầu" {
		t.Fatalf("Message = %q", appErr.Message)
	}
	if appErr.Cause == nil {
		t.Fatal("Cause must be preserved")
	}
}

func TestWrapNilReturnsNil(t *testing.T) {
	if Wrap(nil) != nil {
		t.Fatal("Wrap(nil) must return nil")
	}
}

func TestFromMapsInvalidUUIDToBadRequest(t *testing.T) {
	appErr := From(errors.New(`invalid UUID length: 3`))

	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("Status = %d, want 400", appErr.Status)
	}
	if appErr.Message == GenericInternalMessage {
		t.Fatal("invalid UUID must not degrade to the generic message")
	}
}

func TestFromHidesUnknownError(t *testing.T) {
	cause := errors.New("connection refused")
	appErr := From(cause)

	if appErr.Status != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want 500", appErr.Status)
	}
	if appErr.Message != GenericInternalMessage {
		t.Fatalf("Message = %q, want generic message", appErr.Message)
	}
	if appErr.Cause != cause {
		t.Fatal("Cause must keep the original error")
	}
}

func TestFromReturnsExistingAppErrorUnchanged(t *testing.T) {
	original := NewForbidden("cấm")
	if From(original) != original {
		t.Fatal("From must return the existing apperror unchanged")
	}
}

func TestFromNilReturnsNil(t *testing.T) {
	if From(nil) != nil {
		t.Fatal("From(nil) must return nil")
	}
}

func TestIsComparesStatusAndMessage(t *testing.T) {
	a := NewUnauthorized("Vui lòng đăng nhập")
	b := NewUnauthorized("Vui lòng đăng nhập")
	c := NewUnauthorized("Token không hợp lệ")

	if !errors.Is(a, b) {
		t.Fatal("same status and message must match")
	}
	if errors.Is(a, c) {
		t.Fatal("different message must not match")
	}
	if errors.Is(a, errors.New("plain")) {
		t.Fatal("non-apperror target must not match")
	}
}

func TestStaticHelpersUseVietnameseClientMessages(t *testing.T) {
	cases := []struct {
		name   string
		err    *Error
		status int
	}{
		{"ErrInvalidToken", ErrInvalidToken(), http.StatusUnauthorized},
		{"ErrInvalidAccessToken", ErrInvalidAccessToken(), http.StatusUnauthorized},
		{"ErrUnauthorized", ErrUnauthorized(), http.StatusUnauthorized},
		{"ErrForbidden", ErrForbidden(), http.StatusForbidden},
		{"ErrSearchIndexNotFound", ErrSearchIndexNotFound(), http.StatusNotFound},
		{"ErrUnexpectedSigningToken", ErrUnexpectedSigningToken(), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Status != tc.status {
				t.Fatalf("Status = %d, want %d", tc.err.Status, tc.status)
			}
			if tc.err.Message == "" {
				t.Fatal("client-facing message must not be empty")
			}
			if tc.err.Message == "4011" {
				t.Fatal("legacy placeholder message must not return")
			}
		})
	}
}
