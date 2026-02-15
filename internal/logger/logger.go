package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Log — глобальный логгер; инициализируется через Initialize.
var Log *zap.Logger = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		gin.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Initialize настраивает глобальный логгер по уровню (info, debug и т.д.).
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

// HTTPLogger — middleware, логирующий метод, путь, статус, размер и время запроса.
func HTTPLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rd := &responseData{status: 0, size: 0}
		c.Writer = &loggingResponseWriter{ResponseWriter: c.Writer, responseData: rd}
		c.Next()
		Log.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
