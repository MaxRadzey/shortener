// Package handler содержит HTTP-хендлеры эндпоинтов коротких ссылок.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MaxRadzey/shortener/internal/audit"
	"github.com/MaxRadzey/shortener/internal/contextkeys"
	"github.com/MaxRadzey/shortener/internal/logger"
	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/service"
	"github.com/MaxRadzey/shortener/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler обрабатывает HTTP-запросы к эндпоинтам коротких ссылок.
type Handler struct {
	Service       *service.Service
	Audit         *audit.Notifier
	TrustedSubnet string
}

type errResp struct {
	Error string `json:"error"`
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

// sendErrorJSON отправляет JSON ответ с ошибкой в формате {"error": "..."}
func (h *Handler) sendErrorJSON(c *gin.Context, statusCode int, errorMsg string) {
	c.JSON(statusCode, errResp{Error: errorMsg})
}

// sendJSONResponse отправляет JSON ответ и обрабатывает ошибки кодирования
func (h *Handler) sendJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(statusCode)
	if err := json.NewEncoder(c.Writer).Encode(data); err != nil {
		logger.Log.Error("Failed to encode JSON response", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
	}
}

// decodeJSONBody декодирует JSON тело запроса в указанную структуру.
// При ошибке декодирования возвращает false и отправляет ответ с HTTP 400 Bad Request.
func (h *Handler) decodeJSONBody(c *gin.Context, target interface{}) bool {
	if err := json.NewDecoder(c.Request.Body).Decode(target); err != nil {
		h.sendErrorJSON(c, http.StatusBadRequest, "invalid request")
		return false
	}
	return true
}

// CreateURL — POST /: принимает URL в теле (text/plain), возвращает короткую ссылку (201) или конфликт (409).
func (h *Handler) CreateURL(c *gin.Context) {
	userID, ok := h.userIDFromContext(c)
	if !ok {
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.sendErrorJSON(c, http.StatusBadRequest, "Invalid Body")
		return
	}

	text := string(body)

	if !utils.IsValidURL(text) {
		h.sendErrorJSON(c, http.StatusBadRequest, "Invalid Body")
		return
	}

	result, err := h.Service.CreateShortURL(text, userID)
	if err != nil {
		var conflictErr *service.ErrURLConflict
		if errors.As(err, &conflictErr) {
			if h.Audit != nil {
				h.Audit.Notify(audit.NewEvent("shorten", userID, text))
			}
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusConflict, conflictErr.ShortURL)
			return
		}
		logger.Log.Error("Failed to create URL", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	if h.Audit != nil {
		h.Audit.Notify(audit.NewEvent("shorten", userID, text))
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusCreated, result)
}

// GetURL — GET /:short_path: редирект на оригинальный URL (307), иначе 404 или 410 если удалён.
func (h *Handler) GetURL(c *gin.Context) {
	shortPath := c.Param("short_path")
	longURL, err := h.Service.GetLongURL(shortPath)

	if err != nil {
		var goneErr *service.ErrGone
		if errors.As(err, &goneErr) {
			h.sendErrorJSON(c, http.StatusGone, "Gone")
			return
		}
		h.sendErrorJSON(c, http.StatusNotFound, "Not found")
		return
	}

	userID, _ := h.userIDFromContext(c)
	if h.Audit != nil {
		h.Audit.Notify(audit.NewEvent("follow", userID, longURL))
	}
	c.Redirect(http.StatusTemporaryRedirect, longURL)
}

// GetURLJSON — POST /api/shorten: тело JSON {"url":"..."}, в ответе {"result":"короткая ссылка"}.
func (h *Handler) GetURLJSON(c *gin.Context) {
	userID, ok := h.userIDFromContext(c)
	if !ok {
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	var req models.Request
	if !h.decodeJSONBody(c, &req) {
		return
	}

	if !utils.IsValidURL(req.URL) {
		h.sendErrorJSON(c, http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.Service.CreateShortURL(req.URL, userID)
	if err != nil {
		var conflictErr *service.ErrURLConflict
		if errors.As(err, &conflictErr) {
			if h.Audit != nil {
				h.Audit.Notify(audit.NewEvent("shorten", userID, req.URL))
			}
			resp := models.Response{Result: conflictErr.ShortURL}
			h.sendJSONResponse(c, http.StatusConflict, resp)
			return
		}
		logger.Log.Error("Failed to get URL", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	if h.Audit != nil {
		h.Audit.Notify(audit.NewEvent("shorten", userID, req.URL))
	}
	resp := models.Response{Result: result}
	h.sendJSONResponse(c, http.StatusCreated, resp)
}

// Ping — GET /ping: проверка доступности хранилища, 200 или 500.
func (h *Handler) Ping(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.Service.Ping(ctx); err != nil {
		logger.Log.Error("Failed to ping database", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Database connection failed")
		return
	}

	c.String(http.StatusOK, "OK")
}

// CreateURLBatch — POST /api/shorten/batch: массив {correlation_id, original_url}, ответ — массив {correlation_id, short_url}.
func (h *Handler) CreateURLBatch(c *gin.Context) {
	var reqItems []models.BatchRequestItem
	if !h.decodeJSONBody(c, &reqItems) {
		return
	}

	if len(reqItems) == 0 {
		h.sendErrorJSON(c, http.StatusBadRequest, "invalid request")
		return
	}

	for _, item := range reqItems {
		if !utils.IsValidURL(item.OriginalURL) {
			h.sendErrorJSON(c, http.StatusBadRequest, "invalid request")
			return
		}
	}

	userID, ok := h.userIDFromContext(c)
	if !ok {
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	ctx := c.Request.Context()
	responseItems, err := h.Service.CreateShortURLBatch(ctx, reqItems, userID)
	if err != nil {
		logger.Log.Error("Failed to create batch URLs", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.sendJSONResponse(c, http.StatusCreated, responseItems)
}

// GetUserURLs — GET /api/user/urls: список коротких ссылок пользователя (200) или 204 если пусто.
func (h *Handler) GetUserURLs(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	items, err := h.Service.GetUserURLs(ctx, userID)
	if err != nil {
		logger.Log.Error("Failed to get user URLs", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(items) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	h.sendJSONResponse(c, http.StatusOK, items)
}

// DeleteURLs — DELETE /api/user/urls: тело — JSON-массив short URL; удаление асинхронное, ответ 202.
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
		h.sendErrorJSON(c, http.StatusBadRequest, "invalid request")
		return
	}

	go func() {
		if err := h.Service.DeleteURLs(context.Background(), userID, shortUrls); err != nil {
			logger.Log.Error("Failed to delete URLs", zap.Error(err))
		}
	}()

	c.Status(http.StatusAccepted)
}

// GetInternalStats — GET /api/internal/stats: возвращает количество URL и пользователей.
func (h *Handler) GetInternalStats(c *gin.Context) {
	clientIP := c.GetHeader("X-Real-IP")

	if !utils.IsIPInTrustedSubnet(clientIP, h.TrustedSubnet) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	ctx := c.Request.Context()
	urls, users, err := h.Service.Stats(ctx)
	if err != nil {
		logger.Log.Error("Failed to get stats", zap.Error(err))
		h.sendErrorJSON(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"urls": urls, "users": users})
}
