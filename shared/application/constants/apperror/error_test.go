package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

func msg(key string, args ...any) string {
	return i18n.Default().T(i18n.DefaultLocale, key, args...)
}

func TestNewSetsRFC9457Defaults(t *testing.T) {
	err := New(http.StatusBadRequest, i18n.ErrorTitleBadRequest, i18n.ErrorInvalidUUID)

	if err.Type != "about:blank" {
		t.Fatalf("Type = %q, want about:blank", err.Type)
	}
	if err.Status != http.StatusBadRequest {
		t.Fatalf("Status = %d, want %d", err.Status, http.StatusBadRequest)
	}
	if err.Title != msg(i18n.ErrorTitleBadRequest) {
		t.Fatalf("Title = %q", err.Title)
	}
	if err.Detail != msg(i18n.ErrorInvalidUUID) {
		t.Fatalf("Detail = %q", err.Detail)
	}
	if err.Message != msg(i18n.ErrorInvalidUUID) {
		t.Fatalf("Message = %q", err.Message)
	}
	if err.TitleKey != i18n.ErrorTitleBadRequest || err.MessageKey != i18n.ErrorInvalidUUID {
		t.Fatalf("keys not preserved: title=%q message=%q", err.TitleKey, err.MessageKey)
	}
	if err.Stack() == "" {
		t.Fatal("Stack() must capture the creation stack")
	}
}

func TestLocalizeRendersRequestedLocale(t *testing.T) {
	err := NewNotFound(i18n.ErrorSearchIndexNotFound)

	vi := err.Localize(i18n.VI)
	en := err.Localize(i18n.EN)

	if vi.Detail != msg(i18n.ErrorSearchIndexNotFound) {
		t.Fatalf("vi detail = %q", vi.Detail)
	}
	if en.Detail != i18n.Default().T(i18n.EN, i18n.ErrorSearchIndexNotFound) {
		t.Fatalf("en detail = %q", en.Detail)
	}
	if en.Title != i18n.Default().T(i18n.EN, i18n.ErrorTitleNotFound) {
		t.Fatalf("en title = %q", en.Title)
	}
	if err.Detail != vi.Detail {
		t.Fatal("Localize must not mutate the original error")
	}
}

func TestErrorStringPrefersMessageThenCause(t *testing.T) {
	withMessage := New(http.StatusBadRequest, i18n.ErrorTitleBadRequest, i18n.ErrorInvalidUUID)
	if withMessage.Error() != msg(i18n.ErrorInvalidUUID) {
		t.Fatalf("Error() = %q", withMessage.Error())
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
	original := NewNotFound(i18n.ErrorSearchIndexNotFound)
	wrapped := Wrap(original, i18n.ErrorInternalGeneric)

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
	if appErr.MessageKey != GenericInternalMessage {
		t.Fatalf("MessageKey = %q, want generic key", appErr.MessageKey)
	}
	if appErr.Message != msg(i18n.ErrorInternalGeneric) {
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
	wrapped := Wrap(errors.New("db down"), i18n.ErrorUpstreamUnavailable)

	var appErr *Error
	if !errors.As(wrapped, &appErr) {
		t.Fatal("Wrap must return an *Error")
	}
	if appErr.Message != msg(i18n.ErrorUpstreamUnavailable) {
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
	if appErr.MessageKey != i18n.ErrorInvalidUUID {
		t.Fatalf("MessageKey = %q, want invalid uuid key", appErr.MessageKey)
	}
}

func TestFromHidesUnknownError(t *testing.T) {
	cause := errors.New("connection refused")
	appErr := From(cause)

	if appErr.Status != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want 500", appErr.Status)
	}
	if appErr.MessageKey != GenericInternalMessage {
		t.Fatalf("MessageKey = %q, want generic key", appErr.MessageKey)
	}
	if appErr.Cause != cause {
		t.Fatal("Cause must keep the original error")
	}
}

func TestFromReturnsExistingAppErrorUnchanged(t *testing.T) {
	original := NewForbidden(i18n.ErrorForbiddenAccess)
	if From(original) != original {
		t.Fatal("From must return the existing apperror unchanged")
	}
}

func TestFromNilReturnsNil(t *testing.T) {
	if From(nil) != nil {
		t.Fatal("From(nil) must return nil")
	}
}

func TestIsComparesStatusAndMessageKey(t *testing.T) {
	a := NewUnauthorized(i18n.ErrorAuthRequired)
	b := NewUnauthorized(i18n.ErrorAuthRequired)
	c := NewUnauthorized(i18n.ErrorAuthInvalidToken)

	if !errors.Is(a, b) {
		t.Fatal("same status and message key must match")
	}
	if errors.Is(a, c) {
		t.Fatal("different message key must not match")
	}
	if errors.Is(a, errors.New("plain")) {
		t.Fatal("non-apperror target must not match")
	}
}

func TestStaticHelpersUseLocalizedClientMessages(t *testing.T) {
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
			if tc.err.MessageKey == "" {
				t.Fatal("message key must not be empty")
			}
			if tc.err.Message == "4011" {
				t.Fatal("legacy placeholder message must not return")
			}
		})
	}
}
