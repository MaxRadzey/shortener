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

func (p *PostgresStorage) CreateBatch(ctx context.Context, items []BatchItem) error {
	// Используем транзакцию для атомарности
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Подготавливаем batch insert с множественными VALUES
	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue("INSERT INTO urls (short_path, original_url) VALUES ($1, $2) ON CONFLICT (short_path) DO NOTHING", item.ShortPath, item.FullURL)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// Выполняем все запросы
	for i := 0; i < len(items); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert batch item: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
