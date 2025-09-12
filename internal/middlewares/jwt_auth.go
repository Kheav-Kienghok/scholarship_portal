package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

const ClaimsKey = "claims"

func JWTAuth(allowedRolesOrPurposes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.JSONIndent(c, http.StatusUnauthorized, "Missing or invalid token", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		var claims tokens.ClaimsInterface
		var err error

		userClaims, err := tokens.ParseToken(tokenStr)
		if err == nil {
			claims = userClaims
		} else {
			// Fallback to Setup token
			setupClaims, err := tokens.ParseSetupToken(tokenStr)
			if err != nil {
				utils.JSONIndent(c, http.StatusUnauthorized, "Invalid token", nil)
				c.Abort()
				return
			}
			claims = setupClaims
		}

		if len(allowedRolesOrPurposes) > 0 {
			allowed := false
			for _, val := range allowedRolesOrPurposes {
				if claims.GetRole() == val || claims.GetPurpose() == val {
					allowed = true
					break
				}
			}
			if !allowed {
				utils.JSONIndent(c, http.StatusForbidden, "Forbidden", nil)
				c.Abort()
				return
			}
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

func RequireAdminAuth() gin.HandlerFunc {
	return JWTAuth("admin", "setup")
}

func RequireUserAuth() gin.HandlerFunc {
	return JWTAuth("student")
}
