package grpc

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryLoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	code := status.Code(err)
	level := logLevel(code)
	requestID := requestIDFromContext(ctx)
	service, method := splitFullMethod(info.FullMethod)
	durationMS := time.Since(start).Milliseconds()
	errorMessage := "-"
	if err != nil {
		errorMessage = err.Error()
	}

	slog.LogAttrs(
		ctx,
		level,
		"grpc request",
		slog.String("component", "grpc"),
		slog.String("request_id", requestID),
		slog.String("grpc_service", service),
		slog.String("grpc_method", method),
		slog.String("grpc_full_method", info.FullMethod),
		slog.String("grpc_code", code.String()),
		slog.Int64("duration_ms", durationMS),
		slog.String("error", errorMessage),
	)

	return resp, err
}

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "-"
	}

	for _, key := range []string{"x-request-id", "request-id", "trace-id"} {
		values := md.Get(key)
		if len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}

	return "-"
}

func splitFullMethod(fullMethod string) (service string, method string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return "-", "-"
	}

	return parts[0], parts[1]
}

func logLevel(code codes.Code) slog.Level {
	switch code {
	case codes.OK:
		return slog.LevelInfo
	case codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition,
		codes.Aborted, codes.OutOfRange, codes.Unauthenticated:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func UnaryRecoveryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			requestID := requestIDFromContext(ctx)
			service, method := splitFullMethod(info.FullMethod)

			slog.ErrorContext(
				ctx,
				"grpc panic recovered",
				"component", "grpc",
				"event", "panic",
				"request_id", requestID,
				"grpc_service", service,
				"grpc_method", method,
				"grpc_full_method", info.FullMethod,
				"grpc_code", codes.Internal.String(),
				"panic", r,
				"stack", string(debug.Stack()),
			)
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
