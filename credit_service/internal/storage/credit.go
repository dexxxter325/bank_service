package storage

import (
	"bank/credit_service/internal/domain/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage struct {
	DB *mongo.Client
}

func NewStorage(DB *mongo.Client) *Storage {
	return &Storage{DB: DB}
}

func (s *Storage) CreateCredit(credit *models.Credit) (int, error) {

}

func (s *Storage) GetCredits() (*models.Credit, error) {

}

func (s *Storage) GetCreditById(int) (*models.Credit, error) {

}

func (s *Storage) UpdateCredit(int) (*models.Credit, error) {

}

func (s *Storage) DeleteCredit(int) error {

}
