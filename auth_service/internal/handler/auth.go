package handler

import (
	"bank/auth_service/gen"
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	Register(ctx context.Context, username, password string) (int64, error)
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

type AuthServer struct { //=Handler
	gen.UnimplementedAuthServer
	service Service
	logger  *logrus.Logger
}

func NewAuthServer(service Service, logger *logrus.Logger) *AuthServer {
	return &AuthServer{
		UnimplementedAuthServer: gen.UnimplementedAuthServer{},
		service:                 service,
		logger:                  logger,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterResponse, error) {
	if req.GetUsername() == "" {
		s.logger.Error("username is required")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.GetPassword() == "" {
		s.logger.Error("password is required")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userId, err := s.service.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		s.logger.Errorf("method register failed:%s", err)
		return nil, status.Errorf(codes.Internal, "register failed:%s", err)
	}

	return &gen.RegisterResponse{
		UserId: userId,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *gen.LoginRequest) (*gen.LoginResponse, error) {
	if req.GetUsername() == "" {
		s.logger.Error("username is required")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.GetPassword() == "" {
		s.logger.Error("password is required")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	accessToken, refreshToken, err := s.service.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		s.logger.Errorf("method login failed:%s", err)
		return nil, status.Errorf(codes.Internal, "login failed:%s", err)
	}

	return &gen.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *gen.RefreshTokenRequest) (*gen.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		s.logger.Error("refresh token required")
		return nil, status.Error(codes.InvalidArgument, "refresh token required")
	}

	newAccessToken, newRefreshToken, err := s.service.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		s.logger.Errorf("method refresh token failed:%s", err)
		return nil, status.Errorf(codes.Internal, "method refresh token failed:%s", err)
	}

	return &gen.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
