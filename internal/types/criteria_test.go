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

func TestTimeFilterConversion(t *testing.T) {
	now := time.Now()
	pb := &generated.TimeFilter{
		After:  convertTimeToTimestamppb(&now),
		Before: convertTimeToTimestamppb(&now),
	}

	t.Run("all fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		f := TimeFilterFromProto(pb)
		pb2 := f.Proto()

		require.True(t, cmp.Equal(pb, pb2, protocmp.Transform()), "fields are not set correctly")
	})
}

func TestUserFilterConversion(t *testing.T) {
	now := time.Now()
	testString := "test"
	pb := &generated.SearchFilter{
		Ids:       []string{uuid.New().String()},
		FirstName: &testString,
		LastName:  &testString,
		Nickname:  &testString,
		Email:     &testString,
		Countries: []string{"UK"},
		Created: &generated.TimeFilter{
			After:  convertTimeToTimestamppb(&now),
			Before: convertTimeToTimestamppb(&now),
		},
		Updated: &generated.TimeFilter{
			After:  convertTimeToTimestamppb(&now),
			Before: convertTimeToTimestamppb(&now),
		},
	}

	t.Run("fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		f, err := UserFilterFromProto(pb)
		require.NoError(t, err)
		pb2 := f.Proto()

		require.True(t, cmp.Equal(pb, pb2, protocmp.Transform()), "fields are not set correctly")
	})

	t.Run("sad case function fails on invalid uuid", func(t *testing.T) {
		pbInner := &generated.SearchFilter{Ids: []string{"invalid-uuid"}}
		_, err := UserFilterFromProto(pbInner)
		require.Error(t, err, "should not accept invalid uuids")
	})
}
