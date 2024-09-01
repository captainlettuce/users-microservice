package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/captainlettuce/users-microservice/internal/fixtures_test"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"
)

// generateTestUser generates a user that is guaranteed to have all it's public fields set to a non-zero value
// the id is random for each call
func generateTestUser() types.User {
	return fixtures_test.NewUserWith(func(u *types.User) {
		u.Id = uuid.New()
	})
}

func runTestWithMongoConnection(t *testing.T, test func(mr *mongoRepository)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	repo, err := NewMongoRepository(ctx, uri, dbName, collection)
	require.NoError(t, err)

	mr := repo.(*mongoRepository)
	defer func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*1)
		defer timeoutCancel()
		//
		err2 := mr.collection.Database().Drop(timeoutCtx)
		if err2 != nil {
			slog.Debug(fmt.Sprintf("could not drop test database '%s': %s", dbName, err2.Error()))
		}

		err2 = mr.collection.Database().Client().Disconnect(timeoutCtx)
		if err2 != nil {
			slog.Debug(fmt.Sprintf("could not disconnect from mongodb '%s': %s", uri, err2.Error()))
		}
	}()

	test(mr)
}

func TestTestFixtureCreation(t *testing.T) {
	t.Run("all fields are set", func(t *testing.T) {
		usr := generateTestUser()
		b, err := json.Marshal(usr)
		require.NoError(t, err, "json marshal error")

		usrMap := map[string]any{}
		err = json.Unmarshal(b, &usrMap)
		require.NoError(t, err, "json unmarshal error")

		for k, v := range usrMap {
			require.NotZero(t, v, "%s is set to it's zero-value", k)
		}
	})
}

func TestMongoRepository_CRUD(t *testing.T) {

	runTestWithMongoConnection(t, func(mr *mongoRepository) {
		usr := generateTestUser()
		t.Run("Add", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			require.NoError(t, mr.Add(ctx, &usr), "user could not be added")
			require.ErrorIs(t, mr.Add(ctx, &usr), types.ErrDuplicateUserId, "duplicate user id allowed")
		})

		// Just validating that all fields unset read
		t.Run("Get by Id", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			u2, cnt, err := mr.List(ctx, types.UserFilter{Ids: []uuid.UUID{usr.Id}}, types.Paging{Limit: 1, Offset: 0})
			require.NoError(t, err, "got error fetching user from db")
			require.Equal(t, cnt, uint64(1), "more than 1 result returned")
			require.Equal(t, usr, u2[0])
		})

		t.Run("update partial", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			newFirstName := "not test"
			u, err := mr.UpdatePartial(
				ctx,
				types.UserFilter{Ids: []uuid.UUID{usr.Id}},
				types.UpdateUserFields{
					FirstName: &newFirstName, // will be updated
					LastName:  ref(""),       // will be unset
				},
			)

			if err != nil {
				t.Errorf("got unexpected error updating user: %v", err)
			}

			if u.FirstName != newFirstName {
				t.Errorf("first name was not set correctly %s", cmp.Diff(newFirstName, u.FirstName))
			}
			if u.LastName != "" {
				t.Errorf("last name was not cleared properly")
			}
		})

		t.Run("delete", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			require.NoError(t, mr.Delete(ctx, usr.Id), "got error trying to delete user")
			users, cnt, err := mr.List(ctx, types.UserFilter{Ids: []uuid.UUID{usr.Id}}, types.Paging{Limit: 1, Offset: 0})
			require.NoError(t, err, "got error fetching user from db")
			require.Equal(t, cnt, uint64(0))
			require.Len(t, users, 0)
		})
	})
}

func TestMongoRepository_List(t *testing.T) {
	runTestWithMongoConnection(t, func(mr *mongoRepository) {
		usr := generateTestUser()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		require.NoError(t, mr.Add(ctx, &usr))

		timeBeforeUserCreate := usr.CreatedAt.Add(-1 * time.Hour)
		timeAfterUserCreate := usr.CreatedAt.Add(1 * time.Hour)

		timeBeforeUserUpdate := usr.UpdatedAt.Add(-1 * time.Hour)
		timeAfterUserUpdate := usr.UpdatedAt.Add(1 * time.Hour)

		tests := []struct {
			name      string
			filter    types.UserFilter
			wantCount uint64
		}{
			{
				name:      "happy match id",
				filter:    types.UserFilter{Ids: []uuid.UUID{usr.Id}},
				wantCount: 1,
			},
			{
				name:      "sad match id",
				filter:    types.UserFilter{Ids: []uuid.UUID{uuid.Nil}},
				wantCount: 0,
			},
			{
				name:      "happy match first name",
				filter:    types.UserFilter{FirstName: usr.FirstName},
				wantCount: 1,
			},
			{
				name:   "sad match first name",
				filter: types.UserFilter{FirstName: "this-is-not-a-name"},
			},
			{
				name:      "happy match last name",
				filter:    types.UserFilter{LastName: usr.LastName},
				wantCount: 1,
			},
			{
				name:   "sad match last name",
				filter: types.UserFilter{LastName: "this-is-not-a-name"},
			},
			{
				name:      "happy match nickname",
				filter:    types.UserFilter{Nickname: usr.Nickname},
				wantCount: 1,
			},
			{
				name:   "sad match nickname",
				filter: types.UserFilter{Nickname: "this-is-not-a-name"},
			},
			{
				name:      "happy match email",
				filter:    types.UserFilter{Email: usr.Email},
				wantCount: 1,
			},
			{
				name:   "sad match email",
				filter: types.UserFilter{Email: "this-is-not-an-email"},
			},
			{
				name:      "happy match countries",
				filter:    types.UserFilter{Countries: []string{usr.Country}},
				wantCount: 1,
			},
			{
				name:   "sad match country",
				filter: types.UserFilter{Email: "this-is-not-a-country"},
			},
			{
				name:      "happy match created before",
				filter:    types.UserFilter{Created: &types.TimeFilter{Before: &timeAfterUserCreate}},
				wantCount: 1,
			},
			{
				name:   "sad match created before",
				filter: types.UserFilter{Created: &types.TimeFilter{Before: &timeBeforeUserCreate}},
			},
			{
				name:      "happy match created after",
				filter:    types.UserFilter{Created: &types.TimeFilter{After: &timeBeforeUserCreate}},
				wantCount: 1,
			},
			{
				name:   "sad match created after",
				filter: types.UserFilter{Created: &types.TimeFilter{After: &timeAfterUserCreate}},
			},
			{
				name:      "happy match updated before",
				filter:    types.UserFilter{Updated: &types.TimeFilter{Before: &timeAfterUserUpdate}},
				wantCount: 1,
			},
			{
				name:   "sad match updated before",
				filter: types.UserFilter{Updated: &types.TimeFilter{Before: &timeBeforeUserUpdate}},
			},
			{
				name:      "happy match updated after",
				filter:    types.UserFilter{Updated: &types.TimeFilter{After: &timeBeforeUserUpdate}},
				wantCount: 1,
			},
			{
				name:   "sad match updated after",
				filter: types.UserFilter{Updated: &types.TimeFilter{After: &timeAfterUserUpdate}},
			},
			{
				name: "happy match created before and after",
				filter: types.UserFilter{
					Created: &types.TimeFilter{
						Before: &timeAfterUserCreate,
						After:  &timeBeforeUserCreate,
					},
				},
				wantCount: 1,
			},
			{
				name:      "sad don't match mis-matched names",
				filter:    types.UserFilter{FirstName: usr.FirstName, LastName: "this-is-not-a-name"},
				wantCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel2()

				users, cnt, err := mr.List(ctx2, tt.filter, types.Paging{Limit: 10, Offset: 0})
				require.NoError(t, err, "got unexpected error from db")
				require.Equal(t, tt.wantCount, cnt, "unexpected result count")
				require.Equal(t, uint64(len(users)), tt.wantCount, "unexpected result slice count")
			})
		}
	})
}

func Test_userCriteriaToMongoFilter(t *testing.T) {
	now := time.Now()
	filter := types.UserFilter{
		Ids:       []uuid.UUID{uuid.New()},
		FirstName: "test",
		LastName:  "test",
		Nickname:  "test",
		Email:     "test@example.com",
		Countries: []string{"UK"},
		Created: &types.TimeFilter{
			Before: &now,
			After:  &now,
		},
		Updated: &types.TimeFilter{
			Before: &now,
			After:  &now,
		},
	}

	// This is quite a bad test, really, it needs to be updated as filterable fields are added...
	// the real test should probably happen in a testing-environment against an actual mongo-instance
	t.Run("all fields should be set", func(t *testing.T) {
		f := userFilterToMongoFilter(filter)

		reflectedFilter := reflect.ValueOf(filter)

		require.Equal(t, len(reflect.VisibleFields(reflectedFilter.Type())), len(f), "should have all fields set")
	})

	// Equally bad test
	t.Run("no unset fields should be set", func(t *testing.T) {
		f := userFilterToMongoFilter(types.UserFilter{})
		require.Len(t, f, 0, "unset input-fields have been set in output")
	})
}
