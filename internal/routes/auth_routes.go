package routes

import (
	"database/sql"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/auth"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(api *gin.RouterGroup, db *sql.DB, queries *importDB.Queries) {

	loginController := controllers.LoginControllerHandler(queries)
	registerController := controllers.RegisterControllerHandler(queries)
	authHandler := auth.NewGoogleAuthHandler(queries)

	// Google OAuth
	api.GET("/auth/google/login", authHandler.GoogleLogin)
	api.GET("/auth/google/callback", authHandler.GoogleCallback)

	// Register + Login
	api.POST("/register", registerController.Register)
	api.POST("/login", loginController.Login)
}
