package storage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConsumerMongoDB struct {
	userIDCollection *mongo.Collection
}

func NewConsumerMongoDB(DB *mongo.Database, userIDCollection string) *ConsumerMongoDB {
	return &ConsumerMongoDB{
		userIDCollection: DB.Collection(userIDCollection),
	}
}

func (d *ConsumerMongoDB) NewUserIDCollection(ctx context.Context, userID int64) error {

	if IsUserIdNOTExist(ctx, userID, d.userIDCollection) {
		query := bson.M{"userID": userID}

		_, err := d.userIDCollection.InsertOne(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to insert userID:%s", err)
		}

		return nil
	}

	return fmt.Errorf("userID:%v already inserted into MongoDB", userID)
}

func IsUserIdNOTExist(ctx context.Context, userID int64, userIDCollection *mongo.Collection) bool {
	query := bson.M{"userID": userID}
	res := userIDCollection.FindOne(ctx, query)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return true
	}

	if res.Err() != nil {
		return true
	}

	return false
}
