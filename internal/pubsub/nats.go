package pubsub

import (
	"context"
	"errors"
	"fmt"
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

var usersTopic = "users"

type natsClient struct {
	logger *slog.Logger
	client *nats.Conn
}

func NewNatsClient(logger *slog.Logger, uri string) (internal.PubSubService, error) {
	nc, err := nats.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("could not connect to nats: %w", err)
	}

	return &natsClient{
			logger: logger.With(slog.String("component", "nats")),
			client: nc,
		},
		nil
}

func (nc *natsClient) SubscribeToUserChanges(ctx context.Context, req types.SubscriptionRequest) (<-chan types.SubscriptionPayload, error) {

	var ch = make(chan types.SubscriptionPayload)

	topic, err := natsTopicFromSubRequest(req)
	if err != nil {
		return nil, err
	}

	sub, err := nc.client.SubscribeSync(topic)
	if err != nil {
		return nil, errors.Join(types.ErrUnknownError, err)
	}

	go func() {
		defer close(ch)

		for {
			msg, err := sub.NextMsgWithContext(ctx)
			if err != nil {
				slog.With(slog.Any("error", err)).WarnContext(ctx, "Got unexpected error fetching message")
			}
			if ctx.Err() != nil {
				return
			}
			if msg == nil {
				continue
			}

			resp := &generated.SubscriptionResponse{}
			if err := proto.Unmarshal(msg.Data, resp); err != nil {
				slog.With(slog.Any("error", err)).WarnContext(ctx, "Got unexpected error unmarshalling message to protobuf")
				continue
			}

			res, err := types.SubscriptionPayloadFromProto(resp)
			if err != nil {
				slog.With(slog.Any("error", err)).WarnContext(ctx, "Got unexpected error converting from protobuf")
				continue
			}

			ch <- res
		}
	}()

	return ch, nil
}

func (nc *natsClient) PublishUserChange(result types.SubscriptionPayload) error {
	pb, ok := result.Proto()
	if !ok {
		return errors.Join(types.ErrUnknownError, errors.New("could not convert result to protobuf"))
	}

	b, err := proto.Marshal(pb)
	if err != nil {
		return errors.Join(types.ErrUnknownError, err)
	}

	err = nc.client.Publish(natsTopicFromSubResult(result), b)
	if err != nil {
		return errors.Join(types.ErrUnknownError, err)
	}
	return nil
}

func (nc *natsClient) GracefulShutdown(_ context.Context) error {
	nc.client.Close()
	return nil
}

func natsTopicFromSubRequest(req types.SubscriptionRequest) (string, error) {
	var (
		err         error
		changeField = "*"
		userIdField = "*"
	)
	if req.Change != nil {
		changeField = string(*req.Change)
	}
	if req.UserId != nil {
		if *req.UserId == uuid.Nil {
			return "", types.ErrInvalidUserId
		}
		userIdField = req.UserId.String()
	}

	return fmt.Sprintf("%s.%s.%s", usersTopic, changeField, userIdField), err
}

func natsTopicFromSubResult(resp types.SubscriptionPayload) string {
	return fmt.Sprintf("%s.%s.%s", usersTopic, resp.Change, resp.UserId.String())
}
