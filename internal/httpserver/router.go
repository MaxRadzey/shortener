// Package httpserver: router настраивает маршруты и middleware для HTTP-сервера.
package httpserver

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/httpserver/handler"
	"github.com/MaxRadzey/shortener/internal/httpserver/middleware"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// SetupRouter создаёт роутер с логгером, gzip, auth и маршрутами коротких ссылок.
func SetupRouter(h *handler.Handler, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	setupMiddleware(r)

	r.Use(logger.HTTPLogger())
	r.Use(middleware.Gzip())
	r.Use(middleware.Auth(cfg.SigningKey))

	r.POST("/", h.CreateURL)
	r.GET("/:short_path", h.GetURL)
	r.POST("/api/shorten", h.GetURLJSON)
	r.POST("/api/shorten/batch", h.CreateURLBatch)
	r.GET("/ping", h.Ping)
	r.GET("/api/internal/stats", h.GetInternalStats)
	r.GET("/api/user/urls", h.GetUserURLs)
	r.DELETE("/api/user/urls", h.DeleteURLs)

	if cfg.DevMode {
		r.GET("/swagger/doc.json", func(c *gin.Context) {
			doc, _ := swag.ReadDoc("swagger")
			c.Data(http.StatusOK, "application/json", []byte(doc))
		})
	}

	return r
}

func setupMiddleware(r *gin.Engine) {
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})
}
