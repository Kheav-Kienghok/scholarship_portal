package routes

import (
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the application routes
func SetupRoutes(router *gin.Engine, db *database.Database) {
    // Add Recovery middleware first (catches panics and recovers)
    router.Use(gin.Recovery())

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	// Serve static files from public directory
	router.Static("/public", "./public")

	// CORS - Allow all origins
	router.Use(cors.Default())

	// // CORS
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:5500"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// }))

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Build queries once
	queries := importDB.New(db.DB)

	// Mount APIs
	api := router.Group("/api/v1")
	{
		RegisterHomeRoutes(api)
		RegisterAuthRoutes(api, db.DB, queries)
		RegisterUserRoutes(api, db.DB, queries)

		RegisterAdminRoutes(api, db.DB, queries)

		RegisterScholarshipRoutes(api, queries)
	}
}