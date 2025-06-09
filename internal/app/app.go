package app

import (
	"net/http"

	"github.com/MaxRadzey/shortener/internal/config"

	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

// Run запускает http сервер.
func Run(AppConfig *config.Config) error {
	storage := dbstorage.NewStorage()
	handler := &httphandlers.Handler{Storage: storage, AppConfig: *AppConfig}

	r := SetupRouter(handler)

	return r.Run(AppConfig.Address)
}

func SetupRouter(handler *httphandlers.Handler) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true

	r.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed!")
	})

	r.POST("/", handler.CreateURL)
	r.GET("/:short_path", handler.GetURL)

	return r
}
