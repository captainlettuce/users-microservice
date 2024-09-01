package repository

import (
	"github.com/captainlettuce/users-microservice/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"testing"
)

func Test_createUpdateDocument(t *testing.T) {
	var (
		emptystring = ""

		firstname = "firstname"
		lastname  = "latsname"
		nickname  = "nickname"
		email     = "email@example.com"
		password  = "secret-hash"
		country   = "UK"
	)

	type args struct {
		in any
	}
	type output struct {
		set   bson.D
		unset bson.D
	}
	tests := []struct {
		name    string
		args    args
		want    output
		wantErr bool
	}{
		{
			name: "update all fields",
			args: args{
				in: types.UpdateUserFields{
					FirstName: &firstname,
					LastName:  &lastname,
					Nickname:  &nickname,
					Email:     &email,
					Password:  &password,
					Country:   &country,
				},
			},
			want: output{
				set: bson.D{
					{Key: "first_name", Value: firstname},
					{Key: "last_name", Value: lastname},
					{Key: "nickname", Value: nickname},
					{Key: "email", Value: email},
					{Key: "password", Value: password},
					{Key: "country", Value: country},
				},
			},
		},
		{
			name: "unset all fields",
			args: args{
				in: types.UpdateUserFields{
					FirstName: &emptystring,
					LastName:  &emptystring,
					Nickname:  &emptystring,
					Email:     &emptystring,
					Password:  &emptystring,
					Country:   &emptystring,
				},
			},
			want: output{
				unset: bson.D{
					{Key: "first_name", Value: ""},
					{Key: "last_name", Value: ""},
					{Key: "nickname", Value: ""},
					{Key: "email", Value: ""},
					{Key: "password", Value: ""},
					{Key: "country", Value: ""},
				},
			},
		},
		{
			name: "test non string types",
			args: args{
				in: bson.D{
					{"setInt", 1},
					{"unsetInt", 0},
					{"setSlice", []string{"test1", "test2"}},
					{"unsetSlice", []string{}},
					{"setMap", map[string]int{"test1": 1}},
					{"unsetMap", map[string]int{}},
				},
			},
			want: output{
				unset: bson.D{
					{Key: "unsetInt", Value: 0},
					{Key: "unsetSlice", Value: []string{}},
					{Key: "unsetMap", Value: map[string]int{}},
				},
				set: bson.D{
					{Key: "setInt", Value: 1},
					{Key: "setSlice", Value: []string{"test1", "test2"}},
					{Key: "setMap", Value: map[string]int{"test1": 1}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createUpdateDocument(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("createUpdateDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Since we are lazy and just sending marshalled values we need to recreate the objects in partly-serialized form
			// Order of execution matters here since bson.D is ordered
			var want bson.D
			if tt.want.unset != nil {
				unset := make(bson.D, len(tt.want.unset))
				for i, v := range tt.want.unset {
					bt, b, err := bson.MarshalValue(v.Value)
					if err != nil {
						t.Errorf("Unexpected error marshalling bson: %v", err)
					}
					unset[i] = bson.E{Key: v.Key, Value: bson.RawValue{Type: bt, Value: b}}
				}
				want = append(want, bson.E{Key: "$unset", Value: unset})
			}

			if tt.want.set != nil {
				set := make(bson.D, len(tt.want.set))
				for i, v := range tt.want.set {
					bt, b, err := bson.MarshalValue(v.Value)
					if err != nil {
						t.Errorf("Unexpected error marshalling bson: %v", err)
					}
					set[i] = bson.E{Key: v.Key, Value: bson.RawValue{Type: bt, Value: b}}
				}
				want = append(want, bson.E{Key: "$set", Value: set})
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("createUpdateDocument() got = %v\n want %v", got, want)
			}
		})
	}
}
