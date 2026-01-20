package app

import (
	"context"
	"net/http"

	gzipMiddleware "github.com/MaxRadzey/shortener/internal/gzip"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/migrations"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/gin-gonic/gin"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	if err := logger.Initialize(AppConfig.LogLevel); err != nil {
		return err
	}

	storage, db := initializeStorage(AppConfig)

	urlService := service.NewService(storage, *AppConfig, db)
	handler := &httphandlers.Handler{
		Service: urlService,
	}

	r := SetupRouter(handler)

	logger.Log.Info("Starting HTTP server", zap.String("address", AppConfig.Address))
	return r.Run(AppConfig.Address)
}

// initializeStorage выбирает и инициализирует хранилище согласно приоритетам:
// 1. PostgreSQL (если указан DATABASE_DSN)
// 2. Файловое хранилище (если указан FILE_PATH)
// 3. In-memory (fallback)
// Возвращает выбранное хранилище и пул соединений БД (может быть nil).
func initializeStorage(AppConfig *config.Config) (dbstorage.URLStorage, *pgxpool.Pool) {
	var storage dbstorage.URLStorage
	var db *pgxpool.Pool
	var err error

	// Приоритет 1: PostgreSQL
	if AppConfig.DatabaseDSN != "" {
		logger.Log.Info("Attempting to connect to PostgreSQL", zap.String("dsn", utils.MaskDSN(AppConfig.DatabaseDSN)))
		db, err = initDatabase(AppConfig.DatabaseDSN)
		if err == nil && db != nil {
			// Запустить миграции используя тот же DSN
			if err := migrations.Run(AppConfig.DatabaseDSN); err != nil {
				logger.Log.Warn("Migrations failed", zap.Error(err))
			} else {
				logger.Log.Info("Migrations completed successfully")
			}

			postgresStorage, err := dbstorage.NewPostgresStorage(db)
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
	if storage == nil && AppConfig.FilePath != "" {
		logger.Log.Info("Attempting to use file storage", zap.String("path", AppConfig.FilePath))
		fileStorage, err := dbstorage.NewStorage(AppConfig.FilePath)
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
		storage = dbstorage.NewMemoryStorage()
	}

	// Логируем финальный выбор хранилища
	switch storage.(type) {
	case *dbstorage.PostgresStorage:
		logger.Log.Info("Storage selected: PostgreSQL")
	case *dbstorage.Storage:
		logger.Log.Info("Storage selected: File")
	case *dbstorage.MemoryStorage:
		logger.Log.Info("Storage selected: In-Memory")
	}

	return storage, db
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

func SetupMiddleware(router *gin.Engine) {
	router.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed!")
	})
}

func SetupRouter(handler *httphandlers.Handler) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	SetupMiddleware(r)

	r.Use(logger.RequestLogger())
	r.Use(logger.ResponseLogger())
	r.Use(gzipMiddleware.GzipMiddleware())

	r.POST("/", handler.CreateURL)
	r.GET("/:short_path", handler.GetURL)
	r.POST("/api/shorten", handler.GetURLJSON)
	r.POST("/api/shorten/batch", handler.CreateURLBatch)
	r.GET("/ping", handler.Ping)

	return r
}
