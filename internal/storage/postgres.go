package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
			return "", &ErrNotFound{ShortPath: short}
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return originalURL, nil
}

func (p *PostgresStorage) Create(item URLEntry) error {
	ctx := context.Background()

	_, err := p.db.Exec(ctx, "INSERT INTO urls (short_path, original_url, user_id) VALUES ($1, $2, $3)", item.ShortPath, item.FullURL, item.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var existingShortPath string
			queryErr := p.db.QueryRow(ctx, "SELECT short_path FROM urls WHERE original_url = $1", item.FullURL).Scan(&existingShortPath)
			if queryErr != nil {
				return fmt.Errorf("failed to get existing short_path: %w", queryErr)
			}
			return &ErrURLAlreadyExists{ShortPath: existingShortPath}
		}
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (p *PostgresStorage) CreateBatch(ctx context.Context, items []URLEntry) error {
	// Используем транзакцию для атомарности
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue("INSERT INTO urls (short_path, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (short_path) DO NOTHING", item.ShortPath, item.FullURL, item.UserID)
	}

	results := tx.SendBatch(ctx, batch)

	// Выполняем все запросы
	for i := 0; i < len(items); i++ {
		_, err := results.Exec()
		if err != nil {
			results.Close()
			tx.Rollback(ctx)
			return fmt.Errorf("failed to insert batch item: %w", err)
		}
	}

	if err := results.Close(); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PostgresStorage) GetByUserID(ctx context.Context, userID string) ([]UserURL, error) {
	rows, err := p.db.Query(ctx, "SELECT short_path, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get URLs by user: %w", err)
	}
	defer rows.Close()

	var out []UserURL
	for rows.Next() {
		var u UserURL
		if err := rows.Scan(&u.ShortPath, &u.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}
