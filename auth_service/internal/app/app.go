package app

import (
	"bank/auth_service/gen"
	"bank/auth_service/internal/config"
	"bank/auth_service/internal/handler"
	"bank/auth_service/internal/service"
	"bank/auth_service/internal/storage"
	"bank/auth_service/pkg/postgres"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
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
		logrus.Info("postgres connection closed")
	}()

	storages := storage.NewStorage(db)
	services := service.NewService(storages, logger)
	handlers := handler.NewAuthServer(services)

	srv := grpc.NewServer()
	gen.RegisterAuthServer(srv, handlers)

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
		if err != nil {
			logger.Fatalf("listen grpc failed:%s", err)
		}
		if err := srv.Serve(listener); err != nil {
			logger.Fatalf("failed to serve grpc:%s", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	srv.GracefulStop()

	logger.Info("grpc stopped")
}
