package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type UserFilter struct {
	Ids       []uuid.UUID
	FirstName string
	LastName  string
	Nickname  string
	Email     string
	Countries []string
	Created   *TimeFilter
	Updated   *TimeFilter
}

type TimeFilter struct {
	Before *time.Time
	After  *time.Time
}

func (tc *TimeFilter) Proto() *generated.TimeFilter {
	if tc == nil {
		return nil
	}
	return &generated.TimeFilter{
		Before: convertTimeToTimestamppb(tc.Before),
		After:  convertTimeToTimestamppb(tc.After),
	}
}

func TimeFilterFromProto(proto *generated.TimeFilter) *TimeFilter {
	if proto == nil {
		return nil
	}

	return &TimeFilter{
		Before: convertTimestamppbToTime(proto.GetBefore()),
		After:  convertTimestamppbToTime(proto.GetAfter()),
	}
}

func (uf *UserFilter) Proto() *generated.SearchFilter {
	return &generated.SearchFilter{
		Ids:       convertUUIDsToStrings(uf.Ids),
		FirstName: &uf.FirstName,
		LastName:  &uf.LastName,
		Nickname:  &uf.Nickname,
		Email:     &uf.Email,
		Countries: uf.Countries,
		Created:   uf.Created.Proto(),
		Updated:   uf.Updated.Proto(),
	}
}

func UserFilterFromProto(proto *generated.SearchFilter) (UserFilter, error) {
	if proto == nil {
		return UserFilter{}, nil
	}

	ids, err := convertStringsToUUIDs(proto.Ids)
	if err != nil {
		return UserFilter{}, err
	}

	return UserFilter{
		Ids:       ids,
		FirstName: proto.GetFirstName(),
		LastName:  proto.GetLastName(),
		Nickname:  proto.GetNickname(),
		Email:     proto.GetEmail(),
		Countries: proto.GetCountries(),
		Created:   TimeFilterFromProto(proto.Created),
		Updated:   TimeFilterFromProto(proto.Updated),
	}, nil
}

func (uf *UserFilter) LogValue() slog.Value {
	if uf == nil {
		return slog.Value{}
	}

	vals := []slog.Attr{
		slog.Any("ids", uf.Ids),
		slog.Any("nickname", uf.Nickname),
		slog.Any("created", uf.Created),
		slog.Any("updated", uf.Updated),
	}

	if uf.FirstName != "" {
		vals = append(vals, slog.Any("first_name", "REDACTED"))
	}

	if uf.LastName != "" {
		vals = append(vals, slog.Any("last_name", "REDACTED"))
	}
	if uf.Email != "" {
		vals = append(vals, slog.Any("email", "REDACTED"))
	}

	return slog.GroupValue(vals...)
}
