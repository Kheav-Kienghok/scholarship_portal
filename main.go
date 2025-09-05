package main

import (
	"log"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/server"
	"github.com/joho/godotenv"
)


// @title EduVision for Scholarship Portal API
// @version 1.0
// @description API documentation for EduVision Scholarship Portal
// @host localhost:8080
// @BasePath /api/v1
func main() {

	// Initialize the logging system
	if err := logging.InitLogger(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get database connection string (use env or hardcoded for demo)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
	db := database.NewDatabase(connStr)
	if err := db.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := db.Migrate("migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create a new server instance on port 8080
	srv := server.NewServer("8080", db)

	// Start the server
	if err := srv.Start(); err != nil {
		logging.LogError(err, "server startup")
		log.Fatal("Failed to start server:", err)
	}

	// Close DB on exit (optional, for graceful shutdown)
	defer db.Close()
}
