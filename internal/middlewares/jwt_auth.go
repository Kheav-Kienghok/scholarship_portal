package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

const ClaimsKey = "claims"

func JWTAuth(allowedRolesOrPurposes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for OPTIONS (preflight) requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Try lowercase
			authHeader = c.GetHeader("authorization")
		}

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") && !strings.HasPrefix(authHeader, "bearer ") {
			utils.JSONIndent(c, http.StatusUnauthorized, "Missing or invalid token", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "bearer "))

		var claims tokens.ClaimsInterface

		// Try all token types in order of preference
		if userClaims, err := tokens.ParseToken(tokenStr); err == nil {
			claims = userClaims
		} else if setupClaims, err := tokens.ParseSetupToken(tokenStr); err == nil {
			claims = setupClaims
		} else if serviceClaims, err := tokens.ParseTempToken(tokenStr); err == nil {
			claims = serviceClaims
		} else {
			go logging.Error("Failed to parse token: ", err)
			utils.JSONIndent(c, http.StatusUnauthorized, "Invalid token", nil)
			c.Abort()
			return
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
	return JWTAuth("admin", "setup", "admin_2fa")
}

func RequireUserAuth() gin.HandlerFunc {
	return JWTAuth("student")
}
