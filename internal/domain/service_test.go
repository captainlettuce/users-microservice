package domain

import (
	"context"
	"errors"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/fixtures_test"
	"github.com/captainlettuce/users-microservice/internal/mocks"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newTestService(mockRepo internal.UserRepository, mockPubSub internal.PubSubService) internal.UserService {
	return NewUserService(mockRepo, mockPubSub)
}

func Test_userService_Add(t *testing.T) {

	// CreatedAt & UpdatedAt fields of user creates problems when using cmp so init them to a static time
	staticTimestamp := time.Now().UTC()
	staticUUID := uuid.New()
	tests := []struct {
		name                         string
		user                         types.User
		wantErr                      bool
		repoError                    error
		discardMockExpectation       bool
		discardPubSubMockExpectation bool
		pubsubErr                    error
	}{
		{
			name: "happy case",
			user: types.User{
				Id:        staticUUID,
				CreatedAt: staticTimestamp,
				UpdatedAt: &staticTimestamp,
			},
		},
		{
			name:                         "sad case error from service",
			wantErr:                      true,
			repoError:                    errors.New("error"),
			discardPubSubMockExpectation: true,
		},
		{
			name:      "sad case pubsub error",
			wantErr:   false,
			pubsubErr: errors.New("error"),
			user: types.User{
				Id:        staticUUID,
				CreatedAt: staticTimestamp,
				UpdatedAt: &staticTimestamp,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			userCopy := tt.user
			mr := mocks.NewMockUserRepository(t)
			mps := mocks.NewMockPubSubService(t)

			if !tt.discardMockExpectation {
				mr.EXPECT().Add(ctx, &userCopy).Return(tt.repoError)
				if !tt.discardPubSubMockExpectation {
					mps.EXPECT().PublishUserChange(types.SubscriptionPayload{UserId: userCopy.Id, Change: types.UserChangeTypeCreated}).Return(tt.pubsubErr)
				}
			}

			s := newTestService(mr, mps)

			if err := s.Add(ctx, &userCopy); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				require.NotEqual(t, tt.user.CreatedAt, userCopy.CreatedAt, "CreatedAt was not updated")
				require.Nil(t, userCopy.UpdatedAt, "UpdatedAt was not set to nil")
			}
		})
	}
}

func Test_userService_Delete(t *testing.T) {

	tests := []struct {
		name                         string
		userId                       uuid.UUID
		wantErr                      bool
		discardMockExpectation       bool
		repoErr                      error
		discardPubSubMockExpectation bool
		pubsubErr                    error
	}{
		{
			name:   "happy case",
			userId: uuid.New(),
		},
		{
			name:                         "sad case error from service",
			userId:                       uuid.New(),
			wantErr:                      true,
			repoErr:                      errors.New("error"),
			discardPubSubMockExpectation: true,
		},
		{
			name:                         "sad case nil uuid",
			wantErr:                      true,
			discardMockExpectation:       true,
			discardPubSubMockExpectation: true,
		},
		{
			name:      "sad case error from pubsub",
			userId:    uuid.New(),
			wantErr:   false,
			pubsubErr: errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			mr := mocks.NewMockUserRepository(t)
			mps := mocks.NewMockPubSubService(t)

			if !tt.discardMockExpectation {
				mr.EXPECT().Delete(ctx, tt.userId).Return(tt.repoErr)
			}

			if !tt.discardPubSubMockExpectation {
				mps.EXPECT().PublishUserChange(types.SubscriptionPayload{UserId: tt.userId, Change: types.UserChangeTypeDeleted}).Return(tt.pubsubErr)
			}

			s := newTestService(mr, mps)

			if err := s.Delete(ctx, tt.userId); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userService_List(t *testing.T) {

	tests := []struct {
		name                   string
		filter                 types.UserFilter
		paging                 types.Paging
		wantErr                bool
		errFromMock            error
		discardMockExpectation bool
	}{
		{
			name: "happy case",
			paging: types.Paging{
				Limit: 5,
			},
		},
		{
			name:        "sad case error from repository",
			wantErr:     true,
			errFromMock: errors.New("error"),
			paging: types.Paging{
				Limit: 5,
			},
		},
		{
			name:                   "neutral case limit zero early return",
			discardMockExpectation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			m := mocks.NewMockUserRepository(t)

			if !tt.discardMockExpectation {
				m.EXPECT().List(ctx, tt.filter, tt.paging).Return([]types.User{}, 0, tt.errFromMock)
			}

			s := newTestService(m, nil)

			if _, _, err := s.List(ctx, tt.filter, tt.paging); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userService_UpdatePartial(t *testing.T) {
	var (
		ctx  = context.Background()
		user = fixtures_test.NewUser()
	)

	tests := []struct {
		name                         string
		wantErr                      bool
		discardRepoMockExpectation   bool
		repoError                    error
		discardPubSubMockExpectation bool
		pubsubError                  error
	}{
		{
			name: "happy path",
		},
		{
			name:                         "Sad path failed repo",
			repoError:                    errors.New("error"),
			discardPubSubMockExpectation: true,
			wantErr:                      true,
		},
		{
			name:        "Sad path (soft-)failed pubsub-publish",
			pubsubError: errors.New("error"),
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockUserRepository(t)
			pubsub := mocks.NewMockPubSubService(t)

			if !tt.discardRepoMockExpectation {
				repo.EXPECT().UpdatePartial(ctx, types.UserFilter{}, types.UpdateUserFields{}).Return(user, tt.repoError)
			}
			if !tt.discardPubSubMockExpectation {
				pubsub.EXPECT().PublishUserChange(types.SubscriptionPayload{UserId: user.Id, Change: types.UserChangeTypeUpdated}).Return(tt.pubsubError)
			}
			us := &userService{
				repo:   repo,
				pubsub: pubsub,
			}

			if err := us.UpdatePartial(ctx, types.UserFilter{}, types.UpdateUserFields{}); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePartial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
