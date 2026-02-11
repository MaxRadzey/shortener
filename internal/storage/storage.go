package storage

import (
	"context"

	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/repository"
	"github.com/MaxRadzey/shortener/internal/repository/file"
	"github.com/MaxRadzey/shortener/internal/repository/memory"
	"github.com/MaxRadzey/shortener/internal/repository/postgres"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// StorageResult содержит результат инициализации хранилища.
type StorageResult struct {
	Repository repository.URLRepository
	DB         *pgxpool.Pool
}

// InitializeStorage выбирает и инициализирует хранилище согласно приоритетам:
// 1. PostgreSQL (если указан DATABASE_DSN)
// 2. Файловое хранилище (если указан FILE_PATH)
// 3. In-memory (fallback)
// Возвращает выбранный репозиторий и пул соединений БД (может быть nil).
func InitializeStorage(databaseDSN, filePath string) (*StorageResult, error) {
	var repo repository.URLRepository
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

			postgresRepo, err := postgres.NewPostgresRepository(db)
			if err == nil {
				repo = postgresRepo
				logger.Log.Info("PostgreSQL repository initialized")
			} else {
				logger.Log.Warn("Failed to initialize PostgreSQL repository", zap.Error(err))
			}
		} else {
			logger.Log.Warn("Failed to connect to PostgreSQL, will try fallback repository", zap.Error(err))
		}
	}

	// Приоритет 2: Файловое хранилище
	if repo == nil && filePath != "" {
		logger.Log.Info("Attempting to use file repository", zap.String("path", filePath))
		fileRepo, err := file.NewFileRepository(filePath)
		if err == nil {
			repo = fileRepo
			logger.Log.Info("File repository initialized")
		} else {
			logger.Log.Warn("Failed to initialize file repository", zap.Error(err))
		}
	}

	// Приоритет 3: In-memory (fallback)
	if repo == nil {
		logger.Log.Info("Using in-memory repository as fallback")
		repo = memory.NewMemoryRepository()
	}

	// Логируем финальный выбор репозитория
	switch repo.(type) {
	case *postgres.PostgresRepository:
		logger.Log.Info("Repository selected: PostgreSQL")
	case *file.FileRepository:
		logger.Log.Info("Repository selected: File")
	case *memory.MemoryRepository:
		logger.Log.Info("Repository selected: In-Memory")
	}

	return &StorageResult{
		Repository: repo,
		DB:         db,
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
