package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ref[T any](v T) *T {
	return &v
}

var mongoIndices = []mongo.IndexModel{
	{
		Keys: bson.D{{Key: "email", Value: 1}}, Options: &options.IndexOptions{
			Name: ref("users_email_single_asc"),
		},
	},

	// Add more indices here when there is need
}

func (mr *mongoRepository) setupIndices(ctx context.Context) error {
	for _, v := range mongoIndices {
		_, err := mr.collection.Indexes().CreateOne(ctx, v)
		if err != nil {
			return err
		}
	}
	return nil
}
