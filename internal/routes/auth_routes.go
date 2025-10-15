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
	googleHandler := auth.NewGoogleAuthHandler(queries)
	// facebookHandler := auth.NewFBAuthHandler(queries)

	// // Facebook OAuth
	// api.GET("/auth/facebook/login", facebookHandler.LoginHandler)
	// api.GET("/auth/facebook/callback", facebookHandler.CallbackHandler)

	// Google OAuth
	api.GET("/auth/google/login", googleHandler.GoogleLogin)
	api.GET("/auth/google/callback", googleHandler.GoogleCallback)

	// Register + Login
	api.POST("/register", registerController.Register)
	api.POST("/login", loginController.Login)
}
