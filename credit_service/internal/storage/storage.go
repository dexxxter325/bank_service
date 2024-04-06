package storage

import "go.mongodb.org/mongo-driver/mongo"

type MongoDB struct {
	*AuthMongoDB
	*ConsumerMongoDB
}

func NewStorage(DB *mongo.Database, creditCollection, userIDCollection string) *MongoDB {
	return &MongoDB{
		AuthMongoDB:     NewAuthMongoDB(DB, creditCollection, userIDCollection),
		ConsumerMongoDB: NewConsumerMongoDB(DB, userIDCollection),
	}
}
