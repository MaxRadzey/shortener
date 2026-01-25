package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Service *service.Service
}

// sendJSONResponse отправляет JSON ответ и обрабатывает ошибки кодирования
func (h *Handler) sendJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(statusCode)
	if err := json.NewEncoder(c.Writer).Encode(data); err != nil {
		logger.Log.Error("Failed to encode JSON response", zap.Error(err))
		c.String(http.StatusInternalServerError, "Internal server error!")
	}
}

// CreateURL хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func (h *Handler) CreateURL(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("Failed to create URL", zap.Error(err))
		c.String(http.StatusInternalServerError, "Error occurred while reading body!")
		return
	}

	text := string(body)

	if !utils.IsValidURL(text) {
		c.String(http.StatusBadRequest, "Invalid Body!")
		return
	}

	result, err := h.Service.CreateShortURL(text)
	if err != nil {
		// Проверяем, является ли ошибка конфликтом существующего URL
		var conflictErr *service.ErrURLConflict
		if errors.As(err, &conflictErr) {
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusConflict, conflictErr.ShortURL)
			return
		}
		logger.Log.Error("Failed to create URL", zap.Error(err))
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

	if !utils.IsValidURL(req.URL) {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.Service.CreateShortURL(req.URL)
	if err != nil {
		// Проверяем, является ли ошибка конфликтом существующего URL
		var conflictErr *service.ErrURLConflict
		if errors.As(err, &conflictErr) {
			resp := models.Response{Result: conflictErr.ShortURL}
			h.sendJSONResponse(c, http.StatusConflict, resp)
			return
		}
		logger.Log.Error("Failed to get URL", zap.Error(err))
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	resp := models.Response{Result: result}
	h.sendJSONResponse(c, http.StatusCreated, resp)
}

// Ping хендлер проверяет соединение с базой данных.
// Возвращает HTTP 200 OK при успешной проверке, 500 Internal Server Error при неуспешной.
func (h *Handler) Ping(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.Service.Ping(ctx); err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		c.String(http.StatusInternalServerError, "Database connection failed")
		return
	}

	c.String(http.StatusOK, "OK")
}

// CreateURLBatch хендлер обрабатывает POST-запросы,
// принимает массив объектов с correlation_id и original_url,
// создает короткие URL для всех URL и возвращает массив объектов с correlation_id и short_url.
func (h *Handler) CreateURLBatch(c *gin.Context) {
	var reqItems []models.BatchRequestItem

	if err := json.NewDecoder(c.Request.Body).Decode(&reqItems); err != nil {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	if len(reqItems) == 0 {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	for _, item := range reqItems {
		if !utils.IsValidURL(item.OriginalURL) {
			c.String(http.StatusBadRequest, "invalid request")
			return
		}
	}

	ctx := c.Request.Context()
	responseItems, err := h.Service.CreateShortURLBatch(ctx, reqItems)
	if err != nil {
		logger.Log.Error("Failed to create batch URLs", zap.Error(err))
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	h.sendJSONResponse(c, http.StatusCreated, responseItems)
}
