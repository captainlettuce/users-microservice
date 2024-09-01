package service

import (
	"context"
	"errors"
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/captainlettuce/users-microservice/generated/generated_mocks"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/mocks"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"os"
	"testing"
	"time"
)

func newTestService(mock internal.UserService) generated.UsersServiceServer {
	return NewUsersGrpc(mock, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func Test_usersGrpc_Add(t *testing.T) {

	// CreatedAt & UpdatedAt fields of user creates problems when using cmp so init them to a static time
	staticTimestamp := timestamppb.New(time.Now().UTC())

	tests := []struct {
		name                   string
		user                   *generated.User
		wantErr                bool
		errFromMock            error
		discardMockExpectation bool
	}{
		{
			name: "happy case",
			user: &generated.User{
				Id:        uuid.Nil.String(),
				CreatedAt: staticTimestamp,
			},
		},
		{
			name: "sad case bad uuid",
			user: &generated.User{
				Id: "invalid-uuid",
			},
			wantErr:                true,
			errFromMock:            nil,
			discardMockExpectation: true,
		},
		{
			name: "sad case error from service",
			user: &generated.User{
				Id: uuid.Nil.String(),
			},
			wantErr:     true,
			errFromMock: errors.New("mock error"),
		},
		{
			name: "sad case duplicate userId",
			user: &generated.User{
				Id: uuid.Nil.String(),
			},
			wantErr:     true,
			errFromMock: types.ErrInvalidUserId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			typesUser, _ := types.UserFromProto(tt.user)

			m := mocks.NewMockUserService(t)
			if !tt.discardMockExpectation {
				m.EXPECT().Add(ctx, &typesUser).Return(tt.errFromMock)
			}

			u := newTestService(m)

			_, err := u.Add(ctx, &generated.AddUserRequest{User: tt.user})
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				require.Truef(t, ok, "No status was found on returned error")
				require.NotEqual(t, codes.Unknown, st.Code(), "unknown status code set on returned error")
			}
		})
	}
}

func Test_usersGrpc_Delete(t *testing.T) {

	tests := []struct {
		name                   string
		req                    *generated.DeleteUserRequest
		wantErr                bool
		errFromMock            error
		discardMockExpectation bool
	}{
		{
			name: "happy case",
			req:  &generated.DeleteUserRequest{Id: uuid.Nil.String()},
		},
		{
			name:                   "sad case bad userId",
			req:                    &generated.DeleteUserRequest{Id: "invalid-uuid"},
			wantErr:                true,
			discardMockExpectation: true,
		},
		{
			name:        "sad case error from service",
			req:         &generated.DeleteUserRequest{Id: uuid.Nil.String()},
			wantErr:     true,
			errFromMock: errors.New("mock error"),
		},
		{
			name:        "sad case invalid userId",
			req:         &generated.DeleteUserRequest{Id: uuid.Nil.String()},
			wantErr:     true,
			errFromMock: types.ErrInvalidUserId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := mocks.NewMockUserService(t)
			if !tt.discardMockExpectation {
				userId, err := uuid.Parse(tt.req.Id)
				require.NoError(t, err, "Invalid uuid parsed when not discarding mock (broken test)")
				m.EXPECT().Delete(ctx, userId).Return(tt.errFromMock)
			}

			u := newTestService(m)

			_, err := u.Delete(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				require.Truef(t, ok, "No status was found on returned error")
				require.NotEqual(t, codes.Unknown, st.Code(), "unknown status code set on returned error")
			}
		})
	}
}

func Test_usersGrpc_Update(t *testing.T) {

	// CreatedAt & UpdatedAt fields of user creates problems when using cmp so init them to a static time
	staticTimestamp := timestamppb.New(time.Now().UTC())

	tests := []struct {
		name                   string
		req                    *generated.UpdateUserRequest
		user                   *generated.User
		wantErr                bool
		errFromMock            error
		discardMockExpectation bool
	}{
		{
			name: "happy case",
			req: &generated.UpdateUserRequest{
				User: &generated.User{
					Id:        uuid.Nil.String(),
					CreatedAt: staticTimestamp,
				},
			},
		},
		{
			name: "sad case error from service",
			req: &generated.UpdateUserRequest{
				User: &generated.User{
					Id: uuid.Nil.String(),
				},
			},
			wantErr:     true,
			errFromMock: errors.New("mock error"),
		},
		{
			name: "sad case invalid field_mask",
			req: &generated.UpdateUserRequest{
				User:       &generated.User{Id: uuid.Nil.String()},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"invalid"}},
			},
			wantErr:                true,
			discardMockExpectation: true,
		},
		{
			name: "sad case invalid filter",
			req: &generated.UpdateUserRequest{
				Filter: &generated.SearchFilter{Ids: []string{"invalid"}},
			},
			wantErr:                true,
			discardMockExpectation: true,
		},
		{
			name:        "sad case user not found",
			req:         &generated.UpdateUserRequest{},
			wantErr:     true,
			errFromMock: types.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := mocks.NewMockUserService(t)
			if !tt.discardMockExpectation {
				m.EXPECT().UpdatePartial(ctx, mock.Anything, mock.Anything).Return(tt.errFromMock)
			}

			u := newTestService(m)

			_, err := u.Update(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				require.Truef(t, ok, "No status was found on returned error")
				require.NotEqual(t, codes.Unknown, st.Code(), "unknown status code set on returned error")
				return
			}

		})
	}
}

func Test_usersGrpc_List(t *testing.T) {

	// used for *string values in struct literals
	testString := "test"

	tests := []struct {
		name                   string
		req                    *generated.ListUsersRequest
		filter                 types.UserFilter
		paging                 types.Paging
		wantErr                bool
		errFromMock            error
		discardMockExpectation bool
	}{
		{
			name: "happy case",
			req: &generated.ListUsersRequest{
				Paging:  &generated.Paging{Limit: 1, Offset: 0},
				Filters: &generated.SearchFilter{FirstName: &testString},
			},
			paging: types.Paging{Limit: 1, Offset: 0},
			filter: types.UserFilter{FirstName: testString},
		},
		{
			name: "sad case error from service",
			req: &generated.ListUsersRequest{
				Paging:  &generated.Paging{},
				Filters: &generated.SearchFilter{},
			},
			wantErr:     true,
			errFromMock: errors.New("mock error"),
		},
		{
			name: "sad case invalid filter",
			req: &generated.ListUsersRequest{Filters: &generated.SearchFilter{
				Ids: []string{"invalid"},
			}},
			wantErr:                true,
			discardMockExpectation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			m := mocks.NewMockUserService(t)

			if !tt.discardMockExpectation {
				m.EXPECT().List(ctx, tt.filter, tt.paging).Return([]types.User{{}}, 1, tt.errFromMock)
			}

			u := newTestService(m)

			_, err := u.List(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				require.True(t, ok, "No status was found on returned error")
				require.NotEqual(t, codes.Unknown, st.Code(), "unknown status code set on returned error")
			}

		})
	}
}

func Test_usersGrpc_Subscribe(t *testing.T) {
	invalidId := "invalid-uuid"
	tests := []struct {
		name                   string
		req                    *generated.SubscriptionRequest
		wantErr                bool
		discardMockExpectation bool
		mockError              error
	}{
		{
			name:    "happy case",
			req:     &generated.SubscriptionRequest{},
			wantErr: true, // we close the channel from the producer which should generate an error
		},
		{
			name:      "sad case error from service",
			req:       &generated.SubscriptionRequest{},
			wantErr:   true,
			mockError: errors.New("error"),
		},
		{
			name:                   "sad case invalid input",
			req:                    &generated.SubscriptionRequest{Params: &generated.SubscriptionParameters{UserId: &invalidId}},
			wantErr:                true,
			discardMockExpectation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ms := mocks.NewMockUserService(t)
			mgrpcServer := generated_mocks.NewMockUsersService_SubscribeServer(t)
			result := types.SubscriptionPayload{UserId: uuid.New(), Change: types.UserChangeTypeCreated}

			req, _ := types.SubscriptionRequestFromProto(tt.req)

			var ch chan types.SubscriptionPayload
			if !tt.discardMockExpectation {

				mgrpcServer.EXPECT().Context().Return(ctx)

				if tt.mockError == nil {

					// setup subscription mock
					ch = make(chan types.SubscriptionPayload)
					go func() {
						ch <- result
						close(ch)
					}()

					// setup server receiving mock
					pb, ok := result.Proto()
					require.True(t, ok)

					mgrpcServer.EXPECT().Send(pb).Return(nil)
				}

				ms.EXPECT().SubscribeToUserChanges(ctx, req).Return(ch, tt.mockError)
			}

			u := newTestService(ms)

			err := u.Subscribe(tt.req, mgrpcServer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				require.True(t, ok, "No status was found on returned error")
				require.NotEqual(t, codes.Unknown, st.Code(), "unknown status code set on returned error")
			}
		})
	}
}
