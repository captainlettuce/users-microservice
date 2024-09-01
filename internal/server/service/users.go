package service

import (
	"context"
	"errors"
	"github.com/captainlettuce/field_mask"
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type usersGrpc struct {
	generated.UnimplementedUsersServiceServer
	service internal.UserService
	logger  *slog.Logger
}

func NewUsersGrpc(service internal.UserService, logger *slog.Logger) generated.UsersServiceServer {
	return &usersGrpc{
		service: service,
		logger:  logger.With(slog.String("component", "grpc.service")),
	}
}

func (u *usersGrpc) Add(ctx context.Context, req *generated.AddUserRequest) (*generated.AddUserResponse, error) {
	user, err := types.UserFromProto(req.User)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = u.service.Add(ctx, &user)
	if err != nil {
		if errors.Is(err, types.ErrInvalidUserId) || errors.Is(err, types.ErrDuplicateUserId) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		u.logger.With(
			slog.Any("error", err),
			slog.Any("userId", user.Id),
		).WarnContext(ctx, "Got unexpected error deleting user")

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &generated.AddUserResponse{User: user.Proto()}, nil
}

func (u *usersGrpc) Update(ctx context.Context, req *generated.UpdateUserRequest) (*generated.UpdateUserResponse, error) {
	var (
		updateRequest types.UpdateUserFields
		filter        types.UserFilter
		err           error
	)

	if err = field_mask.Apply(req.UpdateMask, req.User, &updateRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if filter, err = types.UserFilterFromProto(req.Filter); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = u.service.UpdatePartial(ctx, filter, updateRequest); err != nil {
		if errors.Is(err, types.ErrNotFound) {

			return nil, status.Error(codes.NotFound, err.Error())
		}

		u.logger.With(
			slog.Any("error", err),
			slog.Any("filter", updateRequest),
		).WarnContext(ctx, "Got unexpected error updating user")

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &generated.UpdateUserResponse{}, nil
}

func (u *usersGrpc) Delete(ctx context.Context, req *generated.DeleteUserRequest) (*generated.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = u.service.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, types.ErrInvalidUserId) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		u.logger.With(
			slog.Any("error", err),
			slog.Any("req", req),
		).WarnContext(ctx, "Got unexpected error deleting user")

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &generated.DeleteUserResponse{}, nil
}

func (u *usersGrpc) List(ctx context.Context, req *generated.ListUsersRequest) (*generated.ListUsersResponse, error) {

	filters, err := types.UserFilterFromProto(req.GetFilters())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	users, total, err := u.service.List(ctx, filters, types.PagingFromProto(req.GetPaging()))
	if err != nil {
		u.logger.With(
			slog.Any("error", err),
			slog.Any("req", req),
		).WarnContext(ctx, "Got unexpected error listing users")
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &generated.ListUsersResponse{
		Paging: &generated.PagingMetadata{
			Count: total,
		},
	}

	for _, u := range users {
		resp.Users = append(resp.Users, u.Proto())
	}

	return resp, nil
}

func (u *usersGrpc) Subscribe(req *generated.SubscriptionRequest, serv generated.UsersService_SubscribeServer) error {

	r, err := types.SubscriptionRequestFromProto(req)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := serv.Context()
	ch, err := u.service.SubscribeToUserChanges(ctx, r)
	if err != nil {
		u.logger.With(slog.Any("error", err)).WarnContext(ctx, "Got unexpected error subscribing to user updates")
		return status.Error(codes.Internal, err.Error())
	}

	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				return status.Error(codes.Internal, "subscription channel closed")
			}
			pb, ok := resp.Proto()
			if !ok {
				u.logger.With(slog.Any("error", err), slog.Any("req", req)).WarnContext(serv.Context(), "Got invalid object from user change subscription")
				continue
			}

			err = serv.Send(pb)
			if err != nil {
				u.logger.With(slog.Any("error", err)).WarnContext(serv.Context(), "Got unexpected error sending grpc message to subscription")
				continue
			}
		case <-serv.Context().Done():
			return nil
		}
	}
}
