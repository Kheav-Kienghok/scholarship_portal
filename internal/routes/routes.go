package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/auth"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/Kheav-Kienghok/scholarship_portal/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
)

// SetupRoutes configures all the application routes
func SetupRoutes(router *gin.Engine, db *database.Database) {

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204) // No Content
	})

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5500"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	dbConn := db.DB // db is *database.Database, db.DB is *sql.DB
	queries := importDB.New(dbConn)

	homeController := controllers.NewHomeController()
	loginController := controllers.LoginControllerHandler(queries)
	registerController := controllers.RegisterControllerHandler(queries)
	userController := controllers.UserControllerHandler(dbConn, queries)

	auth := auth.NewGoogleAuthHandler(queries)

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/", homeController.GetHome)

		// api.GET("/auth/google/url", auth.GetLoginURL)

		// Public auth endpoints
		api.GET("/auth/google/login", auth.GoogleLogin)
		api.GET("/auth/google/callback", auth.GoogleCallback)

		api.POST("/register", registerController.Register)
		api.POST("/login", loginController.Login)

		// Protected endpoints (example: everything below requires auth)
		protected := api.Group("")
		protected.Use(middlewares.JWTAuthSingle("student"))
		{
			protected.GET("/profile", userController.GetUserProfile)

			protected.PATCH("/update-profile", userController.UpdateUserAndProfile)
		}
	}

}
