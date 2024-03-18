package service

import (
	"bank/credit_service/internal/domain/models"
	"context"
)

type Storage interface {
	CreateCredit(ctx context.Context, credit models.Credit) (string, error)
	GetCredits(ctx context.Context) ([]models.Credit, error)
	GetCreditById(ctx context.Context, id string) (models.Credit, error)
	UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error)
	DeleteCredit(ctx context.Context, id string) error
}
