// Package grpcx bridges service errors and gRPC status codes so a contract
// observes the same error shape in-process or remote (Constitution II / V).
// Clients use ToAppError; servers use ToGRPCError.
//
// Wire contract for rich errors:
//   - business `code` travels in google.rpc.ErrorInfo.Reason;
//   - field-level errors (reserved) travel in
//     google.rpc.BadRequest.FieldViolations.
//
// The detail list is additive: a sender may add a new detail type without
// breaking older receivers, which must ignore detail types they do not know.
package grpcx

import (
	"context"
	"errors"
	"net/http"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToAppError maps a gRPC transport error to a centralized apperror.
// Server-produced messages (already localized by the caller's locale) are
// preserved; unrecognized codes degrade to a generic message.
func ToAppError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}

	st, ok := status.FromError(err)
	if !ok {
		return apperror.NewUpstreamUnavailable(err)
	}

	var mapped *apperror.Error
	switch st.Code() {
	case codes.NotFound:
		mapped = apperror.NewNotFound(st.Message())
	case codes.InvalidArgument, codes.FailedPrecondition:
		mapped = apperror.NewBadRequest(st.Message())
	case codes.Unauthenticated:
		mapped = apperror.NewUnauthorized(st.Message())
	case codes.PermissionDenied:
		mapped = apperror.NewForbidden(st.Message())
	case codes.AlreadyExists, codes.Aborted:
		mapped = apperror.NewConflict(st.Message())
	case codes.ResourceExhausted:
		mapped = apperror.NewTooManyRequest(st.Message())
	case codes.Unavailable, codes.DeadlineExceeded:
		mapped = apperror.NewUpstreamUnavailable(err)
	default:
		mapped = apperror.NewInternalError(apperror.GenericInternalMessage)
		mapped.Cause = err
	}

	applyDetails(mapped, st)
	return mapped
}

// ToGRPCError converts a centralized apperror into a gRPC status error using
// the default locale.
func ToGRPCError(err error) error {
	return ToGRPCErrorWithContext(context.Background(), err)
}

// ToGRPCErrorWithContext converts a centralized apperror into a gRPC status
// error rendered for the locale carried by ctx. A non-empty apperror `code` is
// attached as an ErrorInfo detail together with the message key and locale so a
// receiving service can re-render the same message.
func ToGRPCErrorWithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	locale := i18n.FromContext(ctx)

	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		return status.Error(codes.Internal, i18n.Default().T(locale, i18n.ErrorInternalGeneric))
	}

	localized := appErr.Localize(locale)
	code, message := grpcStatusFor(localized)
	st := status.New(code, message)
	if appErr.Code != "" {
		info := &errdetails.ErrorInfo{Reason: appErr.Code}
		if appErr.MessageKey != "" {
			info.Metadata = map[string]string{
				"messageKey": appErr.MessageKey,
				"locale":     string(locale),
			}
		}
		if withDetails, detailErr := st.WithDetails(info); detailErr == nil {
			st = withDetails
		}
	}
	return st.Err()
}

func grpcStatusFor(appErr *apperror.Error) (codes.Code, string) {
	message := appErr.Message
	if message == "" {
		message = i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalGeneric)
	}
	switch appErr.Status {
	case http.StatusNotFound:
		return codes.NotFound, message
	case http.StatusBadRequest:
		return codes.InvalidArgument, message
	case http.StatusUnauthorized:
		return codes.Unauthenticated, message
	case http.StatusForbidden:
		return codes.PermissionDenied, message
	case http.StatusConflict:
		return codes.AlreadyExists, message
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted, message
	case http.StatusServiceUnavailable:
		return codes.Unavailable, message
	default:
		return codes.Internal, i18n.Default().T(i18n.DefaultLocale, i18n.ErrorInternalGeneric)
	}
}

// applyDetails decodes standard gRPC error details into the apperror.
// Unknown detail types are ignored on purpose so a newer sender never breaks
// an older receiver.
func applyDetails(appErr *apperror.Error, st *status.Status) {
	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			if reason := d.GetReason(); reason != "" {
				appErr.Code = reason
			}
		case *errdetails.BadRequest:
			for _, violation := range d.GetFieldViolations() {
				appErr.Errors = append(appErr.Errors, apperror.ValidationErrorItem{
					Field:   violation.GetField(),
					Message: violation.GetDescription(),
				})
			}
		}
	}
}
