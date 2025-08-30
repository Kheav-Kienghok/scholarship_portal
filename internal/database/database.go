package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

// Database wraps the sql.DB and connection string
type Database struct {
	ConnString string
	DB         *sql.DB
}

// NewDatabase creates a new database instance
func NewDatabase(connString string) *Database {
	return &Database{ConnString: connString}
}

// Connect establishes a connection to the database
func (d *Database) Connect() error {
	db, err := sql.Open("pgx", d.ConnString)
	if err != nil {
		logging.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		return err
	}
	d.DB = db
	logging.Info("Database connection established")
	return nil
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
		return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
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
