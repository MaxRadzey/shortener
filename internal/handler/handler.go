package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
)

type Handler struct {
	Storage dbstorage.UrlStorage
}

// CreateUrl хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func (h *Handler) CreateUrl(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		c.String(http.StatusBadRequest, "Invalid Content-Type!")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error occurred while reading body!")
		return
	}

	text := string(body)
	if !utils.IsValidURL(text) {
		c.String(http.StatusBadRequest, "Invalid Body!")
		return
	}

	shortPath, err := utils.GetShortPath(text)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	h.Storage.Create(shortPath, text)

	result := fmt.Sprintf("http://%s/%s", c.Request.Host, shortPath)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusCreated, result)
}

// GetUrl хэндлер, обрабатывает GET-запросы, получает в качестве параметра маршрута сокращенное значение URL,
// ищет в БД совпадение длинного пути и производит редирект на него (307), иначе отдает (404) ошибку.
func (h *Handler) GetUrl(c *gin.Context) {
	shortPath := c.Param("short_path")
	longUrl := h.Storage.Get(shortPath)

	if longUrl == "" {
		c.String(http.StatusNotFound, "Not found!")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longUrl)
}
