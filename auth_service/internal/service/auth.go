package service

import (
	"bank/auth_service/internal/config"
	"bank/auth_service/internal/domain/models"
	"bank/auth_service/pkg/jwt"
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Storage interface {
	SaveUser(ctx context.Context, username string, hashedPassword []byte) (int64, error)
	GetUserByUsername(ctx context.Context, username string) (user models.User, err error)
	GetUserById(ctx context.Context, userId int64) (user models.User, err error)
}

type Service struct {
	storage Storage
	logger  *logrus.Logger
	cfg     *config.Config
}

func NewService(storage Storage, logger *logrus.Logger, cfg *config.Config) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
		cfg:     cfg,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (int64, error) {
	s.logger.Infof("received registration req: username=%s, password=%s", username, password)

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //salt already inside.Cost-hash security lvl
	if err != nil {
		s.logger.Error("failed to generate hash password")
		return 0, err
	}

	userId, err := s.storage.SaveUser(ctx, username, hashPassword)
	if err != nil {
		s.logger.Errorf("register failed:%s", err)
		return 0, err
	}

	s.logger.Info("user registered")

	return userId, err
}

func (s *Service) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	s.logger.Infof("received login req: username=%s, password=%s", username, password)

	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Errorf("failed to get user by username:%s", err)
		return "", "", err
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil { //валидность введенного юзером пароля
		s.logger.Errorf("password wrong:%s", err)
		return "", "", err
	}

	secretKey := s.cfg.Auth.SecretKey

	accessTokenTTLStr := s.cfg.Auth.AccessTokenTTL
	accessTokenTTL, err := time.ParseDuration(accessTokenTTLStr)
	if err != nil {
		return "", "", err
	}
	accessToken, err = jwt.GenerateAccessToken(user, accessTokenTTL, secretKey)
	if err != nil {
		return "", "", err
	}

	refreshTokenTTLStr := s.cfg.Auth.RefreshTokenTTL
	refreshTokenTTL, err := time.ParseDuration(refreshTokenTTLStr)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = jwt.GenerateRefreshToken(user.ID, refreshTokenTTL, secretKey)
	if err != nil {
		return "", "", err
	}

	s.logger.Info("user logged in")

	return accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	s.logger.Infof("received refreshToken req: refreshToken=%s", refreshToken)

	secretKey := s.cfg.Auth.SecretKey

	userId, err := jwt.ParseRefreshToken(refreshToken, secretKey)
	if err != nil {
		return "", "", err
	}
	user, err := s.storage.GetUserById(ctx, userId)
	if err != nil {
		return "", "", err
	}

	accessTokenTTLStr := s.cfg.Auth.AccessTokenTTL
	accessTokenTTL, err := time.ParseDuration(accessTokenTTLStr)
	if err != nil {
		return "", "", err
	}
	newAccessToken, err = jwt.GenerateAccessToken(user, accessTokenTTL, secretKey)
	if err != nil {
		return "", "", err
	}

	refreshTokenTTLStr := s.cfg.Auth.RefreshTokenTTL
	refreshTokenTTL, err := time.ParseDuration(refreshTokenTTLStr)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err = jwt.GenerateRefreshToken(user.ID, refreshTokenTTL, secretKey)
	if err != nil {
		return "", "", err
	}

	s.logger.Info("tokens generated")

	return newAccessToken, newRefreshToken, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, accessToken string) (bool, error) {
	s.logger.Infof("received validateAccessToken req: accessToken=%s", accessToken)

	secretKey := s.cfg.Auth.SecretKey

	ok, err := jwt.ValidateAccessToken(accessToken, secretKey)
	if !ok && err != nil {
		return false, err
	}

	s.logger.Info("access token validated")

	return ok, nil
}
