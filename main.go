package main

import (
	"log"
	"os"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/cache"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/server"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Check if .env exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		gin.SetMode(gin.ReleaseMode)
	} else {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

// @title EduVision Scholarship Portal API
// @version 1.0
// @description This API allows managing student profiles, authentication, and scholarship applications for the EduVision portal.
// @termsOfService https://eduvsion.example.com/terms
// @contact.name EduVision Support
// @contact.url https://eduvsion.example.com/support
// @contact.email support@eduvsion.example.com
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Initialize logger
	logging.InitAsyncLogger("logs/app.log")

	// Initialize cache
	cache.InitCache()

	// Get DB connection string
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	// Connect to DB
	db := database.NewDatabase(connStr)
	if err := db.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := db.Migrate("migrations"); err != nil {

		if strings.Contains(err.Error(), "does not exist") {
			log.Println("Migrations folder not found, skipping migrations")
		} else {
			log.Fatal("Failed to run migrations:", err)
		}
	}

	storage.InitS3()

	// Get server port from .env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server (Gin engine created here)
	srv := server.NewServer(port, db)
	if err := srv.Start(); err != nil {
		logging.LogError(err, "server startup")
		log.Fatal("Failed to start server:", err)
	}

	defer db.Close()
}
