package main

import (
	"context"
	"github.com/captainlettuce/users-microservice/cmd"
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/captainlettuce/users-microservice/internal/server"
	"github.com/captainlettuce/users-microservice/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	killSignal := make(chan struct{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		select {
		case <-c:
			close(killSignal)
		case <-killSignal:
			return
		}
	}()

	app := cmd.Bootstrap()
	grpcServer := server.NewGrpc(func(s *grpc.Server, hs *health.Server) {
		generated.RegisterUsersServiceServer(s, app.UserGrpcServer)
		hs.SetServingStatus(types.GrpcServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
		hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	})

	app.AddShutdownFunction(func(_ context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})

	go func(s chan struct{}) {
		listen, err := net.Listen("tcp", ":"+app.GRPCPort)
		if err != nil {
			app.Logger.With("error", err).Error("failed to listen on grpc port")
			close(s)
			return
		}

		app.Logger.With(slog.String("port", app.GRPCPort)).Info("grpc server starting")

		err = grpcServer.Serve(listen)
		if err != nil {
			app.Logger.Error("failed to serve grpc server", slog.Any("error", err))
			close(s)
		}
	}(killSignal)

	<-killSignal
	app.GracefulShutdown()
}
