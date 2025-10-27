package routes

import (
	"database/sql"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/auth"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(api *gin.RouterGroup, db *sql.DB, queries *importDB.Queries, limiter *middlewares.RateLimiter) {

	loginController := controllers.LoginControllerHandler(queries)
	registerController := controllers.RegisterControllerHandler(db, queries)
	googleHandler := auth.NewGoogleAuthHandler(queries)

	// Create auth sub-group
	authGroup := api.Group("/auth")
	{
		// Google OAuth
		authGroup.GET("/google/login", googleHandler.GoogleLogin)
		authGroup.GET("/google/callback", googleHandler.GoogleCallback)

		// Email Verification
		authGroup.GET("/verify-email", registerController.VerifyEmail)
		authGroup.POST("/resend-verification", registerController.ResendVerification)
	}

	// Register + Login (directly under /api/v1)
	api.POST("/register", limiter.Middleware(), registerController.Register)
	api.POST("/login", limiter.Middleware(), loginController.Login)
}
