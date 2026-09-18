// Package database provides a shared PostgreSQL connection pool for all
// Store4bots microservices.  Use InitDB to open the pool, GetDB to retrieve
// the singleton, and CloseDB to shut it down gracefully.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DB is the global database connection pool.
var DB *sql.DB

// Config holds database connection settings.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns a Config populated from environment variables
// or sensible defaults for local development.
func DefaultConfig() Config {
	return Config{
		Port:            getEnvInt("DB_PORT", 5432),
		Host:            getEnv("DB_HOST", "postgres"),
		User:            getEnv("DB_USER", "app"),
		Password:        getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "appdb"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 15)) * time.Minute,
		ConnMaxIdleTime: time.Duration(getEnvInt("DB_CONN_MAX_IDLE_TIME_MIN", 5)) * time.Minute,
	}
}

// InitDB opens the database connection pool and pings the server.
// It is safe to call multiple times; subsequent calls return the existing pool.
func InitDB(cfg Config) (*sql.DB, error) {
	if DB != nil {
		return DB, nil
	}

	var connStr string
	if cfg.Host != "" {
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	} else {
		connStr = os.Getenv("DATABASE_URL")
		if connStr == "" {
			connStr = "postgres://app:password@postgres:5432/appdb?sslmode=disable"
		}
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure the connection pool
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	DB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := DB.PingContext(ctx); err != nil {
		DB = nil
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[database] Connected to PostgreSQL")
	return DB, nil
}

// InitDBFromURL is a convenience wrapper that uses DATABASE_URL.
func InitDBFromURL() (*sql.DB, error) {
	if DB != nil {
		return DB, nil
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://app:password@postgres:5432/appdb?sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(15 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := DB.PingContext(ctx); err != nil {
		DB = nil
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[database] Connected to PostgreSQL (via URL)")
	return DB, nil
}

// GetDB returns the global database connection pool.
// Returns nil if InitDB has not been called.
func GetDB() *sql.DB {
	return DB
}

// CloseDB closes the global database pool. Safe to call when DB is nil.
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("[database] error closing DB: %v", err)
		}
		log.Println("[database] Connection closed")
		DB = nil
	}
}

// Health checks if the database is reachable.
func Health() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return DB.PingContext(ctx)
}

// Stats returns pool statistics.
func Stats() sql.DBStats {
	if DB != nil {
		return DB.Stats()
	}
	return sql.DBStats{}
}

// =========================================================================
// internal helpers
// =========================================================================

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}
