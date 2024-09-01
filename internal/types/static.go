package types

import (
	"errors"
)

const (
	GrpcServiceName = "users.v1"
	ServiceName     = "users"
)

var (
	ErrInvalidUserId   = errors.New("invalid userId")
	ErrDuplicateUserId = errors.New("duplicate userId")
	ErrNotFound        = errors.New("not found")
	ErrUnknownError    = errors.New("unknown error")
)

var ()
