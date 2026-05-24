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
	duration := time.Since(start)

	logGRPCRequest(ctx, info.FullMethod, code, duration, err)

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

func logGRPCRequest(
	ctx context.Context,
	fullMethod string,
	code codes.Code,
	duration time.Duration,
	err error,
) {
	service, method := splitFullMethod(fullMethod)

	slog.LogAttrs(
		ctx,
		logLevel(code),
		"grpc request",
		slog.String("component", "grpc"),
		slog.String("request_id", requestIDFromContext(ctx)),
		slog.String("grpc_service", service),
		slog.String("grpc_method", method),
		slog.String("grpc_full_method", fullMethod),
		slog.String("grpc_code", code.String()),
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.String("error", errorMessage(err)),
	)
}

func errorMessage(err error) string {
	if err == nil {
		return "-"
	}
	return err.Error()
}

func UnaryRecoveryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			logRecoveredPanic(ctx, info.FullMethod, r)
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}

func logRecoveredPanic(ctx context.Context, fullMethod string, recovered any) {
	service, method := splitFullMethod(fullMethod)

	slog.ErrorContext(
		ctx,
		"grpc panic recovered",
		"component", "grpc",
		"event", "panic",
		"request_id", requestIDFromContext(ctx),
		"grpc_service", service,
		"grpc_method", method,
		"grpc_full_method", fullMethod,
		"grpc_code", codes.Internal.String(),
		"panic", recovered,
		"stack", string(debug.Stack()),
	)
}
