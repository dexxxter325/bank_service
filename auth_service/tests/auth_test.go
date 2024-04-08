package tests

import (
	"bank/auth_service/gen"
	"bank/auth_service/pkg/jwt"
	"bank/auth_service/tests/suite"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestAuth_Ok(t *testing.T) {
	ctx, st := suite.New(t)

	username := gofakeit.Username()
	password := fakePass()
	secretKey := st.Cfg.Auth.SecretKey

	registerResp, err := st.AuthClient.Register(ctx, &gen.RegisterRequest{
		Username: username,
		Password: password,
	})
	require.NoError(t, err) //проверяем,что err=nil.если нет-тест дропнется

	id := registerResp.GetUserId()
	require.NotEmpty(t, id)

	loginResp, err := st.AuthClient.Login(ctx, &gen.LoginRequest{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	accessToken := loginResp.GetAccessToken()
	require.NotEmpty(t, accessToken)

	ok, err := jwt.ValidateAccessToken(accessToken, secretKey)
	require.True(t, ok)
	require.NoError(t, err)

	refreshToken := loginResp.GetRefreshToken()
	require.NotEmpty(t, refreshToken)

	refreshTokenResp, err := st.AuthClient.RefreshToken(ctx, &gen.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	require.NoError(t, err)

	newAccessToken := refreshTokenResp.GetAccessToken()
	require.NotEmpty(t, newAccessToken)

	newRefreshToken := refreshTokenResp.GetRefreshToken()
	require.NotEmpty(t, newRefreshToken)

	md := metadata.New(map[string]string{"Authorization": "Bearer " + accessToken})
	ctx = metadata.NewOutgoingContext(ctx, md)
	validateAccessTokenResp, err := st.AuthClient.ValidateAccessToken(ctx, &gen.ValidateAccessTokenRequest{})
	require.NoError(t, err)
	validAccessToken := validateAccessTokenResp.GetValid()
	require.NotEmpty(t, validAccessToken)
}

func TestRegister_Fail(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		password    string
		expectedErr string
	}{
		{
			name:        "empty fields",
			username:    "",
			password:    "",
			expectedErr: "you must fill the 'Username' value",
		},
		{
			name:        "empty username",
			username:    "",
			password:    fakePass(),
			expectedErr: "you must fill the 'Username' value",
		},
		{
			name:        "empty password",
			username:    gofakeit.Username(),
			password:    "",
			expectedErr: "you must fill the 'Password' value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &gen.RegisterRequest{
				Username: tt.username,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr) //есть ли ошибка в err.error в tt.expectedErr(выводятся также ненужные данные-rpc method...)
		})
	}
}

func TestLogin_Fail(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		username    string
		password    string
		expectedErr string
	}{
		{
			name:        "empty fields",
			username:    "",
			password:    "",
			expectedErr: "you must fill the 'Username' value",
		},
		{
			name:        "empty username",
			username:    "",
			password:    fakePass(),
			expectedErr: "you must fill the 'Username' value",
		},
		{
			name:        "empty password",
			username:    gofakeit.Username(),
			password:    "",
			expectedErr: "you must fill the 'Password' value",
		},
		{
			name:        "incorrect data",
			username:    gofakeit.Username(),
			password:    fakePass(),
			expectedErr: "user not found with username",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &gen.LoginRequest{
				Username: tt.username,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestRefreshToken_Fail(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name         string
		refreshToken string
		expectedErr  string
	}{
		{
			name:         "empty refreshToken",
			refreshToken: "",
			expectedErr:  "you must fill the 'RefreshToken' value",
		},
		{
			name:         "incorrect number of characters in RefreshToken",
			refreshToken: "qwerty123",
			expectedErr:  "token contains an invalid number of segments",
		},
		{
			name:         "invalid RefreshToken",
			refreshToken: "000000000000000",
			expectedErr:  "parse refreshTOken failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.RefreshToken(ctx, &gen.RefreshTokenRequest{
				RefreshToken: tt.refreshToken,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateAccessToken_Fail(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		accessToken string
		expectedErr string
	}{
		{
			name:        "empty accessToken",
			accessToken: "",
			expectedErr: "empty auth token",
		},
		{
			name:        "incorrect number of characters in access token",
			accessToken: "qwerty123",
			expectedErr: "token contains an invalid number of segments",
		},
		{
			name:        "invalid access token",
			accessToken: "000000000000000",
			expectedErr: "parse accessToken failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.New(map[string]string{"Authorization": "Bearer " + tt.accessToken})
			ctx = metadata.NewOutgoingContext(ctx, md)
			_, err := st.AuthClient.ValidateAccessToken(ctx, &gen.ValidateAccessTokenRequest{})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func fakePass() string {
	return gofakeit.Password(true, true, true, true, false, 15)
}

/*lower-символы нижнего регистра (a-z)
upper-символы верхнего регистра (A-Z).
numeric-цифры (0-9).
special-специальные символы.
space-пробелы
num-кол-во символов в пароле*/
