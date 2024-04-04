package service

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"github.com/sirupsen/logrus"
	"math"
	"time"
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

func (s *Service) CreateCredit(ctx context.Context, credit models.Credit) (createdCredit models.Credit, err error) {
	s.logger.Info("received create credit req")

	credit.MonthlyPayment, credit.DateOfIssue, credit.MaturityDate = CalculateCreditParams(credit.Term, credit.Amount, credit.AnnualInterestRate)

	createdCredit, err = s.storage.CreateCredit(ctx, credit)
	if err != nil {
		s.logger.Errorf("failed to create credit:%s", err)
		return models.Credit{}, err
	}

	s.logger.Info("credit created")

	return createdCredit, nil
}

func (s *Service) GetCredits(ctx context.Context) ([]models.Credit, error) {
	s.logger.Info("received get credits req")

	credits, err := s.storage.GetCredits(ctx)
	if err != nil {
		s.logger.Errorf("failed to get credits:%s", err)
		return []models.Credit{}, err
	}

	s.logger.Info("credits got")

	return credits, nil
}

func (s *Service) GetCreditById(ctx context.Context, id string) (models.Credit, error) {
	s.logger.Info("received get credit by id req")

	credit, err := s.storage.GetCreditById(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to get credit by id:%s", err)
		return models.Credit{}, err
	}

	s.logger.Info("credit by id got")

	return credit, err
}

func (s *Service) GetCreditsByUserId(ctx context.Context, userID int64) ([]models.Credit, error) {
	s.logger.Info("received get credit by userId req")

	credit, err := s.storage.GetCreditsByUserId(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get credit by userId:%s", err)
		return nil, err
	}

	s.logger.Info("credit by userId got")

	return credit, err
}

func (s *Service) UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error) {
	s.logger.Info("received update credit req")

	credit.MonthlyPayment, credit.DateOfIssue, credit.MaturityDate = CalculateCreditParams(credit.Term, credit.Amount, credit.AnnualInterestRate)

	updatedCredit, err = s.storage.UpdateCredit(ctx, credit)
	if err != nil {
		s.logger.Errorf("failed to update credit:%s", err)
		return models.Credit{}, err
	}

	s.logger.Info("credit updated")

	return updatedCredit, err
}

func (s *Service) DeleteCredit(ctx context.Context, id string) error {
	s.logger.Info("received delete credit req")

	if err := s.storage.DeleteCredit(ctx, id); err != nil {
		s.logger.Errorf("failed to delete credit:%s", err)
		return err
	}

	s.logger.Info("credit deleted")

	return nil
}

func CalculateCreditParams(term, amount int, annualInterestRate float64) (monthlyPayment int, dateOfIssue, maturityDate string) {
	monthlyInterestRate := annualInterestRate / 100 / 12 //месячная % ставка
	numerator := monthlyInterestRate * float64(amount)
	denominator := 1 - math.Pow(1+monthlyInterestRate, -float64(term))
	monthlyPayment = int(numerator / denominator) //формула аннуитетного платежа
	dateOfIssue = time.Now().Format("1 January 2024")
	maturityDate = time.Now().AddDate(0, term, 0).Format("1 January 2024")
	return monthlyPayment, dateOfIssue, maturityDate
}
