package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Postgres {
	return &Postgres{db: db}
}

func (s *Postgres) Register(ctx context.Context, username, password string) (int64, error) {
	panic("")
}

func (s *Postgres) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	panic("")
}

func (s *Postgres) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	panic("")
}
