package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type GrpcLoggingInterceptor struct {
	lg Logger
}

func NewGrpcLoggingInterceptor(lg Logger) GrpcLoggingInterceptor {
	return GrpcLoggingInterceptor{lg: lg}
}

func (l GrpcLoggingInterceptor) grpcLoggingMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	latency := time.Since(start)
	l.lg.InfoWithParams("", map[string]string{
		"path":     info.FullMethod,
		"latency":  latency.String(),
		"protocol": "grpc",
	})
	return resp, err
}
