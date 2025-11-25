package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/middlewares"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/routes"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	port   string
	db     *database.Database
	cancel context.CancelFunc
}

// NewServer creates a new server instance
func NewServer(port string, db *database.Database) *Server {

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery(), logging.GinLogger(), middlewares.RequestLogger())

	// allow up to 100 MB multipart form in memory before using temp files
	router.MaxMultipartMemory = 100 << 20 // 100 MB

	// Enable gzip compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 100<<20)
		c.Next()
	})

	// Initialize validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = utils.InitValidator(v)
	}

	// Setup routes (CORS is configured in routes.go)
	routes.SetupRoutes(router, db)

	// Handle unknown paths with JSON 404
	router.NoRoute(func(c *gin.Context) {
		utils.JSONIndent(c, http.StatusNotFound, "404 Not Found", nil)
	})

	// --- START REMINDER CRON HERE ---
	ctx, cancel := context.WithCancel(context.Background())
	reminderStore := database.NewReminderStore(db) // pass the Database wrapper
	// utils.StartDailyEmailCheck(ctx, reminderStore, "") // testing every minute
	utils.StartDailyEmailCheck(ctx, reminderStore, "0 8 * * *") // at 8:00:00 every day
	// --- END REMINDER CRON ---

	return &Server{
		router: router,
		port:   port,
		db:     db,
		cancel: cancel,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	logging.Info(fmt.Sprintf("Server starting on port %s", s.port))
	return srv.ListenAndServe()
}
