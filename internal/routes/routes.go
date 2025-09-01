package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/auth"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the application routes
func SetupRoutes(router *gin.Engine, db *database.Database) {

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204) // No Content
	})

	// Initialize controllers
	homeController := controllers.NewHomeController()
	loginController := controllers.LoginControllerHandler(db)
	registerController := controllers.RegisterControllerHandler(db)

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/", homeController.GetHome)

		api.GET("/auth/google/login", auth.GoogleLogin)
		api.GET("/auth/google/callback", auth.GoogleCallback)

		api.POST("/register", registerController.Register)
		api.POST("/login", loginController.Login)

		api.POST("/update-password", middlewares.JWTAuthMultiple("student", "admin"), loginController.UpdatePassword)
	}

}
