package routes

import (
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/auth"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/controllers"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"

	// "github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
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
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	dbConn := db.DB // db is *database.Database, db.DB is *sql.DB
	queries := importDB.New(dbConn)

	homeController := controllers.NewHomeController()
	loginController := controllers.LoginControllerHandler(queries)
	registerController := controllers.RegisterControllerHandler(queries)

	// In your routes setup:
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/", homeController.GetHome)

		api.GET("/auth/google/login", auth.GoogleLogin)
		api.GET("/auth/google/callback", auth.GoogleCallback)
		api.GET("/auth/google/url", auth.GetLoginURL)

		api.POST("/register", registerController.Register)
		api.POST("/login", loginController.Login)

		// api.POST("/update-password", middlewares.JWTAuthMultiple("student", "admin"), loginController.UpdatePassword)
	}

}
