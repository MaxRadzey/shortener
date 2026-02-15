package router

import (
	"net/http"
	"net/http/pprof"

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
	r.GET("/ping", h.Ping)
	r.GET("/api/user/urls", h.GetUserURLs)
	r.DELETE("/api/user/urls", h.DeleteURLs)

	if cfg.DevMode {
		setupPprof(r)
	}

	return r
}

// setupPprof регистрирует эндпоинты pprof по /debug/pprof/. Вызывается только при включённом DevMode.
func setupPprof(r *gin.Engine) {
	gr := r.Group("/debug/pprof")
	{
		gr.GET("/", gin.WrapF(pprof.Index))
		gr.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		gr.GET("/profile", gin.WrapF(pprof.Profile))
		gr.GET("/symbol", gin.WrapF(pprof.Symbol))
		gr.GET("/trace", gin.WrapF(pprof.Trace))
		gr.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		gr.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		gr.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		gr.GET("/block", gin.WrapH(pprof.Handler("block")))
		gr.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	}
}

// SetupMiddleware настраивает middleware для роутера.
func SetupMiddleware(router *gin.Engine) {
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})
}
