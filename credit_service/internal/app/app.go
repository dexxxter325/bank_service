package app

import (
	"bank/credit_service/internal/config"
	"bank/credit_service/internal/kafka/consumer"
	"bank/credit_service/internal/rest"
	"bank/credit_service/internal/service"
	"bank/credit_service/internal/storage"
	"bank/credit_service/pkg/mongodb"
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

	db, err := mongodb.ConnToMongoDB(cfg)
	if err != nil {
		logger.Fatalf("connect to mongo failed:%s", err)
	}

	defer func() {
		if err = db.Client().Disconnect(context.Background()); err != nil {
			logger.Fatalf("mongodb close failed:%s", err)
		}
		logger.Info("mongodb connection closed")
	}()

	storages := storage.NewStorage(db, cfg.MongoDb.CreditCollection, cfg.MongoDb.UserIDCollection)
	services := service.NewService(logger, storages)
	handlers := rest.NewHandler(logger, services)
	kc := consumer.NewKafkaConsumer(storages)

	srv := &http.Server{
		Addr:    ":" + cfg.Rest.Port,
		Handler: handlers.InitRoutes(r),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		logger.Infof("kafka consumer starting:%s", cfg.Kafka.Brokers)
		if err = kc.KafkaConsumer(context.Background(), cfg, logger, db.Client(), stop); err != nil {
			logger.Fatalf("failed to start kafka consumer:%s", err)
		}
	}()

	go func() {
		logger.Infof("rest starting on port:%s", cfg.Rest.Port)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logger.Fatalf("run application failed:%s", err)
		}
	}()

	<-stop

	if err = srv.Shutdown(context.Background()); err != nil {
		logger.Fatalf("shutdown rest failed:%s", err)
	}

	logger.Info("kafka consumer stopped")
	logger.Info("rest server stopped")
}
