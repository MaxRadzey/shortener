package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(db *pgxpool.Pool) (*PostgresStorage, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Проверяем соединение
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{
		db: db,
	}, nil
}

func (p *PostgresStorage) Get(short string) (string, error) {
	ctx := context.Background()
	var originalURL string

	err := p.db.QueryRow(ctx, "SELECT original_url FROM urls WHERE short_path = $1", short).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return originalURL, nil
}

func (p *PostgresStorage) Create(short, full string) error {
	ctx := context.Background()

	_, err := p.db.Exec(ctx, "INSERT INTO urls (short_path, original_url) VALUES ($1, $2) ON CONFLICT (short_path) DO NOTHING", short, full)
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}
