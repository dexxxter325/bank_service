package producer

import (
	"bank/auth_service/internal/config"
	"bank/auth_service/internal/service"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type KafkaProducer struct {
	storage service.Storage
}

func NewKafkaProducer(storage service.Storage) *KafkaProducer {
	return &KafkaProducer{storage: storage}
}

func (k *KafkaProducer) KafkaProducer(logger *logrus.Logger, cfg *config.Config, db *pgxpool.Pool) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	writer := kafka.Writer{ //send messages to broker
		Addr:      kafka.TCP(cfg.Kafka.Brokers),
		Topic:     cfg.Kafka.Topic,
		Balancer:  &kafka.LeastBytes{}, //балансировщик,выбирающий партицию,с наименьшим объемом неподтвержденных данных
		BatchSize: 1,                   //кол-во сообщений ,отправленных за раз
		Async:     true,                //асинхронная отправка сообщений(не ждем подтверждения об успешной отправке)
	}

	defer func() {
		db.Close()
		logger.Info("postgres connection in kafka closed")
		if err := writer.Close(); err != nil {
			logger.Fatalf("failed to close kafka writer:%s", err)
		}
		logger.Info("kafka reader closed")
	}()

	// Запускаем бесконечный цикл для непрерывного мониторинга PostgreSQL
	for {
		select {
		case <-time.After(5 * time.Second): //отпр. данных каждые 5с
			if err := k.storage.PushUserIDs(context.Background(), &writer); err != nil {
				logger.Infof("Error pushing user IDs to Kafka: %v", err)
			}
		case <-stop:
			logger.Info("postgres,kafka reader shutting down...")
			return nil
		}
	}
}
