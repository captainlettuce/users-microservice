package internal

import (
	"context"
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type Application struct {
	Logger *slog.Logger

	Domain         UserService
	Repository     UserRepository
	UserGrpcServer generated.UsersServiceServer
	PubSub         PubSubService

	GRPCPort string

	// ShutdownTimeout represents how long to wait for o
	ShutdownTimout    time.Duration
	shutdownFunctions []func(ctx context.Context) error
}

type UserRepository interface {

	// Shutdown is run before app close and can be used to release resources
	Shutdown(ctx context.Context) error

	// Add adds a new user to the database
	Add(ctx context.Context, user *types.User) error

	// UpdatePartial updates only the fields set to a non-nil pointer
	// If a field is set to a pointer to the corresponding types zero-value; the field will be unset
	UpdatePartial(ctx context.Context, filter types.UserFilter, fields types.UpdateUserFields) (types.User, error)

	// Delete permanently removes a user
	// returns nil if the userId is not found
	Delete(ctx context.Context, userId uuid.UUID) error

	// List filtered users
	// takes a set of filters and an offset/limit-based paging entity
	// returns a slice of users along with a total count for the executed filter
	// keeping track of pages and such is up to the caller
	// the results are sorted by when they were inserted to DB in FIFO ordering
	List(ctx context.Context, filter types.UserFilter, paging types.Paging) (users []types.User, totalCount uint64, err error)
}

type UserService interface {
	// Add a new user
	Add(ctx context.Context, user *types.User) error

	// Update an existing user, returns an error if no user found using the filter provided
	UpdatePartial(ctx context.Context, filter types.UserFilter, fields types.UpdateUserFields) error

	// Delete an existing user, returns nil if the user does not exist
	Delete(ctx context.Context, userId uuid.UUID) error

	// List filtered users
	// takes a set of filters and an offset/limit-based paging entity
	// returns a slice of users along with a total count for the executed filter
	// keeping track of pages and such is up to the caller
	// the results are sorted by when they were inserted to DB in FIFO ordering
	List(ctx context.Context, filter types.UserFilter, paging types.Paging) (users []types.User, totalCount uint64, err error)

	// SubscribeToUserChanges returns a channel that receives a message each time a user is updated
	SubscribeToUserChanges(ctx context.Context, req types.SubscriptionRequest) (<-chan types.SubscriptionPayload, error)
}

type PubSubService interface {
	// SubscribeToUserChange returns a channel that receives a message each time a user is updated
	SubscribeToUserChanges(ctx context.Context, req types.SubscriptionRequest) (<-chan types.SubscriptionPayload, error)

	// PublishUserChange publishes a user change to subscribers
	PublishUserChange(result types.SubscriptionPayload) error

	GracefulShutdown(ctx context.Context) error
}

// AddShutdownFunction adds a hook to be run before application exit
// it can be used to release resources and gracefully terminate connections for example
func (app *Application) AddShutdownFunction(function func(context.Context) error) {
	app.shutdownFunctions = append(app.shutdownFunctions, function)
}

// GracefulShutdown runs all functions added with AddShutdownFunction in LIFO order
func (app *Application) GracefulShutdown() {
	shutdownDone := make(chan struct{})
	timeoutCtx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimout)
	defer cancel()

	go func() {

		// Iterate backwards for LIFO execution order
		for i := len(app.shutdownFunctions) - 1; i >= 0; i-- {
			err := app.shutdownFunctions[i](timeoutCtx)
			if err != nil {
				slog.With(slog.Any("error", err)).Warn("Got error releasing resources on shutdown")
			}
		}
		shutdownDone <- struct{}{}
	}()

	select {
	case <-shutdownDone:
		return
	case <-timeoutCtx.Done():
		slog.Warn("Timed out waiting for shutdown to complete")
		return
	}
}
