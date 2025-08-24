package app

import (
	gzipMiddleware "github.com/MaxRadzey/shortener/internal/gzip"
	"github.com/MaxRadzey/shortener/internal/logger"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/config"
	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	storage, err := dbstorage.NewStorage(AppConfig.FilePath)
	if err != nil {
		return err
	}

	handler := &httphandlers.Handler{Storage: storage, AppConfig: *AppConfig}

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

	return r
}
