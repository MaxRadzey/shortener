package router

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter создает и настраивает HTTP роутер со всеми middleware и маршрутами.
func SetupRouter(h *handler.Handler, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	SetupMiddleware(r)

	r.Use(logger.RequestLogger())
	r.Use(logger.ResponseLogger())
	r.Use(middleware.Gzip())
	r.Use(middleware.Auth(cfg.SigningKey))

	r.POST("/", h.CreateURL)
	r.GET("/:short_path", h.GetURL)
	r.POST("/api/shorten", h.GetURLJSON)
	r.POST("/api/shorten/batch", h.CreateURLBatch)
	r.GET("/api/user/urls", h.GetUserURLs)
	r.GET("/ping", h.Ping)

	return r
}

// SetupMiddleware настраивает middleware для роутера.
func SetupMiddleware(router *gin.Engine) {
	router.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed!")
	})
}
