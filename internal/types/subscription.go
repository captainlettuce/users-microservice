package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/uuid"
	"slices"
)

type UserChangeType string

const (
	UserChangeTypeUnknown UserChangeType = "UNKNOWN"
	UserChangeTypeCreated UserChangeType = "CREATED"
	UserChangeTypeUpdated UserChangeType = "UPDATED"
	UserChangeTypeDeleted UserChangeType = "DELETED"
)

func UserChangeTypeFromString(s string) UserChangeType {
	if us := UserChangeType(s); slices.Contains([]UserChangeType{
		UserChangeTypeUnknown,
		UserChangeTypeCreated,
		UserChangeTypeUpdated,
		UserChangeTypeDeleted,
	}, UserChangeType(s)) {
		return us
	}

	return UserChangeTypeUnknown
}

type SubscriptionPayload struct {
	UserId uuid.UUID
	Change UserChangeType
}

func (sr SubscriptionPayload) Proto() (*generated.SubscriptionResponse, bool) {

	if sr.UserId == uuid.Nil {
		return nil, false
	}

	s, ok := generated.UserChangeType_value[string(sr.Change)]
	if !ok {
		return nil, false
	}

	resp := &generated.SubscriptionResponse{
		Update: &generated.SubscriptionMessage{
			UserId:     sr.UserId.String(),
			ChangeType: generated.UserChangeType(s),
		},
	}

	return resp, true
}

func SubscriptionPayloadFromProto(in *generated.SubscriptionResponse) (SubscriptionPayload, error) {
	if in == nil {
		return SubscriptionPayload{}, nil
	}

	id, err := uuid.Parse(in.GetUpdate().GetUserId())
	if err != nil {
		return SubscriptionPayload{}, err
	}

	return SubscriptionPayload{
		UserId: id,
		Change: UserChangeTypeFromString(in.GetUpdate().GetChangeType().String()),
	}, nil
}

type SubscriptionRequest struct {
	UserId *uuid.UUID
	Change *UserChangeType
}

func SubscriptionRequestFromProto(in *generated.SubscriptionRequest) (SubscriptionRequest, error) {
	sr := SubscriptionRequest{}
	if in == nil {
		return sr, nil
	}
	if userId := in.GetParams().GetUserId(); userId != "" {
		id, err := uuid.Parse(userId)
		if err != nil {
			return sr, err
		}

		sr.UserId = &id
	}

	if in.GetParams() != nil && in.GetParams().ChangeType != nil {
		ch := UserChangeTypeFromString(in.GetParams().GetChangeType().String())
		sr.Change = &ch
	}

	return sr, nil
}
