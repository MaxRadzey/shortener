package storage

import (
	"context"

	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// StorageResult содержит результат инициализации хранилища.
type StorageResult struct {
	Storage URLStorage
	DB      *pgxpool.Pool
}

// InitializeStorage выбирает и инициализирует хранилище согласно приоритетам:
// 1. PostgreSQL (если указан DATABASE_DSN)
// 2. Файловое хранилище (если указан FILE_PATH)
// 3. In-memory (fallback)
// Возвращает выбранное хранилище и пул соединений БД (может быть nil).
func InitializeStorage(databaseDSN, filePath string) (*StorageResult, error) {
	var storage URLStorage
	var db *pgxpool.Pool
	var err error

	// Приоритет 1: PostgreSQL
	if databaseDSN != "" {
		logger.Log.Info("Attempting to connect to PostgreSQL", zap.String("dsn", utils.MaskDSN(databaseDSN)))
		db, err = initDatabase(databaseDSN)
		if err == nil && db != nil {
			// Запустить миграции используя тот же DSN
			if err := RunMigrations(databaseDSN); err != nil {
				logger.Log.Warn("Migrations failed", zap.Error(err))
			} else {
				logger.Log.Info("Migrations completed successfully")
			}

			postgresStorage, err := NewPostgresStorage(db)
			if err == nil {
				storage = postgresStorage
				logger.Log.Info("PostgreSQL storage initialized")
			} else {
				logger.Log.Warn("Failed to initialize PostgreSQL storage", zap.Error(err))
			}
		} else {
			logger.Log.Warn("Failed to connect to PostgreSQL, will try fallback storage", zap.Error(err))
		}
	}

	// Приоритет 2: Файловое хранилище
	if storage == nil && filePath != "" {
		logger.Log.Info("Attempting to use file storage", zap.String("path", filePath))
		fileStorage, err := NewStorage(filePath)
		if err == nil {
			storage = fileStorage
			logger.Log.Info("File storage initialized")
		} else {
			logger.Log.Warn("Failed to initialize file storage", zap.Error(err))
		}
	}

	// Приоритет 3: In-memory (fallback)
	if storage == nil {
		logger.Log.Info("Using in-memory storage as fallback")
		storage = NewMemoryStorage()
	}

	// Логируем финальный выбор хранилища
	switch storage.(type) {
	case *PostgresStorage:
		logger.Log.Info("Storage selected: PostgreSQL")
	case *Storage:
		logger.Log.Info("Storage selected: File")
	case *MemoryStorage:
		logger.Log.Info("Storage selected: In-Memory")
	}

	return &StorageResult{
		Storage: storage,
		DB:      db,
	}, nil
}

// initDatabase создает подключение к PostgreSQL, если указан DSN.
// Возвращает пул соединений или nil, если DSN не указан или подключение не удалось.
func initDatabase(dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		logger.Log.Debug("Database DSN is empty, skipping PostgreSQL connection")
		return nil, nil
	}

	logger.Log.Debug("Creating PostgreSQL connection pool")
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Log.Warn("Failed to create connection pool", zap.Error(err))
		return nil, nil
	}

	logger.Log.Info("Database connection successful")
	return db, nil
}
