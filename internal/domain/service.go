package domain

import (
	"context"
	"fmt"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type userService struct {
	repo   internal.UserRepository
	pubsub internal.PubSubService
}

func NewUserService(repo internal.UserRepository, pubsub internal.PubSubService) internal.UserService {
	return &userService{
		repo:   repo,
		pubsub: pubsub,
	}
}

func (us *userService) Add(ctx context.Context, user *types.User) error {
	// this would be the place to call input-validation if this service is responsible for that
	// if err := validateNew(user); err != nil { ... }

	user.CreatedAt = time.Now()
	user.UpdatedAt = nil

	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}

	if err := us.repo.Add(ctx, user); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	if err := us.pubsub.PublishUserChange(types.SubscriptionPayload{UserId: user.Id, Change: types.UserChangeTypeCreated}); err != nil {
		slog.With(slog.Any("error", err)).WarnContext(ctx, "Failed to publish user change")
		// not error-ing out here since the user actually was created
	}

	return nil
}

func (us *userService) UpdatePartial(ctx context.Context, filter types.UserFilter, fields types.UpdateUserFields) error {
	user, err := us.repo.UpdatePartial(ctx, filter, fields)
	if err != nil {
		return err
	}

	if err := us.pubsub.PublishUserChange(types.SubscriptionPayload{UserId: user.Id, Change: types.UserChangeTypeUpdated}); err != nil {
		slog.With(slog.Any("error", err)).WarnContext(ctx, "Failed to publish user change")
		// not error-ing out here since the update actually was done
	}

	return nil
}

func (us *userService) Delete(ctx context.Context, userId uuid.UUID) error {
	if userId == uuid.Nil {
		return fmt.Errorf("failed to delete user: %w", types.ErrInvalidUserId)
	}
	if err := us.repo.Delete(ctx, userId); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := us.pubsub.PublishUserChange(types.SubscriptionPayload{UserId: userId, Change: types.UserChangeTypeDeleted}); err != nil {
		slog.With(slog.Any("error", err)).WarnContext(ctx, "Failed to publish user change")
		// not error-ing out here since the user actually was created
	}

	return nil
}

func (us *userService) List(ctx context.Context, filter types.UserFilter, paging types.Paging) ([]types.User, uint64, error) {
	if paging.Limit == 0 {
		return []types.User{}, 0, nil
	}

	users, total, err := us.repo.List(ctx, filter, paging)
	if err != nil {
		return nil, total, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (us *userService) SubscribeToUserChanges(ctx context.Context, req types.SubscriptionRequest) (<-chan types.SubscriptionPayload, error) {
	return us.pubsub.SubscribeToUserChanges(ctx, req)
}
