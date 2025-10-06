package server

import (
	"fmt"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/routes"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-contrib/cors"
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

	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// // Add CORS middleware
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:3000", "https://yourdomain.com"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// }))
	router.Use(cors.Default()) // Allows all origins

	router.Use(logging.GinLogger())
	router.Use(middlewares.RequestLogger()) // Add your custom middleware here

	// Setup routes
	routes.SetupRoutes(router, db)

	// Handle unknown paths with JSON 404
	router.NoRoute(func(c *gin.Context) {
		utils.JSONIndent(c, http.StatusNotFound, "404 Not Found", nil)
	})

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
