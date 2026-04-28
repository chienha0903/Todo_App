package grpc

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	log.Printf(
		"INFO: gRPC method=%s code=%s duration=%s",
		info.FullMethod,
		code.String(),
		time.Since(start),
	)

	return resp, err
}

func UnaryRecoveryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ERROR: gRPC panic method=%s panic=%v stack=%s", info.FullMethod, r, debug.Stack())
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
