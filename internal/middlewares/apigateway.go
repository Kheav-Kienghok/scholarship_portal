package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

func APIGatewayMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		logging.Info(fmt.Sprintf("API Gateway: %s %s", c.Request.Method, c.Request.URL.Path))

		// Auth check (JWT)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.JSONIndent(c, http.StatusUnauthorized, "Missing or Invalid token", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.ParseToken(tokenStr)
		if err != nil {
			utils.JSONIndent(c, http.StatusUnauthorized, "Missing or Invalid token", nil)
			c.Abort()
			return
		}

		// Role-based access
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

		// Example: block certain paths or methods (customize as needed)
		if c.Request.URL.Path == "/api/v1/admin" && claims.Role != "admin" {
			utils.JSONIndent(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		// Store claims in context for downstream handlers
		c.Set("claims", claims)
		c.Next()
	}
}
