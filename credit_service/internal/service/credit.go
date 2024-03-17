package service

import (
	"bank/credit_service/internal/domain/models"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger  *logrus.Logger
	storage Storage
}

func NewService(logger *logrus.Logger, storage Storage) *Service {
	return &Service{
		logger:  logger,
		storage: storage,
	}
}

func (s *Service) CreateCredit(credit *models.Credit) (int, error) {

}

func (s *Service) GetCredits() (*models.Credit, error) {

}

func (s *Service) GetCreditById(int) (*models.Credit, error) {

}

func (s *Service) UpdateCredit(int) (*models.Credit, error) {

}

func (s *Service) DeleteCredit(int) error {

}
