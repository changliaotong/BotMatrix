package database

import (
	"database/sql"
	"fmt"
	"log"

	"BotMatrix/common/config"

	_ "github.com/lib/pq"
)

// InitDB initializes the PostgreSQL database connection
func InitDB(cfg *config.AppConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDBName, cfg.PGSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to database %s at %s:%d", cfg.PGDBName, cfg.PGHost, cfg.PGPort)
	return db, nil
}
