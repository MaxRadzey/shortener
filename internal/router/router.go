package router

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/handler"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter создаёт роутер с логгером, gzip, auth и маршрутами коротких ссылок.
func SetupRouter(h *handler.Handler, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	SetupMiddleware(r)

	r.Use(logger.HTTPLogger())

	r.Use(middleware.Gzip())
	r.Use(middleware.Auth(cfg.SigningKey))

	r.POST("/", h.CreateURL)
	r.GET("/:short_path", h.GetURL)
	r.POST("/api/shorten", h.GetURLJSON)
	r.POST("/api/shorten/batch", h.CreateURLBatch)
	r.GET("/ping", h.Ping)
	r.GET("/api/user/urls", h.GetUserURLs)
	r.DELETE("/api/user/urls", h.DeleteURLs)

	return r
}

// SetupMiddleware вешает обработку NoMethod (405) на роутер.
func SetupMiddleware(router *gin.Engine) {
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})
}
