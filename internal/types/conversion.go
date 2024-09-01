package types

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func convertUUIDsToStrings(uuids []uuid.UUID) []string {
	var strs []string
	for _, u := range uuids {
		strs = append(strs, u.String())
	}
	return strs
}

func convertStringsToUUIDs(strs []string) ([]uuid.UUID, error) {
	var uuids []uuid.UUID
	for _, s := range strs {
		u, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, u)
	}
	return uuids, nil
}

func convertTimeToTimestamppb(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func convertTimestamppbToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
