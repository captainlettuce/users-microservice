package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
)

func TestPagingFromProto(t *testing.T) {
	pb := &generated.Paging{
		Limit:  1,
		Offset: 1,
	}

	t.Run("all fields get tested", func(t *testing.T) {
		require.NoError(t, checkProtobufAllFieldsSet(pb))
	})

	t.Run("fields set to correct value", func(t *testing.T) {
		paging := PagingFromProto(pb)
		pb2 := paging.Proto()

		require.True(t, cmp.Equal(pb, pb2, protocmp.Transform()), "fields are not set correctly")
	})
}
