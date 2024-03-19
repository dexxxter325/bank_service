package mongodb

import (
	"bank/credit_service/internal/config"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnToMongoDB(cfg *config.Config) (*mongo.Database, error) {

	uri := fmt.Sprintf("mongodb://%s:%s", cfg.MongoDb.Host, cfg.MongoDb.Port)

	credentials := options.Credential{
		Username: cfg.MongoDb.Username,
		Password: cfg.MongoDb.Password,
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetAuth(credentials))
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect failed:%s", err)
	}
	return client.Database(cfg.MongoDb.Dbname), nil
}
