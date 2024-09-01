package repository

import (
	"context"
	"errors"
	"github.com/captainlettuce/users-microservice/internal"
	"github.com/captainlettuce/users-microservice/internal/types"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(ctx context.Context, uri string, db string, collection string) (internal.UserRepository, error) {

	opts := []*options.ClientOptions{
		options.Client().ApplyURI(uri),
		options.Client().SetTimeout(5 * time.Second),
	}

	client, err := mongo.Connect(
		ctx,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err = client.Ping(timeoutCtx, nil)
	if err != nil {
		if err2 := client.Disconnect(ctx); err2 != nil {
			slog.With(slog.Any("error", err2)).Warn("failed to disconnect mongoDB collection")
		}
		return nil, err
	}

	c := client.Database(db).Collection(collection)

	mr := &mongoRepository{collection: c}
	indexCtx, indexCancel := context.WithTimeout(ctx, 5*time.Second)
	defer indexCancel()

	err = mr.setupIndices(indexCtx)

	return mr, err
}

func (mr *mongoRepository) Shutdown(ctx context.Context) error {
	return mr.collection.Database().Client().Disconnect(ctx)
}

func (mr *mongoRepository) Add(ctx context.Context, user *types.User) error {
	_, err := mr.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return types.ErrDuplicateUserId
	}
	return err
}

func (mr *mongoRepository) UpdatePartial(ctx context.Context, filter types.UserFilter, fields types.UpdateUserFields) (types.User, error) {
	var u types.User

	opts := []*options.FindOneAndUpdateOptions{options.FindOneAndUpdate().SetReturnDocument(options.After)}

	updateFields, err := createUpdateDocument(fields)
	if err != nil {
		return u, err
	}

	res := mr.collection.FindOneAndUpdate(ctx, userFilterToMongoFilter(filter), updateFields, opts...)
	if err := res.Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return u, errors.Join(types.ErrNotFound, err)
		}
		return u, errors.Join(types.ErrUnknownError, err)
	}

	return u, nil
}

func (mr *mongoRepository) Delete(ctx context.Context, userId uuid.UUID) error {
	_, err := mr.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: userId}})
	if err != nil {
		return errors.Join(types.ErrUnknownError, err)
	}

	return nil
}

func (mr *mongoRepository) List(ctx context.Context, filter types.UserFilter, paging types.Paging) ([]types.User, uint64, error) {
	type listReturn struct {
		Users []types.User `bson:"users"`
		Total uint64       `bson:"total"`
	}

	var ret = make([]listReturn, 0, 1)

	matchStage := bson.D{{Key: "$match", Value: userFilterToMongoFilter(filter)}}

	countStage := bson.D{{Key: "$count", Value: "total"}}

	facetStage := bson.D{
		{
			Key: "$facet",
			Value: bson.D{
				{Key: "total", Value: bson.A{countStage}},
				{Key: "users", Value: bson.A{skipStage(paging.Offset), limitStage(paging.Limit)}},
			},
		},
	}

	addFieldsStage := bson.D{
		{
			Key: "$addFields",
			Value: bson.D{
				{
					Key: "total",
					Value: bson.D{
						{
							Key:   "$arrayElemAt",
							Value: bson.A{"$total.total", 0},
						},
					},
				},
			},
		},
	}

	pipeline := mongo.Pipeline{
		matchStage,
		facetStage,
		addFieldsStage,
	}

	res, err := mr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, errors.Join(types.ErrUnknownError, err)
	}

	err = res.All(ctx, &ret)
	if err != nil {
		return nil, 0, errors.Join(types.ErrUnknownError, err)
	}

	if len(ret) != 1 {
		return nil, 0, errors.Join(types.ErrUnknownError, errors.New("invalid return length"))
	}

	return ret[0].Users, ret[0].Total, nil
}

// userFilterToMongoFilter takes a UserFilter and translates it into a mongodb filter document
func userFilterToMongoFilter(filter types.UserFilter) bson.M {
	var ret = make(bson.M)
	if len(filter.Ids) > 0 {
		if len(filter.Ids) == 1 {
			ret["_id"] = filter.Ids[0]
		} else {
			ret["_id"] = bson.E{Key: "$in", Value: filter.Ids}
		}
	}

	if len(filter.Countries) > 0 {
		if len(filter.Countries) == 1 {
			ret["country"] = filter.Countries[0]
		} else {
			ret["country"] = bson.D{{Key: "$in", Value: filter.Countries}}
		}
	}

	if filter.FirstName != "" {
		ret["first_name"] = filter.FirstName
	}

	if filter.LastName != "" {
		ret["last_name"] = filter.LastName
	}

	if filter.Nickname != "" {
		ret["nickname"] = filter.Nickname
	}

	if filter.Email != "" {
		ret["email"] = filter.Email
	}

	if !IsZero(filter.Created) {
		ret["created_at"] = timeCriteriaToMongo(*filter.Created)
	}

	if !IsZero(filter.Updated) {
		ret["updated_at"] = timeCriteriaToMongo(*filter.Updated)
	}

	return ret
}

// timeCriteriaToMongo takes a types.TimeFilter filter to create a mongodb chronological filter
// the function handles both open and closed variants
func timeCriteriaToMongo(filter types.TimeFilter) bson.M {
	var ret = make(bson.M)
	if !IsZero(filter.Before) {
		ret["$lt"] = filter.Before
	}
	if !IsZero(filter.After) {
		ret["$gt"] = filter.After
	}

	if len(ret) == 0 {
		return nil
	}

	return ret
}

// skipStage returns an aggregation-stage that skips n result entities
func skipStage(n int64) bson.D {
	return bson.D{{Key: "$skip", Value: n}}
}

// limitStage returns an aggregation-stage that limits the number of results returned
func limitStage(n int64) bson.D {
	return bson.D{{Key: "$limit", Value: n}}
}
