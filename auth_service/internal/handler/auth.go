package handler

import (
	"bank/auth_service/gen"
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type Service interface {
	Register(ctx context.Context, username, password string) (int64, error)
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	ValidateAccessToken(ctx context.Context, accessToken string) (bool, error)
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

func (s *AuthServer) ValidateAccessToken(ctx context.Context, req *gen.ValidateAccessTokenRequest) (*gen.ValidateAccessTokenResponse, error) {
	requestCtx, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Error("failed to get ctx in interceptor")
		return nil, status.Error(codes.Unauthenticated, "failed to get ctx in interceptor")
	}

	header := requestCtx.Get("Authorization")

	if len(header) == 0 {
		s.logger.Error("authorization header in empty")
		return nil, status.Error(codes.Unauthenticated, "authorization header in empty")
	}

	bearerAndToken := header[0]
	headerParts := strings.Split(bearerAndToken, " ") //делим на 2 части: до пробела и после

	accessToken := headerParts[1]

	if len(headerParts) != 2 && headerParts[0] != "Bearer" {
		s.logger.Error("invalid auth header")
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth header")
	}

	if len(accessToken) == 0 {
		s.logger.Error("empty auth token")
		return nil, status.Error(codes.Unauthenticated, "empty auth token")
	}

	ok, err := s.service.ValidateAccessToken(ctx, accessToken)
	if !ok && err != nil {
		s.logger.Errorf("failed to validate access token:%s", err)
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate access token:%s", err)
	}

	return &gen.ValidateAccessTokenResponse{
		Valid: ok,
	}, nil
}
