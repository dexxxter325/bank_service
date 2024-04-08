package handler

import (
	"bank/auth_service/gen"
	"context"
	"github.com/go-playground/validator/v10"
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

type RequestDTO struct {
	Username      string `validate:"required_if=OperationType authorization"`
	Password      string `validate:"required_if=OperationType authorization"`
	RefreshToken  string `validate:"required_if=OperationType refreshToken"`
	OperationType string
}

func (s *AuthServer) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.RegisterResponse, error) {

	dto := RequestDTO{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	dto.OperationType = "authorization"

	if err := s.ValidateValues(&dto); err != nil {
		return nil, err
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

	dto := RequestDTO{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	dto.OperationType = "authorization"

	if err := s.ValidateValues(&dto); err != nil {
		return nil, err
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

	dto := RequestDTO{
		RefreshToken: req.GetRefreshToken(),
	}

	dto.OperationType = "refreshToken"

	if err := s.ValidateValues(&dto); err != nil {
		return nil, err
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

func (s *AuthServer) ValidateValues(req *RequestDTO) error {
	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		validateErr := err.(validator.ValidationErrors)
		s.logger.Errorf("invalid req:%s", ValidationErrors(validateErr))
		return status.Errorf(codes.InvalidArgument, "invalid req:%s", ValidationErrors(validateErr))
	}

	return nil
}
