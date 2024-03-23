package suite

import (
	"bank/credit_service/internal/app"
	"bank/credit_service/internal/config"
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

type Suite struct {
	Cfg    *config.Config
	t      *testing.T
	Client *http.Client
}

func New(t *testing.T) (st *Suite, ctx context.Context, killContainer func(), closeTestDbConnection func(), port string, err error) {
	t.Helper()
	t.Parallel()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfigByPath("../config/config_tests.yml")
	if err != nil {
		logger.Fatalf("init config failed:%s", err)
		return nil, nil, nil, nil, "", err
	}

	cfg.Rest.Port, err = findFreePort()
	if err != nil {
		logger.Fatalf("failed to find free port:%s", err)
		return nil, nil, nil, nil, "", err
	}

	killContainer, closeTestDbConnection, err = ConnToTestMongoDB(logger, cfg, ctx)
	if err != nil {
		logger.Fatalf("failed to create test MongoDb:%v", err)
		return nil, nil, nil, nil, "", err
	}

	go func() {
		app.RunRest(cfg, logger)
	}()
	time.Sleep(time.Second * 1) //for stop

	client := &http.Client{}
	st = &Suite{
		Cfg:    cfg,
		t:      t,
		Client: client,
	}
	return st, ctx, killContainer, closeTestDbConnection, cfg.Rest.Port, err
}

var mu = &sync.Mutex{}

func ConnToTestMongoDB(log *logrus.Logger, cfg *config.Config, ctx context.Context) (func(), func(), error) {
	var dbClient *mongo.Client
	mu.Lock()
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
		return nil, nil, err
	}
	mu.Unlock()
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
		return nil, nil, err
	}
	// pull mongodb docker image for version 5.0
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "5.0",
		Env: []string{
			// username and password for mongodb superuser
			fmt.Sprintf("MONGO_INITDB_ROOT_USERNAME=%s", cfg.MongoDb.Username),
			fmt.Sprintf("MONGO_INITDB_ROOT_PASSWORD=%s", cfg.MongoDb.Password),
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
		return nil, nil, err
	}

	cfg.MongoDb.Port = resource.GetPort("27017/tcp")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		dbClient, err = mongo.Connect(
			context.TODO(),
			options.Client().ApplyURI(
				fmt.Sprintf("mongodb://%s:%s@%s:%s", cfg.MongoDb.Username, cfg.MongoDb.Password, cfg.MongoDb.Host, resource.GetPort("27017/tcp")),
			),
		)
		if err != nil {
			return err
		}
		return dbClient.Ping(ctx, nil)
	})

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
		return nil, nil, err
	}
	killContainer := func() {
		// When you're done, kill and remove the container
		if err = pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
			return
		}
	}
	// disconnect mongodb client
	dbDisconnect := func() {
		if err = dbClient.Disconnect(context.TODO()); err != nil {
			log.Fatalf("failed to close test mongoDB connect:%s", err)
			return
		}
	}

	return killContainer, dbDisconnect, nil
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
