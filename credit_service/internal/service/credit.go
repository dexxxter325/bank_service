package service

import (
	"bank/credit_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *Service) CreateCredit(ctx context.Context, credit models.Credit) (string, error) {
	s.logger.Info("received create credit req")

	id, err := s.storage.CreateCredit(ctx, credit)
	if err != nil {
		s.logger.Errorf("faield to create credit:%s", err)
		return "", err
	}

	s.logger.Info("credit created")

	return id, nil
}

func (s *Service) GetCredits(ctx context.Context) ([]models.Credit, error) {
	s.logger.Info("received get credits credit req")

	credits, err := s.storage.GetCredits(ctx)
	if err != nil {
		s.logger.Errorf("failed to get credits:%s", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []models.Credit{}, fmt.Errorf("no credits found")
		}
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

func (s *Service) UpdateCredit(ctx context.Context, credit models.Credit) (updatedCredit models.Credit, err error) {
	s.logger.Info("received update credit req")

	updatedCredit, err = s.storage.UpdateCredit(ctx, credit)
	if err != nil {
		s.logger.Errorf("faield to update credit:%s", err)
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
