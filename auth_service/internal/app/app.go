package app

import (
	"bank/auth_service/gen"
	"bank/auth_service/internal/config"
	"bank/auth_service/internal/handler"
	"bank/auth_service/internal/kafka/producer"
	"bank/auth_service/internal/service"
	"bank/auth_service/internal/storage"
	"bank/auth_service/pkg/postgres"
	"context"
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func RunGRPC(cfg *config.Config, logger *logrus.Logger) {
	db, err := postgres.NewPostgres(cfg)
	if err != nil {
		logger.Fatalf("failed to connect to postgres:%s", err)
	}
	defer func() {
		db.Close()
		logger.Info("postgres connection closed")
	}()

	storages := storage.NewStorage(db)
	services := service.NewService(storages, logger, cfg)
	handlers := handler.NewAuthServer(services, logger)
	kp := producer.NewKafkaProducer(storages)

	srv := grpc.NewServer()
	gen.RegisterAuthServer(srv, handlers)

	go func() {
		logger.Infof("kafka producer starting:%s", cfg.Kafka.Brokers)
		if err = kp.KafkaProducer(logger, cfg, db); err != nil {
			logger.Fatalf("failed to start kafka producer:%s", err)
		}
	}()

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
		if err != nil {
			logger.Fatalf("listen grpc failed:%s", err)
		}
		logger.Infof("gRPC starting on port:%s", cfg.GRPC.Port)
		if err := srv.Serve(listener); err != nil {
			logger.Fatalf("failed to serve grpc:%s", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	shutDownGrpcGateway := RunGrpcGateway(context.Background(), logger, cfg)
	defer shutDownGrpcGateway()

	<-stop

	srv.GracefulStop()

	logger.Info("kafka producer stopped")
	logger.Info("grpc stopped")
}

func RunGrpcGateway(ctx context.Context, logger *logrus.Logger, cfg *config.Config) func() {
	// Register gRPC server endpoint
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := gen.RegisterAuthHandlerFromEndpoint(ctx, mux, "auth_service:"+cfg.GRPC.Port, opts); err != nil { ////docker host
		logger.Fatalf("register GRPC gateway failed:%s", err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	srv := &http.Server{
		Addr:    ":" + cfg.GRPCGateway.Port,
		Handler: mux,
	}

	go func() {
		logger.Infof("GrpcGateway started on port:%s", cfg.GRPCGateway.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logger.Fatalf("run GrpcGateway failed:%s", err)
		}
	}()

	shutDownGrpcGateway := func() {
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("shutdown grpc gateway failed:%s", err)
		}
		logger.Info("grpc gateway stopped")
	}

	return shutDownGrpcGateway
}
