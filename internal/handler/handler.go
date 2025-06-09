package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
	"github.com/MaxRadzey/shortener/internal/utils"
)

type Handler struct {
	Storage dbstorage.UrlStorage
}

// CreateUrl хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func (h *Handler) CreateUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Allowed only POST method!", http.StatusMethodNotAllowed)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(res, "Invalid Content-Type!", http.StatusBadRequest)
		return
	}

	if err := req.ParseForm(); err != nil {
		_, _ = res.Write([]byte(err.Error()))
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error occurred while reading body", http.StatusInternalServerError)
		return
	}

	text := string(body)
	if !utils.IsValidURL(text) {
		http.Error(res, "Invalid Body!", http.StatusBadRequest)
		return
	}

	shortPath, err := utils.GetShortPath(text)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	h.Storage.Create(shortPath, text)

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	result := fmt.Sprintf("http://%s/%s", req.Host, shortPath)
	_, _ = res.Write([]byte(result))
}

// GetUrl хэндлер, обрабатывает GET-запросы, получает в качестве параметра маршрута сокращенное значение URL,
// ищет в БД совпадение длинного пути и производит редирект на него (307), иначе отдает (404) ошибку.
func (h *Handler) GetUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortPath := req.PathValue("short_path")
	longUrl := h.Storage.Get(shortPath)

	if longUrl == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", longUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
