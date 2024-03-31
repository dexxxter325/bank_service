package producer

import (
	"bank/auth_service/internal/config"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	sentID = make(map[int64]bool)
)

func KafkaProducer(logger *logrus.Logger, cfg *config.Config, db *pgxpool.Pool) error {
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
			if err := pushUserIDs(context.Background(), db, &writer); err != nil {
				logger.Infof("Error pushing user IDs to Kafka: %v", err)
			}
		case <-stop:
			logger.Info("postgres,kafka reader shutting down...")
			return nil
		}
	}
}

func pushUserIDs(ctx context.Context, db *pgxpool.Pool, writer *kafka.Writer) error {
	query := "select id from users"

	rows, err := db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("error querying user IDs: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var userID int64

		if err = rows.Scan(&userID); err != nil {
			return fmt.Errorf("error scanning user ID: %w", err)
		}

		if sentID[userID] { // Если ID уже был отправлен, пропускаем его
			continue //возвращаемся в начало цикла for
		}
		// Добавляем ID в мапу отправленных ID
		sentID[userID] = true

		userIDBytes := make([]byte, 8)

		binary.BigEndian.PutUint64(userIDBytes, uint64(userID))

		// Отправляем userID в брокер
		err = writer.WriteMessages(ctx, kafka.Message{
			Value: userIDBytes,
		})
		if err != nil {
			return fmt.Errorf("error writing message to Kafka: %s", err)
		}

		logrus.Infof("UserID %v успешно отправлен в брокер", userID)
	}

	if rows.Err() != nil {
		return fmt.Errorf("error iterating over user IDs: %s", err)
	}

	return nil
}
