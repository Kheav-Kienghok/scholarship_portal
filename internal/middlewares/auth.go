package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/gin-gonic/gin"
)

func RequireAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.ParseSetupToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			return
		}

		// Allow if full admin OR setup token
		if claims.Role != "admin" && claims.Purpose != "setup" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		c.Next()
	}
}
