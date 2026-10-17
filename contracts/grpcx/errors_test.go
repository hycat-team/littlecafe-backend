package grpcx

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func genericText() string {
	return i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalGeneric)
}

func TestToGRPCErrorMapsStatusAndMessage(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		want    codes.Code
		message string
	}{
		{"nil", nil, codes.OK, ""},
		{"plain error", errors.New("boom"), codes.Internal, genericText()},
		{"bad request", apperror.NewBadRequest("Dữ liệu không hợp lệ"), codes.InvalidArgument, "Dữ liệu không hợp lệ"},
		{"unauthorized", apperror.NewUnauthorized("Vui lòng đăng nhập"), codes.Unauthenticated, "Vui lòng đăng nhập"},
		{"forbidden", apperror.NewForbidden("Không có quyền"), codes.PermissionDenied, "Không có quyền"},
		{"not found", apperror.NewNotFound("Không tìm thấy"), codes.NotFound, "Không tìm thấy"},
		{"conflict", apperror.NewConflict("Bị trùng"), codes.AlreadyExists, "Bị trùng"},
		{"too many", apperror.NewTooManyRequest("Quá nhanh"), codes.ResourceExhausted, "Quá nhanh"},
		{"unavailable", apperror.NewServiceUnavailable("Bảo trì"), codes.Unavailable, "Bảo trì"},
		{"internal degrades to generic", apperror.NewInternalError("chi tiết nội bộ"), codes.Internal, genericText()},
		{"unmapped status degrades to generic", apperror.NewError(http.StatusTeapot, "Trà", "chi tiết"), codes.Internal, genericText()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				if got := ToGRPCError(nil); got != nil {
					t.Fatalf("ToGRPCError(nil) = %v, want nil", got)
				}
				return
			}

			got := ToGRPCError(tc.err)
			st := status.Convert(got)
			if st.Code() != tc.want {
				t.Fatalf("code = %v, want %v", st.Code(), tc.want)
			}
			if st.Message() != tc.message {
				t.Fatalf("message = %q, want %q", st.Message(), tc.message)
			}
		})
	}
}

func TestToGRPCErrorCarriesCodeAsErrorInfo(t *testing.T) {
	err := apperror.NewNotFound("Không tìm thấy đơn hàng")
	err.Code = "ORDER_0001"

	st := status.Convert(ToGRPCError(err))

	reason := errorInfoReason(t, st)
	if reason != "ORDER_0001" {
		t.Fatalf("ErrorInfo.Reason = %q, want ORDER_0001", reason)
	}
	if st.Message() != "Không tìm thấy đơn hàng" {
		t.Fatalf("message = %q", st.Message())
	}
}

func TestToGRPCErrorWithContextLocalizesAndCarriesMetadata(t *testing.T) {
	err := apperror.NewNotFound(i18n.ErrorSearchIndexNotFound)
	err.Code = "ORDER_0001"

	ctx := i18n.WithLocale(context.Background(), i18n.EN)
	st := status.Convert(ToGRPCErrorWithContext(ctx, err))

	if st.Message() != i18n.Default().T(i18n.EN, i18n.ErrorSearchIndexNotFound) {
		t.Fatalf("message = %q", st.Message())
	}

	info := errorInfo(t, st)
	if info.GetReason() != "ORDER_0001" {
		t.Fatalf("Reason = %q", info.GetReason())
	}
	if info.GetMetadata()["messageKey"] != i18n.ErrorSearchIndexNotFound {
		t.Fatalf("messageKey = %q", info.GetMetadata()["messageKey"])
	}
	if info.GetMetadata()["locale"] != string(i18n.EN) {
		t.Fatalf("locale = %q", info.GetMetadata()["locale"])
	}
}

func TestToGRPCErrorOmitsDetailsWithoutCode(t *testing.T) {
	st := status.Convert(ToGRPCError(apperror.NewBadRequest("Dữ liệu không hợp lệ")))

	if len(st.Details()) != 0 {
		t.Fatalf("details = %v, want none when code is empty", st.Details())
	}
}

func TestRoundTripPreservesCodeStatusAndMessage(t *testing.T) {
	original := apperror.NewConflict("Dữ liệu bị trùng lặp")
	original.Code = "WALLET_0007"

	var roundTripped *apperror.Error
	if !errors.As(ToAppError(ToGRPCError(original)), &roundTripped) {
		t.Fatal("round trip must produce an apperror")
	}

	if roundTripped.Status != http.StatusConflict {
		t.Fatalf("Status = %d, want 409", roundTripped.Status)
	}
	if roundTripped.Code != "WALLET_0007" {
		t.Fatalf("Code = %q, want WALLET_0007", roundTripped.Code)
	}
	if roundTripped.Detail != "Dữ liệu bị trùng lặp" {
		t.Fatalf("Detail = %q", roundTripped.Detail)
	}
}

func TestToAppErrorMapsGRPCStatuses(t *testing.T) {
	cases := []struct {
		name   string
		code   codes.Code
		status int
	}{
		{"not found", codes.NotFound, http.StatusNotFound},
		{"invalid argument", codes.InvalidArgument, http.StatusBadRequest},
		{"failed precondition", codes.FailedPrecondition, http.StatusBadRequest},
		{"unauthenticated", codes.Unauthenticated, http.StatusUnauthorized},
		{"permission denied", codes.PermissionDenied, http.StatusForbidden},
		{"already exists", codes.AlreadyExists, http.StatusConflict},
		{"aborted", codes.Aborted, http.StatusConflict},
		{"resource exhausted", codes.ResourceExhausted, http.StatusTooManyRequests},
		{"unavailable", codes.Unavailable, http.StatusServiceUnavailable},
		{"deadline exceeded", codes.DeadlineExceeded, http.StatusServiceUnavailable},
		{"unknown code", codes.DataLoss, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var appErr *apperror.Error
			if !errors.As(ToAppError(status.Error(tc.code, "lỗi")), &appErr) {
				t.Fatal("ToAppError must produce an apperror")
			}
			if appErr.Status != tc.status {
				t.Fatalf("Status = %d, want %d", appErr.Status, tc.status)
			}
		})
	}
}

func TestToAppErrorIgnoresUnknownDetailsAndDecodesFieldViolations(t *testing.T) {
	st, detailErr := status.New(codes.InvalidArgument, "Dữ liệu không hợp lệ").WithDetails(
		&errdetails.ErrorInfo{Reason: "ORDER_0002"},
		&errdetails.DebugInfo{Detail: "internal debug payload"},
		&errdetails.BadRequest{FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{Field: "quantity", Description: "phải lớn hơn 0"},
		}},
	)
	if detailErr != nil {
		t.Fatalf("WithDetails: %v", detailErr)
	}

	var appErr *apperror.Error
	if !errors.As(ToAppError(st.Err()), &appErr) {
		t.Fatal("ToAppError must produce an apperror")
	}

	if appErr.Code != "ORDER_0002" {
		t.Fatalf("Code = %q, want ORDER_0002", appErr.Code)
	}
	if len(appErr.Errors) != 1 {
		t.Fatalf("Errors = %v, want one field violation", appErr.Errors)
	}
	if appErr.Errors[0].Field != "quantity" || appErr.Errors[0].Message != "phải lớn hơn 0" {
		t.Fatalf("Errors[0] = %+v", appErr.Errors[0])
	}
	if appErr.Detail != "Dữ liệu không hợp lệ" {
		t.Fatalf("Detail = %q", appErr.Detail)
	}
}

func TestToAppErrorNonStatusErrorBecomesUpstreamUnavailable(t *testing.T) {
	cause := errors.New("connection refused")
	var appErr *apperror.Error
	if !errors.As(ToAppError(cause), &appErr) {
		t.Fatal("ToAppError must produce an apperror")
	}

	if appErr.Status != http.StatusServiceUnavailable {
		t.Fatalf("Status = %d, want 503", appErr.Status)
	}
	if appErr.Unwrap() != cause {
		t.Fatal("Cause must keep the transport error")
	}
}

func TestToAppErrorPassthroughAndNil(t *testing.T) {
	original := apperror.NewBadRequest("x")
	if ToAppError(original) != original {
		t.Fatal("existing apperror must be returned unchanged")
	}
	if ToAppError(nil) != nil {
		t.Fatal("ToAppError(nil) must return nil")
	}
}

func errorInfoReason(t *testing.T, st *status.Status) string {
	t.Helper()
	return errorInfo(t, st).GetReason()
}

func errorInfo(t *testing.T, st *status.Status) *errdetails.ErrorInfo {
	t.Helper()
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			return info
		}
	}
	return &errdetails.ErrorInfo{}
}
