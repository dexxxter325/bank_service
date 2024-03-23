package service

import (
	"context"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	Register(ctx context.Context, username, password string) (int64, error)
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

// TODO:import jwt auth from pkg
type Service struct {
	storage Storage
	logger  *logrus.Logger
}

func NewService(storage Storage, logger *logrus.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (int64, error) {
	panic("")
}

func (s *Service) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	panic("")
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	panic("")
}
