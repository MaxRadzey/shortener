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

	// Роуты с автоматической аутентификацией (создают пользователя, если куки нет)
	authenticated := r.Group("")
	authenticated.Use(middleware.Auth(cfg.SigningKey))
	authenticated.POST("/", h.CreateURL)
	authenticated.GET("/:short_path", h.GetURL)
	authenticated.POST("/api/shorten", h.GetURLJSON)
	authenticated.POST("/api/shorten/batch", h.CreateURLBatch)
	authenticated.GET("/ping", h.Ping)

	// Защищенные роуты, требующие валидную куку
	protected := r.Group("")
	protected.Use(middleware.RequireAuth(cfg.SigningKey))
	protected.GET("/api/user/urls", h.GetUserURLs)

	return r
}

// SetupMiddleware настраивает middleware для роутера.
func SetupMiddleware(router *gin.Engine) {
	router.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed!")
	})
}
