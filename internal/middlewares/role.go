package middlewares

import (
	"net/http"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

func RequireRole(queries *db.Queries, role string) gin.HandlerFunc {

	return func(c *gin.Context) {

		claims, exists := c.Get("claims")
		if !exists {
			utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort()
			return
		}

		userClaims, ok := claims.(*models.Claims)
		if !ok {
			utils.JSONIndent(c, http.StatusUnauthorized, "Invalid claims", nil)
			c.Abort()
			return
		}

		email := userClaims.Email

		switch strings.ToLower(role) {
		case "admin":
			admin, err := queries.GetUserByIDOrEmail(c, db.GetUserByIDOrEmailParams{
				ID:    int32(userClaims.ID),
				Email: email,
			})
			if err != nil || admin.Email == "" {
				utils.JSONIndent(c, http.StatusForbidden, "Admins only", nil)
				c.Abort()
				return
			}
		case "user":
			user, err := queries.GetUserByIDOrEmail(c, db.GetUserByIDOrEmailParams{
				ID:    int32(userClaims.ID),
				Email: email,
			})
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
