package routes

import (
	"database/sql"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(api *gin.RouterGroup, db *sql.DB, queries *importDB.Queries) {
	adminController := controllers.AdminControllerHandler(db, queries)

	admin := api.Group("/admin")
	{
		admin.POST("/login", adminController.AdminLogin)
		admin.POST("/verify-otp", adminController.VerifyAdminOTP)

		// Require JWT for 2FA setup and verification
		adminAuth := admin.Group("/")
		adminAuth.Use(middlewares.RequireAdminAuth())
		{
			adminAuth.POST("/enable-2fa", adminController.Enable2FAForAdmin)
			adminAuth.POST("/verify-2fa", adminController.Verify2FAForAdmin)
		}
	}
}
