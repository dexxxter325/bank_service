package jwt

import (
	"bank/auth_service/internal/domain/models"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GenerateAccessToken(user models.User, accessTokenTTL time.Duration, secretKey string) (string, error) {
	accessToken := jwt.New(jwt.SigningMethodHS256)
	claims := accessToken.Claims.(jwt.MapClaims)

	claims["userId"] = user.ID
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(accessTokenTTL).Unix()

	accessTokenString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("convert access token to string failed:%s", err)
	}

	return accessTokenString, err
}

func ValidateAccessToken(accessToken, secretKey string) (bool, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) { //для использования полей токена||проверяет подпись
		return []byte(secretKey), nil
	})
	if err != nil {
		return false, fmt.Errorf("parse accessToken failed:%s", err)
	}

	if !token.Valid { // подпись токена верна,claims прошли проверку,не истек ли
		return false, errors.New("accessToken is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, errors.New("invalid accessToken claims")
	}

	_, ok = claims["userId"].(float64)
	if !ok {
		return false, errors.New("userID is missing or invalid in accessToken")
	}

	return true, nil
}
func GenerateRefreshToken(userId int64, refreshTokenTTl time.Duration, secretKey string) (string, error) {
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	claims := refreshToken.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(refreshTokenTTl).Unix()
	claims["userId"] = userId

	refreshTokenString, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("convert refresh token to string failed:%s", err)
	}

	return refreshTokenString, err
}

func ParseRefreshToken(refreshToken, secretKey string) (int64, error) { //to get user_id for newaccess
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) { //для использования полей токена||проверяет подпись
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse refreshTOken failed:%s", err)
	}

	if !token.Valid { // подпись токена верна,claims прошли проверку,не истек ли
		return 0, errors.New("refresh token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid refresh token claims")
	}

	userIdFloat, ok := claims["userId"].(float64) //float,тк извлекаем из json(весь код jwt),а там это стандартный тип
	if !ok {
		return 0, errors.New("userID is missing or invalid in refresh token")
	}

	return int64(userIdFloat), nil
}
