package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// RunMigrations применяет SQL-миграции к БД (dsn); migrationsPath по умолчанию "migrations".
func RunMigrations(dsn string, migrationsPath string) error {
	if dsn == "" {
		return nil
	}

	logger.Log.Info("Starting database migrations")

	// Если путь к миграциям не указан, используем дефолтный
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	// Получаем абсолютный путь к директории migrations
	absMigrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		logger.Log.Error("Failed to get migrations path", zap.Error(err))
		return fmt.Errorf("failed to get migrations path: %w", err)
	}
	logger.Log.Debug("Migrations path", zap.String("path", absMigrationsPath))

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Error("Failed to open database for migrations", zap.Error(err))
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Создаем экземпляр драйвера PostgreSQL для миграций
	instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Error("Failed to create postgres instance for migrations", zap.Error(err))
		return fmt.Errorf("failed to create postgres instance: %w", err)
	}

	// Создаем экземпляр мигратора
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absMigrationsPath),
		"postgres", instance)
	if err != nil {
		logger.Log.Error("Failed to create migrate instance", zap.Error(err))
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Запускаем миграции
	logger.Log.Info("Running migrations")
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Log.Error("Migrations failed", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Log.Info("Migrations already applied, no changes needed")
	} else {
		logger.Log.Info("Migrations completed successfully")
	}

	return nil
}
