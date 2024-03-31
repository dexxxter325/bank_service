package suite

import (
	"bank/credit_service/internal/app"
	"bank/credit_service/internal/config"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

type Suite struct {
	Cfg         *config.Config
	t           *testing.T
	Client      *http.Client
	MongoClient *mongo.Client
}

func New(t *testing.T) (st *Suite, ctx context.Context, killMongoDBContainer, closeTestDbConnection, killKafkaContainer func(), port string, err error) {
	t.Helper()
	t.Parallel()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfigByPath("../config/config_tests.yml")
	if err != nil {
		logger.Fatalf("init config failed:%s", err)
		return nil, nil, nil, nil, nil, "", err
	}

	cfg.Rest.Port, err = findFreePort()
	if err != nil {
		logger.Fatalf("failed to find free port:%s", err)
		return nil, nil, nil, nil, nil, "", err
	}

	killMongoDBContainer, closeTestDbConnection, mongoClient, err := ConnToTestMongoDB(cfg, context.Background(), t)
	if err != nil {
		logger.Fatalf("failed to create test MongoDb:%v", err)
		return nil, nil, nil, nil, nil, "", err
	}

	killKafkaContainer = NewTestKafka(context.Background(), cfg, t)

	go func() {
		app.RunRest(cfg, logger)
	}()

	time.Sleep(time.Second * 1) //for stop

	client := &http.Client{}
	st = &Suite{
		Cfg:         cfg,
		t:           t,
		Client:      client,
		MongoClient: mongoClient,
	}

	return st, ctx, killMongoDBContainer, closeTestDbConnection, killKafkaContainer, cfg.Rest.Port, err
}

func ConnToTestMongoDB(cfg *config.Config, ctx context.Context, t *testing.T) (func(), func(), *mongo.Client, error) {
	mongoContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:6"),
		mongodb.WithUsername(cfg.MongoDb.Username),
		mongodb.WithPassword(cfg.MongoDb.Password),
	)
	if err != nil {
		t.Fatalf("failed to start container in testmongo: %s", err)
	}

	mongoPort, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get mapped port for MongoDB container: %v", err)
	}

	cfg.MongoDb.Port = strings.TrimSuffix(string(mongoPort), "/tcp")

	mongoURI := fmt.Sprintf("mongodb://%s:%s", cfg.MongoDb.Host, cfg.MongoDb.Port)

	credentials := options.Credential{
		Username: cfg.MongoDb.Username,
		Password: cfg.MongoDb.Password,
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI).SetAuth(credentials))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create TestMongoDB client: %v", err)
	}

	killMongoDBContainer := func() {
		if err = mongoContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	closeTestDbConnection := func() {
		if err = client.Disconnect(ctx); err != nil {
			t.Fatalf("failed to close test MongoDB:%s", err)
		}
	}
	return killMongoDBContainer, closeTestDbConnection, client, nil
}

func NewTestKafka(ctx context.Context, cfg *config.Config, t *testing.T) func() {
	/*var mu = &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()*/

	var kafkaContainer *kafka.KafkaContainer
	var err error

	//due to the heavy load under the tests, the wrong port may be taken
	for {
		kafkaContainer, err = kafka.RunContainer(ctx,
			kafka.WithClusterID("test"),
			testcontainers.WithImage("confluentinc/confluent-local:7.5.0"),
		)
		if err == nil {
			break
		}
		if err = kafkaContainer.Terminate(ctx); err != nil {
			t.Errorf("err in terminate testKafkaContainer:%s", err)
		}

		t.Errorf("failed to start container in test kafka: %s", err)
	}

	kafkaHost, err := kafkaContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get Kafka container IP: %v", err)
	}

	kafkaPort, err := kafkaContainer.MappedPort(ctx, "9093")
	if err != nil {
		t.Fatalf("failed to get Kafka container port: %v", err)
	}

	kafkaPortStr := string(kafkaPort)

	parts := strings.Split(kafkaPortStr, "/")

	logrus.Infof("PORT:%s", parts[0])

	cfg.Kafka.Brokers = fmt.Sprintf("%s:%s", kafkaHost, parts[0])

	killKafkaContainer := func() {
		if err = kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate kafka container: %s", err)
			return
		}
	}

	return killKafkaContainer
}

func findFreePort() (string, error) {
	// ":0" для указания на то, что нужен любой свободный порт
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}

	defer listener.Close()

	address := listener.Addr().String()

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}

	return port, nil
}
