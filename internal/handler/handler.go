package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/service"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Service *service.Service
}

// userIDFromContext возвращает user_id из контекста. ok == false, если нет или пусто.
func (h *Handler) userIDFromContext(c *gin.Context) (userID string, ok bool) {
	v, _ := c.Get(contextkeys.UserIDKey)
	userID, _ = v.(string)
	return userID, userID != ""
}

// requireUserID получает user_id из контекста и проверяет его наличие.
// При отсутствии user_id отправляет HTTP 401 Unauthorized и возвращает false.
// При успехе возвращает userID и true.
func (h *Handler) requireUserID(c *gin.Context) (userID string, ok bool) {
	userID, ok = h.userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return "", false
	}
	return userID, true
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

// decodeJSONBody декодирует JSON тело запроса в указанную структуру.
// При ошибке декодирования возвращает false и отправляет ответ с HTTP 400 Bad Request.
func (h *Handler) decodeJSONBody(c *gin.Context, target interface{}) bool {
	if err := json.NewDecoder(c.Request.Body).Decode(target); err != nil {
		c.String(http.StatusBadRequest, "invalid request")
		return false
	}
	return true
}

// CreateURL хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает его в виде строки с полным URL.
// Ожидается Content-Type: text/plain. user_id в контексте от auth.
func (h *Handler) CreateURL(c *gin.Context) {
	userID, ok := h.userIDFromContext(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Body!")
		return
	}

	text := string(body)

	if !utils.IsValidURL(text) {
		c.String(http.StatusBadRequest, "Invalid Body!")
		return
	}

	result, err := h.Service.CreateShortURL(text, userID)
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
// ищет в БД совпадение длинного пути и производит редирект на него (307).
// Для удалённого URL возвращает 410 Gone, для отсутствующего — 404 Not Found.
func (h *Handler) GetURL(c *gin.Context) {
	shortPath := c.Param("short_path")
	longURL, err := h.Service.GetLongURL(shortPath)

	if err != nil {
		var goneErr *dbstorage.ErrGone
		if errors.As(err, &goneErr) {
			c.String(http.StatusGone, "Gone")
			return
		}
		c.String(http.StatusNotFound, "Not found!")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (h *Handler) GetURLJSON(c *gin.Context) {
	userID, ok := h.userIDFromContext(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	var req models.Request
	if !h.decodeJSONBody(c, &req) {
		return
	}

	if !utils.IsValidURL(req.URL) {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.Service.CreateShortURL(req.URL, userID)
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
	if !h.decodeJSONBody(c, &reqItems) {
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

	userID, ok := h.userIDFromContext(c)
	if !ok {
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	ctx := c.Request.Context()
	responseItems, err := h.Service.CreateShortURLBatch(ctx, reqItems, userID)
	if err != nil {
		logger.Log.Error("Failed to create batch URLs", zap.Error(err))
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	h.sendJSONResponse(c, http.StatusCreated, responseItems)
}

// GetUserURLs возвращает все сокращённые пользователем URL.
func (h *Handler) GetUserURLs(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	items, err := h.Service.GetUserURLs(ctx, userID)
	if err != nil {
		logger.Log.Error("Failed to get user URLs", zap.Error(err))
		c.String(http.StatusInternalServerError, "Internal server error!")
		return
	}

	if len(items) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	h.sendJSONResponse(c, http.StatusOK, items)
}

// DeleteURLs обрабатывает DELETE-запрос для асинхронного удаления сокращённых URL.
// Принимает список идентификаторов сокращённых URL в теле запроса (JSON массив строк).
// Возвращает HTTP 202 Accepted при успешном приёме запроса.
func (h *Handler) DeleteURLs(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	var shortUrls []string
	if !h.decodeJSONBody(c, &shortUrls) {
		return
	}

	if len(shortUrls) == 0 {
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	ctx := c.Request.Context()

	// Запускаем асинхронное удаление в горутине
	go func() {
		if err := h.Service.DeleteURLs(ctx, userID, shortUrls); err != nil {
			logger.Log.Error("Failed to delete URLs", zap.Error(err))
			// Не оповещаем пользователя об ошибках, так как удаление асинхронное
		}
	}()

	c.Status(http.StatusAccepted)
}
