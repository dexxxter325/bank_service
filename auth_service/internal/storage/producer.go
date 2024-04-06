package storage

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaProducerPostgres struct {
	db *pgxpool.Pool
}

func NewKafkaProducerPostgres(db *pgxpool.Pool) *KafkaProducerPostgres {
	return &KafkaProducerPostgres{db: db}
}

func (p *KafkaProducerPostgres) PushUserIDs(ctx context.Context, writer *kafka.Writer) (err error) {
	sentID := make(map[int64]bool)

	query := "select id from users"

	rows, err := p.db.Query(ctx, query)
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
