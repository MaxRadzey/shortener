package app

import (
	"context"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/db"
	gzipMiddleware "github.com/MaxRadzey/shortener/internal/gzip"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	ctx := context.Background()

	storage, err := dbstorage.NewStorage(AppConfig.FilePath)
	if err != nil {
		return err
	}

	var dbpool *pgxpool.Pool
	if AppConfig.DBConfig.DSN != "" {
		dbpool, err = db.NewPool(ctx, AppConfig.DBConfig.DSN)
		if err != nil {
			return err
		}
		defer dbpool.Close()
	}

	handler := &httphandlers.Handler{Storage: storage, AppConfig: *AppConfig, DBPool: dbpool}

	if err := logger.Initialize(AppConfig.LogLevel); err != nil {
		return err
	}

	r := SetupRouter(handler)

	return r.Run(AppConfig.Address)
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
