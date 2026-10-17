package errorutils

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

const projectFramePrefix = "github.com/hycat-team/littlecafe-backend"

func TestFilterStackTraceArrayDropsInternalAndNoiseFrames(t *testing.T) {
	raw := strings.Join([]string{
		"goroutine 1 [running]:",
		projectFramePrefix + "/shared/presentation/middleware.GlobalErrorHandler.func1()",
		"\t/workspace/shared/presentation/middleware/error_handler.go:62 +0x1",
		projectFramePrefix + "/shared/application/constants/apperror.NewBadRequest()",
		"\t/workspace/shared/application/constants/apperror/dynamic_errors.go:10 +0x2",
		projectFramePrefix + "/services/order/internal/modules/order/application/usecase.PlaceOrder()",
		"\t/workspace/services/order/internal/modules/order/application/usecase/place_order.go:20 +0x3",
		"github.com/google/uuid.New()",
		"\t/other/uuid.go:5 +0x4",
	}, "\n")

	frames := FilterStackTraceArray(raw)

	if len(frames) != 2 {
		t.Fatalf("frames = %v, want goroutine line plus one project frame", frames)
	}
	if !strings.HasPrefix(frames[0], "goroutine") {
		t.Fatalf("first frame = %q, want goroutine header", frames[0])
	}
	if !strings.Contains(frames[1], "PlaceOrder") || !strings.Contains(frames[1], "place_order.go:20") {
		t.Fatalf("second frame = %q, want the usecase frame", frames[1])
	}
	for _, frame := range frames {
		if strings.Contains(frame, "apperror") {
			t.Fatalf("apperror frames must be filtered, got %q", frame)
		}
		if strings.Contains(frame, "GlobalErrorHandler") {
			t.Fatalf("middleware noise must be filtered when a real frame exists, got %q", frame)
		}
	}
}

func TestFilterStackTraceArrayDeduplicatesFrames(t *testing.T) {
	raw := strings.Join([]string{
		"goroutine 7 [running]:",
		projectFramePrefix + "/services/order/internal/modules/order/application/usecase.PlaceOrder()",
		"\t/workspace/services/order/internal/modules/order/application/usecase/place_order.go:20 +0x1",
		projectFramePrefix + "/services/order/internal/modules/order/application/usecase.PlaceOrder()",
		"\t/workspace/services/order/internal/modules/order/application/usecase/place_order.go:20 +0x1",
	}, "\n")

	frames := FilterStackTraceArray(raw)

	if len(frames) != 2 {
		t.Fatalf("frames = %v, want duplicate collapsed", frames)
	}
}

func TestFilterStackTraceArrayFallsBackWhenNoProjectFrame(t *testing.T) {
	raw := strings.Join([]string{
		"no goroutine header",
		"runtime.main()",
		"\t/usr/local/go/src/runtime/proc.go:283 +0x1",
	}, "\n")

	frames := FilterStackTraceArray(raw)

	if len(frames) != 1 || frames[0] != "No project-level stack trace available." {
		t.Fatalf("frames = %v, want fallback message", frames)
	}
}

func TestPrimaryStackFrameReturnsFirstProjectFrame(t *testing.T) {
	raw := strings.Join([]string{
		"goroutine 1 [running]:",
		projectFramePrefix + "/services/order/internal/modules/order/application/usecase.PlaceOrder()",
		"\t/workspace/services/order/internal/modules/order/application/usecase/place_order.go:20 +0x1",
	}, "\n")

	frame := PrimaryStackFrame(raw)

	if !strings.Contains(frame, "PlaceOrder") {
		t.Fatalf("PrimaryStackFrame = %q, want usecase frame", frame)
	}
}

func TestToAppError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if ToAppError(nil) != nil {
			t.Fatal("ToAppError(nil) must return nil")
		}
	})

	t.Run("existing apperror is preserved", func(t *testing.T) {
		original := apperror.NewBadRequest("x")
		if ToAppError(original) != original {
			t.Fatal("existing apperror must be returned unchanged")
		}
	})

	t.Run("fallback message is applied", func(t *testing.T) {
		appErr := ToAppError(errors.New("db down"), "Không thể xử lý yêu cầu")

		if appErr.Status != http.StatusInternalServerError {
			t.Fatalf("Status = %d, want 500", appErr.Status)
		}
		if appErr.Message != "Không thể xử lý yêu cầu" {
			t.Fatalf("Message = %q", appErr.Message)
		}
		if appErr.Cause == nil {
			t.Fatal("Cause must be preserved")
		}
	})

	t.Run("unknown error degrades to generic", func(t *testing.T) {
		appErr := ToAppError(errors.New("boom"))

		if appErr.Message != i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalGeneric) {
			t.Fatalf("Message = %q, want generic message", appErr.Message)
		}
		if appErr.MessageKey != apperror.GenericInternalMessage {
			t.Fatalf("MessageKey = %q, want generic key", appErr.MessageKey)
		}
	})
}

func TestWrapErrorKeepsCause(t *testing.T) {
	cause := errors.New("boom")
	wrapped := WrapError(cause, "Thông báo")

	var appErr *apperror.Error
	if !errors.As(wrapped, &appErr) {
		t.Fatal("WrapError must return an apperror")
	}
	if appErr.Unwrap() != cause {
		t.Fatal("Cause must be preserved")
	}
}
