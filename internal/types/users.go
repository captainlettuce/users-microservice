package types

import (
	"github.com/captainlettuce/users-microservice/generated"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"time"
)

type User struct {
	Id        uuid.UUID `bson:"_id,omitempty"`
	FirstName string    `bson:"first_name,omitempty"`
	LastName  string    `bson:"last_name,omitempty"`
	Nickname  string    `bson:"nickname,omitempty"`
	Email     string    `bson:"email,omitempty"`
	Password  string    `bson:"password,omitempty"`
	Country   string    `bson:"country,omitempty"`

	CreatedAt time.Time  `bson:"created_at,omitempty"`
	UpdatedAt *time.Time `bson:"updated_at,omitempty"`
}

// LogValue is used to make sure we don't leak any PII in logs
func (u User) LogValue() slog.Value {
	return slog.StringValue(u.Id.String())
}

func (u User) Proto() *generated.User {
	var protoUpdatedAt *timestamppb.Timestamp

	if u.UpdatedAt != nil {
		protoUpdatedAt = timestamppb.New(*u.UpdatedAt)
	}

	return &generated.User{
		Id:        u.Id.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Password:  u.Password,
		Email:     u.Email,
		Country:   u.Country,

		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: protoUpdatedAt,
	}
}

func UserFromProto(u *generated.User) (User, error) {
	user := User{
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Nickname:  u.GetNickname(),
		Email:     u.GetEmail(),
		Password:  u.GetPassword(),
		Country:   u.GetCountry(),
		CreatedAt: u.CreatedAt.AsTime(),
	}

	if u.Id != "" {
		if id, err := uuid.Parse(u.Id); err != nil {
			return user, ErrInvalidUserId
		} else {
			user.Id = id
		}
	}

	if u.UpdatedAt != nil {
		ua := u.UpdatedAt.AsTime()
		user.UpdatedAt = &ua
	}

	return user, nil
}

type UpdateUserFields struct {
	FirstName *string `bson:"first_name,omitempty" field_mask:"first_name"`
	LastName  *string `bson:"last_name,omitempty" field_mask:"last_name"`
	Nickname  *string `bson:"nickname,omitempty" field_mask:"nickname"`
	Email     *string `bson:"email,omitempty" field_mask:"email"`
	Password  *string `bson:"password,omitempty" field_mask:"password"`
	Country   *string `bson:"country,omitempty" field_mask:"country"`
}
