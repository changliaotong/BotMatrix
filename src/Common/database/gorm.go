package database

import (
	"fmt"
	"log"

	"BotMatrix/common/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// GORMManager manages GORM database operations
type GORMManager struct {
	DB *gorm.DB // Current active database (MSSQL or PG)
}

// NewGORMManager creates a new GORMManager
func NewGORMManager(db *gorm.DB) *GORMManager {
	return &GORMManager{DB: db}
}

// InitDB initializes the GORM database connection based on config
func InitDB(cfg *config.AppConfig) (*gorm.DB, error) {
	// If MSSQL is configured, it takes priority (as per user request)
	if cfg.MSSQLHost != "" {
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cfg.MSSQLUser, cfg.MSSQLPassword, cfg.MSSQLHost, cfg.MSSQLPort, cfg.MSSQLDBName)
		
		db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQL Server: %v", err)
		}
		log.Printf("GORM: Successfully connected to SQL Server database %s", cfg.MSSQLDBName)
		return db, nil
	}

	// Fallback to PostgreSQL or SQLite
	if cfg.PGHost == "" || cfg.PGHost == "sqlite" {
		dbName := cfg.PGDBName
		if dbName == "" {
			dbName = "botmatrix.db"
		}
		log.Printf("GORM: Using SQLite database: %s", dbName)
		return gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=Asia/Shanghai",
		cfg.PGUser, cfg.PGPassword, cfg.PGHost, cfg.PGPort, cfg.PGDBName, cfg.PGSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	log.Printf("GORM: Successfully connected to PostgreSQL database %s", cfg.PGDBName)
	return db, nil
}
