package routes

import (
	"database/sql"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/otpstore"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(api *gin.RouterGroup, db *sql.DB, queries *importDB.Queries, limiter *middlewares.RateLimiter) {
	// OTP store: max 5 attempts, 15 minutes window
	otpStore := otpstore.NewOTPStore(3, 15*time.Minute)

	adminController := controllers.AdminControllerHandler(db, queries, otpStore)

	admin := api.Group("/admin")
	{
		admin.POST("/login", limiter.Middleware(), adminController.AdminLogin)
		admin.POST("/verify-otp", limiter.Middleware(), adminController.VerifyAdminOTP)

		// Require JWT for 2FA setup and verification
		adminAuth := admin.Group("/")
		adminAuth.Use(middlewares.RequireAdminAuth())
		{
			adminAuth.POST("/enable-2fa", adminController.Enable2FAForAdmin)
			adminAuth.POST("/verify-2fa", adminController.Verify2FAForAdmin)
		}
	}
}
