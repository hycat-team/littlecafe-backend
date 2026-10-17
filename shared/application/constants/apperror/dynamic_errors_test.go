package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestDynamicFactoriesMapStatusAndTitle(t *testing.T) {
	cases := []struct {
		name   string
		err    *Error
		status int
		title  string
	}{
		{"NewBadRequest", NewBadRequest("detail"), http.StatusBadRequest, "Thao tác không thành công"},
		{"NewNotFound", NewNotFound("detail"), http.StatusNotFound, "Không tìm thấy dữ liệu"},
		{"NewTooManyRequest", NewTooManyRequest("detail"), http.StatusTooManyRequests, "Quá nhiều yêu cầu"},
		{"NewRequestEntityTooLarge", NewRequestEntityTooLarge("detail"), http.StatusRequestEntityTooLarge, "Dữ liệu gửi lên quá lớn"},
		{"NewConflict", NewConflict("detail"), http.StatusConflict, "Dữ liệu bị trùng lặp"},
		{"NewUnauthorized", NewUnauthorized("detail"), http.StatusUnauthorized, "Lỗi xác thực"},
		{"NewForbidden", NewForbidden("detail"), http.StatusForbidden, "Không có quyền truy cập"},
		{"NewInternalError", NewInternalError("detail"), http.StatusInternalServerError, "Lỗi hệ thống"},
		{"NewServiceUnavailable", NewServiceUnavailable("detail"), http.StatusServiceUnavailable, "Dịch vụ tạm thời gián đoạn"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Status != tc.status {
				t.Fatalf("Status = %d, want %d", tc.err.Status, tc.status)
			}
			if tc.err.Title != tc.title {
				t.Fatalf("Title = %q, want %q", tc.err.Title, tc.title)
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
	if err.Message != "Dịch vụ đang bận, vui lòng thử lại sau ít phút." {
		t.Fatalf("Message = %q", err.Message)
	}
	if err.Unwrap() != cause {
		t.Fatal("Cause must keep the provider error for logging")
	}
	if err.Error() == cause.Error() {
		t.Fatal("raw provider error must not leak through Error()")
	}
}
