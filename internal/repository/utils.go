package repository

import (
	"errors"
	"github.com/captainlettuce/users-microservice/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

// IsZero checks if a pointer is nil or the types zero-value
func IsZero[T comparable](v *T) bool {
	if v == nil {
		return true
	}
	var zero = new(T)
	return zero == v
}

// createUpdateDocument takes either a struct or a map and creates an update object for mongodb
// any fields set to a non-nil, zero-value, pointer will be unset
func createUpdateDocument(in any) (bson.D, error) {
	var (
		raw bson.Raw
		err error

		set   = bson.D{}
		unset = bson.D{}
	)
	raw, err = bson.Marshal(in)
	if err != nil {
		return nil, err
	}

	fields, err := raw.Elements()
	if err != nil {
		return nil, err
	}

	for _, e := range fields {
		var (
			k string
			v bson.RawValue
		)
		if e.Validate() != nil {
			return nil, errors.Join(types.ErrUnknownError, err)
		}

		if k, err = e.KeyErr(); err != nil {
			return nil, errors.Join(types.ErrUnknownError, err)
		}

		if v, err = e.ValueErr(); err != nil {
			return nil, errors.Join(types.ErrUnknownError, err)
		}

		if v.IsZero() {
			unset = append(unset, bson.E{Key: k, Value: v})
		} else {
			var vv any
			err = v.Unmarshal(&vv)
			if err != nil {
				return nil, errors.Join(types.ErrUnknownError, err)
			}

			rv := reflect.ValueOf(vv)
			switch rv.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				if rv.IsNil() || rv.Len() == 0 {
					unset = append(unset, bson.E{Key: k, Value: v})
				} else {
					set = append(set, bson.E{Key: k, Value: v})
				}
			default:
				if rv.IsZero() {
					unset = append(unset, bson.E{Key: k, Value: v})
				} else {
					set = append(set, bson.E{Key: k, Value: v})
				}
			}
		}
	}

	ret := bson.D{}
	if len(unset) > 0 {
		ret = append(ret, bson.E{Key: "$unset", Value: unset})
	}
	if len(set) > 0 {
		ret = append(ret, bson.E{Key: "$set", Value: set})
	}

	return ret, nil
}
