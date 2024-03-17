package service

import (
	"bank/credit_service/internal/domain/models"
)

type Storage interface {
	CreateCredit(credit *models.Credit) (int, error)
	GetCredits() (*models.Credit, error)
	GetCreditById(int) (*models.Credit, error)
	UpdateCredit(int) (*models.Credit, error)
	DeleteCredit(int) error
}
