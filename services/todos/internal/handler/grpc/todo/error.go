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

	var appErr *apperrors.AppError
	if !stderrors.As(err, &appErr) {
		return status.Error(codes.Internal, "internal server error")
	}

	return status.Error(toGRPCCode(appErr.Code), appErr.Message)
}

func toGRPCCode(code apperrors.ErrorCode) codes.Code {
	switch code {
	case apperrors.ErrNotFound:
		return codes.NotFound
	case apperrors.ErrInvalidParameter:
		return codes.InvalidArgument
	case apperrors.ErrAuthN:
		return codes.Unauthenticated
	case apperrors.ErrAuthZ:
		return codes.PermissionDenied
	case apperrors.ErrInternal:
		return codes.Internal
	default:
		return codes.Internal
	}
}
