package fixtures_test

import (
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"time"
)

func NewUser() types.User {
	t, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	return types.User{
		Id:        uuid.MustParse("b9e52eb0-5bb1-4d34-a8b8-802f8f5b7a36"),
		FirstName: "firstname",
		LastName:  "lastname",
		Nickname:  "nickname",
		Email:     "email@email.com",
		Password:  "superSecretPassword",
		Country:   "UK",
		CreatedAt: t,
		UpdatedAt: ref(t.Add(time.Hour)),
	}
}

func NewUserWith(fn func(*types.User)) types.User {
	u := NewUser()
	fn(&u)
	return u
}

func ref[T any](v T) *T {
	return &v
}
