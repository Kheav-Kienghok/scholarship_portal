package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

// Database wraps the sql.DB and connection string
type Database struct {
	ConnString string
	DB         *sql.DB
	Queries    *importDB.Queries
}

// NewDatabase creates a new database instance
func NewDatabase(connString string) *Database {
	return &Database{ConnString: connString}
}

// Connect establishes a connection to the database with robust connection pooling
func (d *Database) Connect() error {
	db, err := sql.Open("pgx", d.ConnString)
	if err != nil {
		logging.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		return fmt.Errorf("database connection failed") // Sanitized error
	}

	// Configure connection pool for robustness
	d.setupConnectionPool(db)

	// Test the connection with retry logic
	if err := d.pingWithRetry(db, 3, 2*time.Second); err != nil {
		db.Close()
		logging.Error(fmt.Sprintf("Failed to ping database after retries: %v", err))
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.DB = db
	d.Queries = importDB.New(db)

	logging.Info("Database connection established")
	return nil
}

// setupConnectionPool configures connection pool settings
func (d *Database) setupConnectionPool(db *sql.DB) {

	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(5)
	// db.SetConnMaxLifetime(5 * time.Minute)
	// db.SetConnMaxIdleTime(1 * time.Minute)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	go logging.Info("Connection pool configured: MaxOpen=25, MaxIdle=5, MaxLifetime=5m, MaxIdleTime=1m")
}

// pingWithRetry attempts to ping the database with retry logic
func (d *Database) pingWithRetry(db *sql.DB, retries int, delay time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			return nil
		}

		logging.Warn(fmt.Sprintf("Database ping attempt %d failed: %v", i+1, err))
		if i < retries-1 {
			time.Sleep(delay * time.Duration(i+1)) // Exponential backoff
		}
	}
	return err
}

// HealthCheck performs a database health check
func (d *Database) HealthCheck() error {
	if d.DB == nil {
		return fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := d.DB.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		logging.Error(fmt.Sprintf("Database health check failed: %v", err))
		d.logConnectionStats()
		return err
	}

	logging.Info(fmt.Sprintf("Database health check passed (%.2fms)", float64(duration.Nanoseconds())/1e6))
	return nil
}

// logConnectionStats logs current connection pool statistics
func (d *Database) logConnectionStats() {
	if d.DB == nil {
		return
	}

	stats := d.DB.Stats()
	logging.Info(fmt.Sprintf("DB Stats - Open: %d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
	))
}

// StartHealthMonitoring starts periodic health checks
func (d *Database) StartHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logging.Info("Database health monitoring stopped")
			return
		case <-ticker.C:
			d.HealthCheck()
		}
	}
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		err := d.DB.Close()
		if err != nil {
			logging.Error(fmt.Sprintf("Error closing database: %v", err))
			return err
		}
		logging.Info("Database connection closed")
	}
	return nil
}

// Migrate runs database migrations using Goose
func (d *Database) Migrate(migrationsDir string) error {
	if d.DB == nil {
		return fmt.Errorf("database not connected")
	}

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		logging.Warn(fmt.Sprintf("Migrations directory not found: %s. Skipping migrations.", migrationsDir))
		return nil // just skip, do not error
	}

	goose.SetDialect("postgres")
	err := goose.Up(d.DB, migrationsDir)
	if err != nil {
		logging.Error(fmt.Sprintf("Migration failed: %v", err))
		return err
	}
	logging.Info("Database migration completed successfully")
	return nil
}
