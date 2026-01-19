package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}
