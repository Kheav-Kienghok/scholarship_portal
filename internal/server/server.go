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

type fakeReminderStore struct{}

func (f *fakeReminderStore) GetPendingReminders(ctx context.Context) ([]utils.ReminderRequest, error) {
	return []utils.ReminderRequest{
		{
			Name:            "Test Student 1",
			Email:           "kheavkienghok@gmail.com",
			ScholarshipName: "Tech Excellence Award",
			Deadline:        time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			ApplyLink:       "https://test-university.edu/tech",
		},
		{
			Name:            "Test Student 2",
			Email:           "khievkeanghok@gmail.com",
			ScholarshipName: "Science Research Grant",
			Deadline:        time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			ApplyLink:       "https://test-university.edu/science",
		},
	}, nil
}

func (f *fakeReminderStore) MarkReminderSent(ctx context.Context, id int64) error {
	// No-op for fake
	return nil
}

// NewServer creates a new server instance
func NewServer(port string, db *database.Database) *Server {

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery(), logging.GinLogger(), middlewares.RequestLogger())

	// Enable gzip compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

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
	// store cancel to allow stopping the reminder goroutine when server shuts down
	// reminderStore := db.Queries
	reminderStore := &fakeReminderStore{}              // use the fake store for testing
	utils.StartDailyEmailCheck(ctx, reminderStore, "") // "" for every minute (testing)
	// --- END REMINDER CRON ---
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
