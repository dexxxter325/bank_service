package consumer

import (
	"bank/credit_service/internal/config"
	"bank/credit_service/internal/domain/models"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

func KafkaConsumer(ctx context.Context, cfg *config.Config, logger *logrus.Logger, db *mongo.Client, stop <-chan os.Signal) error {

	var credit models.Credit

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.Brokers},
		Topic:   cfg.Kafka.Topic,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			logger.Errorf("failed to close kafka reader:%s", err)
			return
		}
		logger.Infof("kafka reader closed")
		if err := db.Disconnect(ctx); err != nil {
			logger.Errorf("failed to close MongoDB in kafka:%s", err)
			return
		}
		logger.Info("mongoDB connection in kafka closed")
	}()
	//if ex reader didn't closed
	for {
		err := tryConnectToKafka(ctx, reader)
		if err == nil {
			break
		}
		logger.Errorf("the last reader has not finished his work yet.Shutting him down...: %s", err)
		time.Sleep(10 * time.Second)
	}

	// Бесконечный цикл для чтения сообщений из Kafka
	for {
		select {
		case <-time.After(5 * time.Second):
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Errorf("Error reading message from Kafka: %v", err)
				return err
			}
			// Проверка длины среза байт перед преобразованием в int64
			if len(msg.Value) < 8 {
				logger.Errorf("Message value is too short to convert to int64: %v", msg.Value)
				return errors.New("message value is too short to convert to int64")
			}
			// Получение userID из сообщения
			credit.UserID = int64(binary.BigEndian.Uint64(msg.Value))

			if err = NewUserIDCollection(ctx, cfg, db, credit.UserID); err != nil {
				if err.Error() == fmt.Sprintf("userID:%v already inserted into MongoDB", credit.UserID) {
					continue //в начало цикла
				}
				logger.Errorf("Error inserting userID into MongoDB: %v", err)
				return err
			}

			logger.Infof("UserID %v successfully inserted into MongoDB", credit.UserID)

		case <-stop:
			logger.Info("MongoDB and reader in kafka shutting down...")
			return nil
		}
	}
}

func tryConnectToKafka(ctx context.Context, reader *kafka.Reader) error {
	_, err := reader.ReadMessage(ctx)
	return err
}

func NewUserIDCollection(ctx context.Context, cfg *config.Config, db *mongo.Client, userID int64) error {
	collection := db.Database(cfg.MongoDb.Dbname).Collection(cfg.MongoDb.UserIDCollection)

	if IsUserIdNOTExist(ctx, userID, collection) {
		query := bson.M{"userID": userID}

		_, err := collection.InsertOne(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to insert userID:%s", err)
		}

		return nil
	}

	return fmt.Errorf("userID:%v already inserted into MongoDB", userID)
}

func IsUserIdNOTExist(ctx context.Context, userID int64, collection *mongo.Collection) bool {
	query := bson.M{"userID": userID}

	res := collection.FindOne(ctx, query)

	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return true
	}

	if res.Err() != nil {
		return true
	}

	return false
}
