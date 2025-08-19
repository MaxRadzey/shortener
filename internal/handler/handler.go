package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaxRadzey/shortener/internal/config"
	"github.com/MaxRadzey/shortener/internal/models"

	"github.com/gin-gonic/gin"

	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
)

type Handler struct {
	Storage   dbstorage.URLStorage
	AppConfig config.Config
}

// CreateURL хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func (h *Handler) CreateURL(c *gin.Context) {
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

	result := fmt.Sprintf("%s/%s", h.AppConfig.ReturningAddress, shortPath)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusCreated, result)
}

// GetURL хэндлер, обрабатывает GET-запросы, получает в качестве параметра маршрута сокращенное значение URL,
// ищет в БД совпадение длинного пути и производит редирект на него (307), иначе отдает (404) ошибку.
func (h *Handler) GetURL(c *gin.Context) {
	shortPath := c.Param("short_path")
	longURL := h.Storage.Get(shortPath)

	if longURL == "" {
		c.String(http.StatusNotFound, "Not found!")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (h *Handler) GetUrlJSON(c *gin.Context) {
	var req models.Request

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	if req.URL == "" {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	shortPath, err := utils.GetShortPath(req.URL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	h.Storage.Create(shortPath, req.URL)

	resp := models.Response{Result: fmt.Sprintf("%s/%s", h.AppConfig.ReturningAddress, shortPath)}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(c.Writer).Encode(resp); err != nil {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}
}
