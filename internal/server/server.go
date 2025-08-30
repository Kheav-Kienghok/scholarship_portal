package server

import (
	"fmt"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/routes"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	port   string
	db     *database.Database
}

// NewServer creates a new server instance
func NewServer(port string, db *database.Database) *Server {

	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(logging.GinLogger())

	// Setup routes
	routes.SetupRoutes(router, db)

	return &Server{
		router: router,
		port:   port,
		db:     db,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logging.Info(fmt.Sprintf("Server starting on port %s", s.port))
	return s.router.Run(":" + s.port)
}
