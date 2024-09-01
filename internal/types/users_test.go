package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
	"time"
)

func TestUserConversion(t *testing.T) {
	now := time.Now()
	text := "test"
	pb := &generated.User{
		Id:        uuid.New().String(),
		FirstName: text,
		LastName:  text,
		Nickname:  text,
		Password:  text,
		Email:     text,
		Country:   text,
		CreatedAt: convertTimeToTimestamppb(&now),
		UpdatedAt: convertTimeToTimestamppb(&now),
	}

	t.Run("all fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		f, err := UserFromProto(pb)
		require.NoError(t, err)
		pb2 := f.Proto()

		require.True(t, cmp.Equal(pb, pb2, protocmp.Transform()), "fields are not set correctly")
	})

	t.Run("sad case function fails on invalid uuid", func(t *testing.T) {
		pbInner := &generated.User{Id: "invalid-uuid"}
		_, err := UserFromProto(pbInner)
		require.Error(t, err, "function should not accept invalid uuid")
	})

	t.Run("happy case empty id gives no error", func(t *testing.T) {
		pb = &generated.User{}
		u, err := UserFromProto(pb)
		require.NoError(t, err)
		require.Equal(t, u.Id, uuid.Nil)
	})
}
