package gzip

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")

		if supportGzip {
			gz := gzip.NewWriter(c.Writer)
			defer gz.Close()
			c.Writer = &compressWriter{Writer: gz, ResponseWriter: c.Writer}
			c.Header("Content-Encoding", "gzip")
		}

		contentEncoding := c.GetHeader("Content-Encoding")
		sendGzip := strings.Contains(contentEncoding, "gzip")
		if sendGzip {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			defer reader.Close()
			c.Request.Body = reader
		}
		c.Next()
	}
}
