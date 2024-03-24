package handler

import (
	"bank/auth_service/internal/config"
	"bank/auth_service/pkg/jwt"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

func UnaryInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		requestCtx, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "failed to get ctx in interceptor")
		}

		header := requestCtx.Get("Authorization")

		if len(header) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header in empty")
		}

		bearerAndToken := header[0]
		headerParts := strings.Split(bearerAndToken, " ") //делим на 2 части: до пробела и после

		accessToken := headerParts[1]

		if len(headerParts) != 2 && headerParts[0] != "Bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth header")
		}

		if len(accessToken) == 0 {
			return nil, status.Error(codes.Unauthenticated, "empty auth token")
		}

		ok, err := jwt.ValidateAccessToken(accessToken, cfg.Auth.SecretKey)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "failed to validate access token:%s", err)
		}

		return handler(ctx, req)
	}
}
