package tests

import (
	"bank/credit_service/internal/storage"
	"bank/credit_service/tests/suite"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"io"
	"net/http"
	"testing"
	"time"
)

type Request struct {
	ID                 string
	UserID             int64
	Amount             int
	Currency           string
	Term               int
	AnnualInterestRate float64
}

type CreditResponse struct {
	CreatedCredit Response `json:"Created Credit"`
}

type Response struct {
	ID     string `json:"ID"`
	UserID int64  `json:"UserID"`
}

func TestCredit_OK(t *testing.T) {
	st, ctx, killMongoDBContainer, closeTestDbConnection, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killMongoDBContainer()
		closeTestDbConnection()
		killKafkaContainer()
		logrus.Infof("Kafka container killed")
	}()

	userID := randomInt64()

	ConsumerMongoDB := storage.NewConsumerMongoDB(st.MongoClient.Database(st.Cfg.MongoDb.Dbname), st.Cfg.MongoDb.UserIDCollection)

	err = ConsumerMongoDB.NewUserIDCollection(ctx, userID)
	require.NoError(t, err)

	createReqData := Request{
		UserID:             userID,
		Amount:             randomInt(),
		Currency:           randomString(5),
		Term:               randomInt(),
		AnnualInterestRate: randomFloat64(),
	}
	jsonData, err := json.Marshal(createReqData)
	require.NoError(t, err)

	createReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/credits", restPort), bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	createResp, err := st.Client.Do(createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, createResp.StatusCode)

	body, err := io.ReadAll(createResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	var response CreditResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	require.Equal(t, createResp.StatusCode, http.StatusOK)

	defer createResp.Body.Close()

	getReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits", restPort), nil)
	require.NoError(t, err)

	getResp, err := st.Client.Do(getReq)
	require.NoError(t, err)

	body, err = io.ReadAll(getResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	require.Equal(t, getResp.StatusCode, http.StatusOK)

	defer getResp.Body.Close()

	getByIdReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits/objectID/%s", restPort, response.CreatedCredit.ID), nil)
	require.NoError(t, err)

	getByIdResp, err := st.Client.Do(getByIdReq)
	require.NoError(t, err)

	body, err = io.ReadAll(getByIdResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	require.Equal(t, getByIdResp.StatusCode, http.StatusOK)

	defer getByIdResp.Body.Close()

	getByUserIdReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits/userID/%v", restPort, response.CreatedCredit.UserID), nil)
	require.NoError(t, err)

	getByUserIdResp, err := st.Client.Do(getByUserIdReq)
	require.NoError(t, err)

	body, err = io.ReadAll(getByUserIdResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	require.Equal(t, getByUserIdResp.StatusCode, http.StatusOK)

	defer getByUserIdResp.Body.Close()

	updateReqData := Request{
		UserID:             userID,
		Amount:             randomInt(),
		Currency:           randomString(5),
		Term:               randomInt(),
		AnnualInterestRate: randomFloat64(),
	}

	jsonData, err = json.Marshal(updateReqData)
	require.NoError(t, err)

	updateReq, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:%s/credits/%s", restPort, response.CreatedCredit.ID), bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	updateResp, err := st.Client.Do(updateReq)
	require.NoError(t, err)

	body, err = io.ReadAll(updateResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	require.Equal(t, updateResp.StatusCode, http.StatusOK)

	defer updateResp.Body.Close()

	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:%s/credits/%s", restPort, response.CreatedCredit.ID), nil)
	require.NoError(t, err)

	deleteResp, err := st.Client.Do(deleteReq)
	require.NoError(t, err)

	body, err = io.ReadAll(deleteResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	require.Equal(t, deleteResp.StatusCode, http.StatusOK)

	defer deleteResp.Body.Close()
}

func TestCreate_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		userID             int64
		amount             int
		currency           string
		term               int
		annualInterestRate float64
		expectedErr        string
		expectedStatusCode int
	}{
		{
			name:               "empty fields",
			expectedErr:        "you must fill the 'UserID' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty userID",
			userID:             0,
			amount:             0,
			currency:           randomString(5),
			term:               randomInt(),
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'UserID' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty amount",
			userID:             randomInt64(),
			amount:             0,
			currency:           randomString(5),
			term:               randomInt(),
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Amount' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty currency",
			userID:             randomInt64(),
			amount:             randomInt(),
			currency:           "",
			term:               randomInt(),
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Currency' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty term",
			userID:             randomInt64(),
			amount:             randomInt(),
			currency:           randomString(5),
			term:               0,
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Term' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty annualInterestRate",
			userID:             randomInt64(),
			amount:             randomInt(),
			currency:           randomString(5),
			term:               randomInt(),
			annualInterestRate: 0,
			expectedErr:        "you must fill the 'AnnualInterestRate' value",
			expectedStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createReqData := Request{
				UserID:             tt.userID,
				Amount:             tt.amount,
				Currency:           tt.currency,
				Term:               tt.term,
				AnnualInterestRate: tt.annualInterestRate,
			}
			jsonData, err := json.Marshal(createReqData)
			require.NoError(t, err)

			createReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/credits", restPort), bytes.NewBuffer(jsonData))
			require.NoError(t, err)

			resp, err := st.Client.Do(createReq)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tt.expectedErr)
		})
	}
}

func TestGetAll_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		expectedErr        string
		expectedStatusCode int
	}{
		{
			name:               "no credits found",
			expectedStatusCode: http.StatusNotFound,
			expectedErr:        "no credits found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits", restPort), nil)
			require.NoError(t, err)

			getResp, err := st.Client.Do(getReq)
			require.NoError(t, err)
			defer getResp.Body.Close()

			body, err := io.ReadAll(getResp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tt.expectedErr)

			require.Equal(t, getResp.StatusCode, tt.expectedStatusCode)

		})
	}
}

func TestGetById_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "no credit with provided id",
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "no credit found with provided ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getByIdReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits/objectID/%v", restPort, randomHex()), nil)
			require.NoError(t, err)

			getByIdResp, err := st.Client.Do(getByIdReq)
			require.NoError(t, err)

			body, err := io.ReadAll(getByIdResp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tt.expectedError)

			require.Equal(t, getByIdResp.StatusCode, tt.expectedStatusCode)

			defer getByIdResp.Body.Close()
		})
	}
}

func TestGetByUserId_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "no credits found for provided userID",
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "no credits found for provided userID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getByIdReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s/credits/userID/%v", restPort, randomInt64()), nil)
			require.NoError(t, err)

			getByIdResp, err := st.Client.Do(getByIdReq)
			require.NoError(t, err)

			body, err := io.ReadAll(getByIdResp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tt.expectedError)

			require.Equal(t, getByIdResp.StatusCode, tt.expectedStatusCode)

			defer getByIdResp.Body.Close()
		})
	}
}

func TestUpdate_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		amount             int
		currency           string
		term               int
		annualInterestRate float64
		expectedErr        string
		expectedStatusCode int
	}{
		{
			name:               "empty fields",
			expectedErr:        "you must fill the 'Amount' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty amount",
			amount:             0,
			currency:           randomString(5),
			term:               randomInt(),
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Amount' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty currency",
			amount:             randomInt(),
			currency:           "",
			term:               randomInt(),
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Currency' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty term",
			amount:             randomInt(),
			currency:           randomString(5),
			term:               0,
			annualInterestRate: randomFloat64(),
			expectedErr:        "you must fill the 'Term' value",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty annualInterestRate",
			amount:             randomInt(),
			currency:           randomString(5),
			term:               randomInt(),
			annualInterestRate: 0,
			expectedErr:        "you must fill the 'AnnualInterestRate' value",
			expectedStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, closeReq := getIdForReq(st, t, restPort)
			defer closeReq()

			updateReqData := Request{
				Amount:             tt.amount,
				Currency:           tt.currency,
				Term:               tt.term,
				AnnualInterestRate: tt.annualInterestRate,
			}

			jsonData, err := json.Marshal(updateReqData)
			require.NoError(t, err)

			updateReq, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:%s/credits/%s", restPort, response.CreatedCredit.ID), bytes.NewBuffer(jsonData))
			require.NoError(t, err)

			updateResp, err := st.Client.Do(updateReq)
			require.NoError(t, err)

			body, err := io.ReadAll(updateResp.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			require.Equal(t, updateResp.StatusCode, tt.expectedStatusCode)
			require.Contains(t, string(body), tt.expectedErr)

			defer updateResp.Body.Close()
		})
	}
}

func TestDelete_Fail(t *testing.T) {
	st, _, killDB, closeDB, killKafkaContainer, restPort, err := suite.New(t)
	require.NoError(t, err)

	defer func() {
		killDB()
		closeDB()
		killKafkaContainer()
	}()

	tests := []struct {
		name               string
		expectedErr        string
		expectedStatusCode int
	}{
		{
			name:               "no credit with provided id",
			expectedStatusCode: http.StatusNotFound,
			expectedErr:        "no credit found with provided ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:%s/credits/%s", restPort, randomHex()), nil)
			require.NoError(t, err)

			deleteResp, err := st.Client.Do(deleteReq)
			require.NoError(t, err)

			body, err := io.ReadAll(deleteResp.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			require.Equal(t, deleteResp.StatusCode, tt.expectedStatusCode)
			require.Contains(t, string(body), tt.expectedErr)

			defer deleteResp.Body.Close()
		})
	}
}

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomInt() int {
	x := rand.Intn(9223372036854775807) + 1 //max int in MongoDB
	return x
}

func randomInt64() int64 {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Int63n(9223372036854775807) //max int in MongoDB
}

func randomFloat64() float64 {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Float64() * 100 // Генерируем случайное число с плавающей запятой в диапазоне от 0 до 100
}

func randomHex() string {
	const hexChars = "0123456789abcdef"
	const length = 24 // ObjectID в MongoDB имеет длину 24 символа
	// Создаем буфер для хранения шестнадцатеричного значения
	buffer := make([]byte, length)
	// Заполняем буфер случайными шестнадцатеричными символами
	for i := 0; i < length; i++ {
		buffer[i] = hexChars[rand.Intn(len(hexChars))]
	}
	// Преобразуем буфер в строку
	return string(buffer)
}

func getIdForReq(st *suite.Suite, t *testing.T, restPort string) (CreditResponse, func() error) {
	userID := randomInt64()

	ConsumerMongoDB := storage.NewConsumerMongoDB(st.MongoClient.Database(st.Cfg.MongoDb.Dbname), st.Cfg.MongoDb.UserIDCollection)

	err := ConsumerMongoDB.NewUserIDCollection(context.Background(), userID)
	require.NoError(t, err)

	createReqData := Request{
		UserID:             userID,
		Amount:             randomInt(),
		Currency:           randomString(5),
		Term:               randomInt(),
		AnnualInterestRate: randomFloat64(),
	}
	jsonData, err := json.Marshal(createReqData)
	require.NoError(t, err)

	createReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/credits", restPort), bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	createResp, err := st.Client.Do(createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, createResp.StatusCode)

	body, err := io.ReadAll(createResp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	var response CreditResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	require.Equal(t, createResp.StatusCode, http.StatusOK)

	return response, createResp.Body.Close
}
