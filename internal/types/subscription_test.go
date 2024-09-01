package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
)

func TestSubscriptionRequestConversion(t *testing.T) {
	staticId := uuid.New()
	staticIdStr := staticId.String()
	change := UserChangeTypeCreated
	changePb, ok := generated.UserChangeType_value[string(change)]
	require.True(t, ok)

	ct := generated.UserChangeType(changePb)

	pb := &generated.SubscriptionRequest{
		Params: &generated.SubscriptionParameters{
			UserId:     &staticIdStr,
			ChangeType: &ct,
		},
	}

	og := SubscriptionRequest{
		UserId: &staticId,
		Change: &change,
	}

	t.Run("all fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		f, err := SubscriptionRequestFromProto(pb)
		require.NoError(t, err)

		require.True(t, cmp.Equal(f, og, protocmp.Transform()), "fields are not set correctly")
	})

	t.Run("sad case function fails on invalid uuid", func(t *testing.T) {
		invalidId := "invalid-uuid"
		pbInner := &generated.SubscriptionRequest{Params: &generated.SubscriptionParameters{UserId: &invalidId}}
		_, err := SubscriptionRequestFromProto(pbInner)
		require.Error(t, err, "function should not accept invalid uuid")
	})

	t.Run("empty Id is treated as valid", func(t *testing.T) {
		emptyId := ""
		_, err := SubscriptionRequestFromProto(&generated.SubscriptionRequest{Params: &generated.SubscriptionParameters{UserId: &emptyId}})
		require.NoError(t, err)
	})
}

func TestSubscriptionPayloadConversion(t *testing.T) {
	change := UserChangeTypeCreated
	changePb, ok := generated.UserChangeType_value[string(change)]
	require.True(t, ok)

	pb := &generated.SubscriptionResponse{
		Update: &generated.SubscriptionMessage{
			UserId:     uuid.New().String(),
			ChangeType: generated.UserChangeType(changePb),
		},
	}

	t.Run("all fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		f, err := SubscriptionPayloadFromProto(pb)
		require.NoError(t, err)

		pb2, ok := f.Proto()
		require.True(t, ok)

		require.True(t, cmp.Equal(pb, pb2, protocmp.Transform()), "fields are not set correctly")
	})

	t.Run("sad case function fails on invalid uuid", func(t *testing.T) {
		pbInner := &generated.SubscriptionResponse{Update: &generated.SubscriptionMessage{UserId: "invalid-uuid"}}
		_, err := SubscriptionPayloadFromProto(pbInner)
		require.Error(t, err, "function should not accept invalid uuid")
	})

	t.Run("sad case empty id is invalid", func(t *testing.T) {
		_, err := SubscriptionPayloadFromProto(&generated.SubscriptionResponse{Update: &generated.SubscriptionMessage{UserId: ""}})
		require.Error(t, err)
	})
}
