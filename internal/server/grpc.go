package server

import (
	"context"
	"github.com/captainlettuce/users-microservice/internal/logging"
	"github.com/google/uuid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"log/slog"
	"runtime/debug"
)

func NewGrpc(configure func(s *grpc.Server, hs *health.Server)) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(

			// recover any uncaught panics
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandler(recoveryHandler),
			),

			logInjectionUnaryServerInterceptor(),

			// add tracing, for example https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc
		),

		grpc.ChainStreamInterceptor(

			// recover any uncaught panics
			recovery.StreamServerInterceptor(
				recovery.WithRecoveryHandler(recoveryHandler),
			),

			logInjectionStreamServerInterceptor(),

			// add tracing, for example https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/google.golang.org/grpc
		),
	)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("grpc.health.v1", grpc_health_v1.HealthCheckResponse_SERVING)

	configure(server, healthServer)

	return server
}

func logInjectionUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = addLoggingAttrsToContext(ctx, info.FullMethod)

		return handler(ctx, req)
	}
}

func logInjectionStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedServ := &middleware.WrappedServerStream{
			ServerStream:   ss,
			WrappedContext: addLoggingAttrsToContext(ss.Context(), info.FullMethod),
		}

		return handler(srv, wrappedServ)
	}
}

func addLoggingAttrsToContext(ctx context.Context, endpoint string) context.Context {
	return logging.LogToContext(
		ctx,
		slog.String("request-id", uuid.New().String()),
		slog.String("endpoint", endpoint),
	)
}

func recoveryHandler(p any) error {
	panicLogger := slog.With(slog.Any("panic", p))
	panicLogger.Warn("panic triggered")

	panicLogger.With(
		slog.String("stack", string(debug.Stack())),
	).Debug("panic debug information")

	return status.Errorf(codes.Internal, "panic triggered: %v", p)
}
