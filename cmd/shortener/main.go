package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// UrlDB представляет собой маппу для хранения пар: сокращенная ссылка - длинная ссылка.
var UrlDB map[string]string

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// run запускает http сервер.
func run() error {
	UrlDB = make(map[string]string)
	mux := http.NewServeMux()

	mux.HandleFunc("/", createUrl)
	mux.HandleFunc("/{shortPath}", getUrl)

	return http.ListenAndServe(":8080", mux)
}

// createUrl хэндлер, обрабатывает POST-запросы, принимает текстовый URL в теле запроса,
// создает короткий путь и возвращает ег ов виде строки с полным URL.
// Ожидается Content-Type: text/plain
func createUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Allowed only POST method!", http.StatusMethodNotAllowed)
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
	shortPath := getShortPath(text)

	UrlDB[shortPath] = text

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	result := fmt.Sprintf("http://%s/%s", req.Host, shortPath)
	_, _ = res.Write([]byte(result))
}

// getUrl хэнедлер, обрабатывает GET-запросы, получает в качестве параметра маршрута сокращенное значение URL,
// ищет в БД совпадение длинного пути и производит редирект на него (307), иначе отдает (404) ошибку.
func getUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortPath := req.PathValue("shortPath")
	longUrl := UrlDB[shortPath]

	if longUrl == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", longUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// getShortPath возвращает короткое уникальное строковое представление пути (URL),
// который был передан. Использует алгоритм шифрования sha1 и кодирование base64.
// Результат обрезается до 6 символов.
func getShortPath(path string) string {
	hasher := sha1.New()
	hasher.Write([]byte(path))
	hash := hasher.Sum(nil)

	short := base64.URLEncoding.EncodeToString(hash)[:6]
	return short
}
