package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuthMultiple checks for a valid JWT token and one of the allowed roles
func JWTAuthMultiple(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.JSONIndent(c, http.StatusUnauthorized, "Missing or invalid token", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.ParseToken(tokenStr)
		if err != nil {
			utils.JSONIndent(c, http.StatusUnauthorized, "Invalid token", nil)
			c.Abort()
			return
		}

		if len(allowedRoles) > 0 {
			roleAllowed := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				utils.JSONIndent(c, http.StatusForbidden, "Forbidden: insufficient privileges", nil)
				c.Abort()
				return
			}
		}

		// Store claims in context for downstream handlers
		c.Set("claims", claims)
		c.Next()
	}
}

// JWTAuthSingle is just a wrapper around JWTAuthMultiple
func JWTAuthSingle(requiredRole string) gin.HandlerFunc {
	return JWTAuthMultiple(requiredRole)
}
