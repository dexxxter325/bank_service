package app

import (
	"bank/credit_service/internal/config"
	"bank/credit_service/internal/rest"
	"bank/credit_service/internal/service"
	"bank/credit_service/internal/storage"
	"bank/credit_service/pkg/mongo"
	"context"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func RunRest(cfg *config.Config, logger *logrus.Logger) {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	db, err := mongo.ConnToMongoDB(cfg)
	if err != nil {
		logger.Fatalf("connect to mongo failed:%s", err)
	}

	storages := storage.NewStorage(db, cfg.MongoDb.CreditCollection)
	services := service.NewService(logger, storages)
	handlers := rest.NewHandler(logger, services)

	srv := &http.Server{
		Addr:    ":" + cfg.Rest.Port,
		Handler: handlers.InitRoutes(r),
	}

	go func() {
		logger.Infof("rest started on port:%s", cfg.Rest.Port)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logger.Fatalf("run application failed:%s", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err = srv.Shutdown(context.Background()); err != nil {
		logger.Fatalf("shutdown rest failed:%s", err)
	}

	logger.Info("rest server stopped")
}
