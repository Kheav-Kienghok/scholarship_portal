package routes

import (
	"net/http"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the application routes
func SetupRoutes(router *gin.Engine, db *database.Database) {

	// CORS MUST be the first middleware
	config := cors.Config{
		// AllowOrigins: []string{
		// 	"http://localhost:5173",
		// 	"https://eduvision.live",
		// 	"https://www.eduvision.live",
		// },
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "Content-Length"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))
	router.Use(middlewares.ErrorHandler())
	router.Use(middlewares.ETagMiddleware())

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Build queries once
	queries := importDB.New(db.DB)

	// Mount APIs
	api := router.Group("/api/v1")
	{
		RegisterHomeRoutes(api)
		RegisterUserRoutes(api, db.DB, queries)

		RegisterScholarshipRoutes(api, queries)

		// rateLimiter := middlewares.NewRateLimiter(10, 15, 10*time.Minute)
		rateLimiter := middlewares.NewRateLimiter(5, 10, 3*time.Minute)
		RegisterAuthRoutes(api, db.DB, queries, rateLimiter)
		RegisterAdminRoutes(api, db.DB, queries, rateLimiter)
	}
}
