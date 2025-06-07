package app

import (
	"net/http"

	httphandlers "github.com/MaxRadzey/shortener/internal/handler"
	dbstorage "github.com/MaxRadzey/shortener/internal/storage"
)

// Run запускает http сервер.
func Run() error {
	storage := dbstorage.NewStorage()
	handler := &httphandlers.Handler{Storage: storage}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.CreateUrl)
	mux.HandleFunc("/{short_path}", handler.GetUrl)

	return http.ListenAndServe(":8080", mux)
}
