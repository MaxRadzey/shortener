package app

import (
	"context"
	"net/http"

	gzipMiddleware "github.com/MaxRadzey/shortener/internal/gzip"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	storage, err := dbstorage.NewStorage(AppConfig.FilePath)
	if err != nil {
		return err
	}

	db, err := initDatabase(AppConfig.DatabaseDSN)
	if err != nil {
		return err
	}

	urlService := service.NewService(storage, *AppConfig, db)
	handler := &httphandlers.Handler{
		Service: urlService,
	}

	if err := logger.Initialize(AppConfig.LogLevel); err != nil {
		return err
	}

	r := SetupRouter(handler)

	return r.Run(AppConfig.Address)
}

// initDatabase создает подключение к PostgreSQL, если указан DSN.
// Возвращает пул соединений или nil, если DSN не указан или подключение не удалось.
func initDatabase(dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, nil
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, nil
	}

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
	r.GET("/ping", handler.Ping)

	return r
}
