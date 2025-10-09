package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

func RequireRole(queries *db.Queries, role string) gin.HandlerFunc {

	return func(c *gin.Context) {

		claims, exists := c.Get("claims")
		if !exists {
			logging.Error("[Middleware]: Claims not found in context")
			utils.JSONIndent(c, http.StatusUnauthorized, "Something went wrong!", nil)
			c.Abort()
			return
		}

		userClaims, ok := claims.(*tokens.UserClaims)
		if !ok {
			logging.Error("Invalid claims type")
			utils.JSONIndent(c, http.StatusUnauthorized, "Invalid claims", nil)
			c.Abort()
			return
		}

		email := userClaims.Email

		switch strings.ToLower(role) {
		case "admin":
			admin, err := queries.GetAdminByEmail(c, email)
			if err != nil || admin.Email == "" {
				utils.JSONIndent(c, http.StatusForbidden, "Admins only", nil)
				c.Abort()
				return
			}
		case "user":
			user, err := queries.GetUserByEmail(c, email)
			if err != nil || user.Email == "" {
				utils.JSONIndent(c, http.StatusForbidden, "Users only", nil)
				c.Abort()
				return
			}
		default:
			utils.JSONIndent(c, http.StatusForbidden, "Invalid role", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
