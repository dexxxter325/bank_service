package storage

import "github.com/jackc/pgx/v5/pgxpool"

type Postgres struct {
	*AuthPostgres
	*KafkaProducerPostgres
}

func NewStorage(db *pgxpool.Pool) *Postgres {
	return &Postgres{
		AuthPostgres:          NewAuthPostgres(db),
		KafkaProducerPostgres: NewKafkaProducerPostgres(db),
	}
}
