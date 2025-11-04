package middlewares

import (
    "fmt"

    "github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
    "github.com/gin-gonic/gin"
)

// RequestLogger logs the request method and path
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        go logging.Info(fmt.Sprintf("Middleware: %s %s", c.Request.Method, c.Request.URL.Path))
        c.Next()
    }
}