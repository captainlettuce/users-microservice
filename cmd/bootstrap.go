package cmd

import (
	"context"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/domain"
	"github.com/captainlettuce/users-microservice/internal/logging"
	"github.com/captainlettuce/users-microservice/internal/pubsub"
	"github.com/captainlettuce/users-microservice/internal/repository"
	"github.com/captainlettuce/users-microservice/internal/server/service"
	"github.com/captainlettuce/users-microservice/internal/types"
	"log/slog"
	"os"
	"strconv"
	"time"
)

func Bootstrap() *internal.Application {

	var loggingOpts []logging.Option

	debugOpt, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if os.Getenv("DEBUG") != "" && err != nil {
		slog.With("debugOpt", os.Getenv("DEBUG")).Warn("Invalid debug environment variable supplied")
	}
	if debugOpt {
		loggingOpts = append(loggingOpts, logging.WithLevel(slog.LevelDebug))
	}

	if handlerOpt := os.Getenv("LOG_FORMAT"); handlerOpt != "" {
		loggingOpts = append(loggingOpts, logging.WithHandler(logging.HandlerType(handlerOpt)))
	}

	logger, err := logging.New(types.ServiceName, loggingOpts...)
	if err != nil {
		panic(err)
	}

	slog.SetDefault(logger)

	app := &internal.Application{
		Logger:         logger,
		GRPCPort:       "8000",
		ShutdownTimout: time.Second * 5,
	}

	var (
		uri        = "mongodb://mongo:27017"
		dbName     = "users"
		collection = "users"
	)

	if u := os.Getenv("MONGO_URI"); u != "" {
		uri = u
	}
	if u := os.Getenv("MONGO_DB"); u != "" {
		dbName = u
	}
	if u := os.Getenv("MONGO_COLLECTION"); u != "" {
		collection = u
	}

	app.Repository, err = repository.NewMongoRepository(context.TODO(), uri, dbName, collection)
	if err != nil {
		app.Logger.With(slog.Any("error", err)).Error("could not connect to repository")
		app.GracefulShutdown()
		os.Exit(1)
	}

	if i, err := strconv.ParseInt(os.Getenv("SHUTDOWN_GRACE"), 10, 64); err == nil {
		app.ShutdownTimout = time.Duration(i) * time.Second
	}
	app.AddShutdownFunction(app.Repository.Shutdown)

	var natsUri = "nats://nats:4222"
	if n := os.Getenv("NATS_URI"); n != "" {
		natsUri = n
	}
	app.PubSub, err = pubsub.NewNatsClient(app.Logger, natsUri)
	if err != nil {
		app.Logger.With(slog.Any("error", err)).Error("could not connect to nats server")
		app.GracefulShutdown()
		os.Exit(1)
	}
	app.AddShutdownFunction(app.PubSub.GracefulShutdown)

	app.Domain = domain.NewUserService(app.Repository, app.PubSub)

	if port := os.Getenv("GRPC_PORT"); port != "" {
		app.GRPCPort = port
	}

	app.UserGrpcServer = service.NewUsersGrpc(app.Domain, app.Logger)

	return app
}
