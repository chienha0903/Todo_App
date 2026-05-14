package resolver

import (
	stderrors "errors"
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/chienha0903/Todo_App/services/todo-bff/internal/apperror"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// Ưu tiên 1: AppError từ service/resolver (validation, business logic)
	var appErr *apperror.Error
	if stderrors.As(err, &appErr) {
		return &gqlerror.Error{
			Message: appErr.Message,
			Extensions: map[string]interface{}{
				"code": appErrCodeToString(appErr.Code),
			},
		}
	}

	// Ưu tiên 2: gRPC status error truyền qua từ gateway
	for e := err; e != nil; e = stderrors.Unwrap(e) {
		if st, ok := status.FromError(e); ok && st.Code() != codes.OK {
			return &gqlerror.Error{
				Message: st.Message(),
				Extensions: map[string]interface{}{
					"code": grpcCodeToString(st.Code()),
				},
			}
		}
	}

	// Fallback
	return &gqlerror.Error{
		Message: "internal server error",
		Extensions: map[string]interface{}{
			"code": "INTERNAL",
		},
	}
}

func appErrCodeToString(code apperror.Code) string {
	switch code {
	case apperror.CodeNotFound:
		return "NOT_FOUND"
	case apperror.CodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case apperror.CodeUnauthorized:
		return "UNAUTHENTICATED"
	case apperror.CodePermissionDenied:
		return "FORBIDDEN"
	case apperror.CodeTimeout:
		return "TIMEOUT"
	case apperror.CodeUnavailable:
		return "UNAVAILABLE"
	default:
		return "INTERNAL"
	}
}

func grpcCodeToString(code codes.Code) string {
	switch code {
	case codes.NotFound:
		return "NOT_FOUND"
	case codes.InvalidArgument:
		return "INVALID_ARGUMENT"
	case codes.Unauthenticated:
		return "UNAUTHENTICATED"
	case codes.PermissionDenied:
		return "FORBIDDEN"
	case codes.AlreadyExists:
		return "ALREADY_EXISTS"
	case codes.DeadlineExceeded:
		return "TIMEOUT"
	case codes.Unavailable:
		return "UNAVAILABLE"
	default:
		return "INTERNAL"
	}
}

// compile-time check
var _ graphql.ErrorPresenterFunc = ErrorPresenter
