package middlewares

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// responseWriter is a custom response writer that captures the response body
type etagResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *etagResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ETagMiddleware generates and validates ETags for responses
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply ETag for GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Create a custom response writer to capture the response body
		writer := &etagResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		if c.Writer.Status() == http.StatusOK {
			// Generate ETag from response body
			hash := md5.New()
			io.Copy(hash, writer.body)
			etag := `W/"` + hex.EncodeToString(hash.Sum(nil)) + `"`  // Weak ETag
			// etag := `"` + hex.EncodeToString(hash.Sum(nil)) + `"`

			// Check if client sent If-None-Match header
			clientETag := c.GetHeader("If-None-Match")

			if clientETag == etag {
				// Content hasn't changed, return 304 Not Modified
				c.Status(http.StatusNotModified)
				c.Abort()
				return
			}

			// 👇 Weak ETag + safe Cache-Control (Cloudflare won’t override)
			c.Header("ETag", etag)
			c.Header("Cache-Control", "no-cache, must-revalidate")
		}
	}
}
