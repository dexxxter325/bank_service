package service

import (
	"bank/auth_service/internal/domain/models"
	"context"
	"github.com/segmentio/kafka-go"
)

type Storage interface {
	Auth
	KafkaProducer
}

type Auth interface {
	SaveUser(ctx context.Context, username string, hashedPassword []byte) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (user models.User, err error)
	GetUserById(ctx context.Context, userId int64) (user models.User, err error)
}

type KafkaProducer interface {
	PushUserIDs(ctx context.Context, writer *kafka.Writer) (err error)
}
