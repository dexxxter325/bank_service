package storage

import (
	"bank/auth_service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthPostgres struct {
	db *pgxpool.Pool
}

func NewAuthPostgres(db *pgxpool.Pool) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (p *AuthPostgres) SaveUser(ctx context.Context, username string, hashedPassword []byte) (userId int64, err error) {
	query := "insert into users (username,password) values($1,$2) returning id"

	row := p.db.QueryRow(ctx, query, username, hashedPassword)

	if err = row.Scan(&userId); err != nil {
		return 0, fmt.Errorf("failed to scan in saveUser:%s", err)
	}

	return userId, err
}

func (p *AuthPostgres) GetUserByUsername(ctx context.Context, username string) (user models.User, err error) {
	query := "select * from users where username=$1"

	row := p.db.QueryRow(ctx, query, username)

	if err = row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found with username:%s", username)
		}
		return models.User{}, fmt.Errorf("failed to scan in getUserByUsername:%s", err)
	}

	return user, nil
}

func (p *AuthPostgres) GetUserById(ctx context.Context, userId int64) (user models.User, err error) {
	query := "select * from users where id=$1"

	row := p.db.QueryRow(ctx, query, userId)

	if err = row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found with id:%v", userId)
		}
		return models.User{}, fmt.Errorf("failed to scan in getUserById:%s", err)
	}

	return user, nil
}
