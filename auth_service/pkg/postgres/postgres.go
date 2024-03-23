package postgres

import (
	"bank/auth_service/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	data := fmt.Sprintf("host=%s port=%v user=%s dbname=%s password=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DbName, cfg.Postgres.Password, cfg.Postgres.Sslmode)

	conn, err := pgxpool.New(context.Background(), data)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres:%s", err)
	}

	return conn, nil
}
