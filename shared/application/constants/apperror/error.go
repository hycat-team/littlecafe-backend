package apperror

import (
	"errors"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

type ValidationErrorItem struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// GenericInternalMessage is the message key for unrecognized errors. The
// client-facing text is resolved from the i18n catalog.
const GenericInternalMessage = i18n.ErrorInternalGeneric

// PanicRecoveryMessage is the message key for recovered panics. It stays
// distinct from GenericInternalMessage because a panic signals a crash-level
// failure where directing the user to the administrator is apt.
const PanicRecoveryMessage = i18n.ErrorInternalPanic

// Error is the RFC 9457 Problem Details payload. TitleKey/MessageKey hold the
// locale-agnostic catalog keys; Title/Detail/Message hold the default-locale
// rendering so logging and non-localized callers keep working. Use Localize to
// render the payload for a request locale.
type Error struct {
	Type         string                `json:"type"`
	Title        string                `json:"title"`
	Status       int                   `json:"status"`
	Detail       string                `json:"detail,omitempty"`
	Instance     string                `json:"instance,omitempty"`
	Code         string                `json:"code,omitempty"`
	TraceID      string                `json:"traceId,omitempty"`
	Errors       []ValidationErrorItem `json:"errors,omitempty"`
	DebugMessage string                `json:"debugMessage,omitempty"`
	StackTrace   any                   `json:"stackTrace,omitempty"`
	Message      string                `json:"-"`

	TitleKey    string `json:"-"`
	MessageKey  string `json:"-"`
	MessageArgs []any  `json:"-"`

	Cause    error  `json:"-"`
	rawStack string `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return http.StatusText(e.Status)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) Stack() string {
	if e == nil {
		return ""
	}
	return e.rawStack
}

func (e *Error) WithStackTrace(stackTrace any) Error {
	res := *e
	res.StackTrace = stackTrace
	return res
}

func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if e.Status != t.Status {
		return false
	}
	if e.MessageKey != "" || t.MessageKey != "" {
		return e.MessageKey == t.MessageKey
	}
	return e.Message == t.Message
}

// Localize returns a copy rendered for loc. Keys resolve through the shared
// catalog; a key that is not registered is treated as a literal message so
// passthrough strings from remote services keep their original text.
func (e *Error) Localize(loc i18n.Locale) *Error {
	if e == nil {
		return nil
	}
	res := *e
	catalog := i18n.Default()
	loc = i18n.Normalize(loc)

	if e.TitleKey != "" {
		res.Title = catalog.T(loc, e.TitleKey)
	}
	if e.MessageKey != "" {
		res.Message = catalog.T(loc, e.MessageKey, e.MessageArgs...)
		res.Detail = res.Message
	} else if res.Detail == "" {
		res.Detail = res.Message
	}
	return &res
}

// New builds an error from catalog keys. Keys not present in the catalog are
// kept verbatim, which keeps remote/literal messages intact.
func New(status int, titleKey, messageKey string, args ...any) *Error {
	e := &Error{
		Type:        "about:blank",
		Status:      status,
		TitleKey:    titleKey,
		MessageKey:  messageKey,
		MessageArgs: args,
		rawStack:    string(debug.Stack()),
	}
	return e.Localize(i18n.DefaultLocale)
}

// Wrap turns an unrecognized error into a 500 and keeps the raw text only in
// Cause for server-side logging. fallbackKey is an optional catalog key.
func Wrap(err error, fallbackKey ...string) error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	key := GenericInternalMessage
	if len(fallbackKey) > 0 && fallbackKey[0] != "" {
		key = fallbackKey[0]
	}

	e := New(http.StatusInternalServerError, i18n.ErrorTitleInternal, key)
	e.Cause = err
	e.rawStack = string(debug.Stack())
	return e
}

func From(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	if strings.Contains(err.Error(), "invalid UUID") {
		return NewBadRequest(i18n.ErrorInvalidUUID)
	}

	// Unrecognized errors must not surface their raw text to the client; the
	// original error stays available as Cause for server-side logging.
	unrecognized := NewInternalError(GenericInternalMessage)
	unrecognized.Cause = err
	return unrecognized
}
