package types

import (
	"github.com/captainlettuce/users-microservice/generated"
)

type Paging struct {
	Offset int64
	Limit  int64
}

func (p Paging) Proto() *generated.Paging {
	return &generated.Paging{
		Offset: p.Offset,
		Limit:  p.Limit,
	}
}

func PagingFromProto(pb *generated.Paging) Paging {
	p := Paging{}

	if pb != nil {
		p.Offset = pb.GetOffset()
		p.Limit = pb.GetLimit()
	}

	return p
}
