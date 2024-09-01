package pubsub

import (
	"fmt"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"testing"
)

func Test_natsTopicFromSubRequest(t *testing.T) {
	staticId := uuid.New()
	nilUuid := uuid.Nil
	changeType := types.UserChangeTypeCreated
	tests := []struct {
		name    string
		req     types.SubscriptionRequest
		want    string
		wantErr bool
	}{
		{
			name: "happy case all wildcard",
			req:  types.SubscriptionRequest{},
			want: usersTopic + ".*.*",
		},
		{
			name: "happy case userId specced",
			req:  types.SubscriptionRequest{UserId: &staticId},
			want: fmt.Sprintf("%s.*.%s", usersTopic, staticId.String()),
		},
		{
			name: "happy case changeType specced",
			req:  types.SubscriptionRequest{Change: &changeType},
			want: fmt.Sprintf("%s.%s.*", usersTopic, types.UserChangeTypeCreated),
		},
		{
			name:    "sad case nil userId",
			req:     types.SubscriptionRequest{UserId: &nilUuid},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := natsTopicFromSubRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("natsTopicFromSubRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("natsTopicFromSubRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
