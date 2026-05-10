package graphql

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toGraphQLError chuyển gRPC error → Go error (GraphQL sẽ format vào response.errors[]).
// Pattern: giống httpStatusFromGRPCCode trong handler/http/todo_handler.go nhưng
// GraphQL không dùng HTTP status code — lỗi luôn trả trong body JSON.
func ToGraphQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("request timed out")
	}

	st, ok := status.FromError(err)
	if !ok {
		return errors.New("internal server error")
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("not found: %s", st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("invalid argument: %s", st.Message())
	case codes.Unauthenticated:
		return errors.New("unauthorized")
	case codes.PermissionDenied:
		return errors.New("permission denied")
	case codes.DeadlineExceeded:
		return errors.New("request timed out")
	case codes.Unavailable:
		return errors.New("service unavailable")
	default:
		return errors.New("internal server error")
	}
}
