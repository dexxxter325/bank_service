package service

import (
	"bank/credit_service/internal/domain/models"
	"context"
)

type Storage interface {
	Auth
	KafkaConsumer
}

type Auth interface {
	CreateCredit(ctx context.Context, credit models.Credit) (models.Credit, error)
	GetCredits(ctx context.Context) ([]models.Credit, error)
	GetCreditById(ctx context.Context, id string) (models.Credit, error)
	GetCreditsByUserId(ctx context.Context, userID int64) ([]models.Credit, error)
	UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error)
	DeleteCredit(ctx context.Context, id string) error
}

type KafkaConsumer interface {
	NewUserIDCollection(ctx context.Context, userID int64) error
}
