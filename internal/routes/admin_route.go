package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(rg *gin.RouterGroup, queries *db.Queries) {
	adminController := controllers.AdminControllerHandler(queries)

	admin := rg.Group("/admin")
	{
		admin.POST("/login", adminController.AdminLogin)

		// Require JWT for 2FA setup and verification
		adminAuth := admin.Group("/")
		adminAuth.Use(middlewares.RequireAdminAuth())
		{
			adminAuth.POST("/enable-2fa", adminController.Enable2FAForAdmin)
			adminAuth.POST("/verify-2fa", adminController.Verify2FAForAdmin)
		}
	}
}
