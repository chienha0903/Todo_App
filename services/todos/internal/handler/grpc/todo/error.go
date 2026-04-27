package todo

import (
	stderrors "errors"

	apperrors "github.com/chienha0903/Todo_App/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	var appErr *apperrors.Error
	if !stderrors.As(err, &appErr) {
		return status.Error(codes.Internal, "internal server error")
	}

	return status.Error(toGRPCCode(appErr.Reason), appErr.Message)
}

func toGRPCCode(reason apperrors.Reason) codes.Code {
	switch reason {
	case apperrors.REASON_NOT_FOUND:
		return codes.NotFound
	case apperrors.REASON_INVALID_PARAMETER:
		return codes.InvalidArgument
	case apperrors.REASON_UNAUTHORIZED:
		return codes.Unauthenticated
	case apperrors.REASON_PERMISSION_DENIED:
		return codes.PermissionDenied
	case apperrors.REASON_INTERNAL_SERVER_ERROR:
		return codes.Internal
	default:
		return codes.Internal
	}
}
