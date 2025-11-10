package cache

import (
    "time"

    "github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
)

// InitCache initializes the cache and starts cleanup routine
func InitCache() {
    logging.Info("Initializing in-memory cache...")
    
    // Start cleanup routine that runs every hour
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()

        for range ticker.C {
            URLCache.CleanExpired()
        }
    }()

    logging.Info("Cache initialized successfully")
}