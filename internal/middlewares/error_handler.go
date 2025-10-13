package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/errors"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware to catch panics and handle them gracefully
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			errors.SanitizedErrorResponse(c,
				fmt.Errorf("panic: %s", err),
				http.StatusInternalServerError,
				"An unexpected error occurred")
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
