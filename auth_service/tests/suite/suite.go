package suite

import (
	"bank/auth_service/gen"
	"bank/auth_service/internal/app"
	"bank/auth_service/internal/config"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // used by migrator
	_ "github.com/golang-migrate/migrate/v4/source/file"       // used by migrator
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strings"
	"testing"
	"time"
)

const (
	migrationPath = "../database/migrations"
)

type Suite struct {
	t          *testing.T
	AuthClient gen.AuthClient
	Cfg        *config.Config
}

func New(t *testing.T) (context.Context, *Suite) {
	ctx := context.Background()

	t.Helper()
	t.Parallel()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfigByPath("../config/config_tests.yml")
	if err != nil {
		t.Fatalf("init config failed:%s", err)
	}

	cfg.GRPC.Port, err = findFreePort()
	if err != nil {
		t.Fatalf("failed to find free port for grpc")
	}

	cfg.GRPCGateway.Port, err = findFreePort()
	if err != nil {
		t.Fatalf("failed to find free port for grpc gateway")
	}

	if err = TestPostgresDB(ctx, cfg); err != nil {
		t.Fatalf("failed to start test postgres container:%s", err)
	}
	go func() {
		app.RunGRPC(cfg, logger)
	}()

	time.Sleep(1 * time.Second)

	clientConn, err := grpc.DialContext(ctx, net.JoinHostPort("localhost", cfg.GRPC.Port), grpc.WithTransportCredentials(insecure.NewCredentials())) //небeзопасное соед.для тестов
	if err != nil {
		t.Fatalf("client conn failed:%s", err)
	}

	return ctx, &Suite{
		t:          t,
		AuthClient: gen.NewAuthClient(clientConn),
		Cfg:        cfg,
	}
}

func TestPostgresDB(ctx context.Context, cfg *config.Config) error {
	pgContainer, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:16.2"),
		postgres.WithDatabase(cfg.Postgres.DbName),
		postgres.WithUsername(cfg.Postgres.User),
		postgres.WithPassword(cfg.Postgres.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return fmt.Errorf("run Container failed:%s", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("conn to container failed:%s", err)
	}

	portWithTcp, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return fmt.Errorf("get mapped port failed: %s", err)
	}

	cfg.Postgres.Port = strings.TrimSuffix(string(portWithTcp), "/tcp")

	if err = applyMigrations(connStr, migrationPath); err != nil {
		return fmt.Errorf("applyMigrations failed:%s", err)
	}

	return nil
}

func applyMigrations(connStr, migrationPath string) error {
	migrations, err := migrate.New(
		"file://"+migrationPath,
		connStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer func() {
		sourceErr, dbErr := migrations.Close() //для освобождения рес-ов и предотвращение их утечки
		if sourceErr != nil {
			logrus.Errorf("close migrations failed:%s", sourceErr)
		}
		if dbErr != nil {
			logrus.Errorf("close migrations failed:%s", dbErr)
		}
	}()

	if err = migrations.Up(); err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations re already applied:%s", err.Error())
	}
	return nil
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
