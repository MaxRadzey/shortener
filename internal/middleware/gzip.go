package middleware

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	gin.ResponseWriter
	Writer *gzip.Writer
}

func (c *compressWriter) Write(data []byte) (int, error) {
	return c.Writer.Write(data)
}

func (c *compressWriter) Close() error {
	return c.Writer.Close()
}

func (c *compressWriter) WriteString(s string) (int, error) {
	return c.Writer.Write([]byte(s))
}

// Gzip обрабатывает сжатие и распаковку gzip для HTTP запросов и ответов.
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		log.Printf("[GzipMiddleware] Accept-Encoding: %s", acceptEncoding)
		supportGzip := strings.Contains(acceptEncoding, "gzip")

		if supportGzip {
			gz := gzip.NewWriter(c.Writer)
			defer func() {
				if err := gz.Close(); err != nil {
					log.Printf("[GzipMiddleware] Error closing gzip writer: %v", err)
				}
			}()
			c.Writer = &compressWriter{Writer: gz, ResponseWriter: c.Writer}
			c.Header("Content-Encoding", "gzip")
		}

		contentEncoding := c.GetHeader("Content-Encoding")
		log.Printf("[GzipMiddleware] Content-Encoding: %s", contentEncoding)
		sendGzip := strings.Contains(contentEncoding, "gzip")

		if sendGzip {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			defer func() {
				if err := reader.Close(); err != nil {
					log.Printf("[GzipMiddleware] Error closing gzip reader: %v", err)
				}
			}()
			c.Request.Body = reader
		}
		c.Next()
		log.Printf("[GzipMiddleware] Finished processing request")
	}
}
