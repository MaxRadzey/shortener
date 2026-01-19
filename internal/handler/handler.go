package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *service.Service
}

// CreateURL хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func (h *Handler) CreateURL(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error occurred while reading body!")
		return
	}

	text := string(body)

	result, err := h.Service.CreateShortURL(text)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			c.String(http.StatusBadRequest, "Invalid Body!")
			return
		}
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusCreated, result)
}

// GetURL хэндлер, обрабатывает GET-запросы, получает в качестве параметра маршрута сокращенное значение URL,
// ищет в БД совпадение длинного пути и производит редирект на него (307), иначе отдает (404) ошибку.
func (h *Handler) GetURL(c *gin.Context) {
	shortPath := c.Param("short_path")
	longURL, err := h.Service.GetLongURL(shortPath)

	if err != nil {
		c.String(http.StatusNotFound, "Not found!")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (h *Handler) GetURLJSON(c *gin.Context) {
	var req models.Request

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.Service.CreateShortURL(req.URL)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			c.String(http.StatusBadRequest, "invalid request")
			return
		}
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	resp := models.Response{Result: result}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(c.Writer).Encode(resp); err != nil {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}
}

// Ping хендлер проверяет соединение с базой данных.
// Возвращает HTTP 200 OK при успешной проверке, 500 Internal Server Error при неуспешной.
func (h *Handler) Ping(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.Service.Ping(ctx); err != nil {
		c.String(http.StatusInternalServerError, "Database connection failed")
		return
	}

	c.String(http.StatusOK, "OK")
}
