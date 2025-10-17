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
            logging.Info("Running cache cleanup...")
            URLCache.CleanExpired()
            logging.Info("Cache cleanup completed. Current size:", URLCache.Size())
        }
    }()

    logging.Info("Cache initialized successfully")
}