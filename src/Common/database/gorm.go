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
	DB       *gorm.DB
	LegacyDB *gorm.DB
}

// NewGORMManager creates a new GORMManager
func NewGORMManager(db *gorm.DB, legacyDB *gorm.DB) *GORMManager {
	return &GORMManager{DB: db, LegacyDB: legacyDB}
}

// InitGORM initializes the GORM database connection
func InitGORM(cfg *config.AppConfig) (*gorm.DB, error) {
	// Support SQLite for local testing if pg_host is empty or set to "sqlite"
	if cfg.PGHost == "" || cfg.PGHost == "sqlite" {
		dbName := cfg.PGDBName
		if dbName == "" {
			dbName = "botmatrix.db"
		}
		log.Printf("GORM: Using SQLite database: %s", dbName)
		return gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.PGHost, cfg.PGUser, cfg.PGPassword, cfg.PGDBName, cfg.PGPort, cfg.PGSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GORM database: %v", err)
	}

	log.Printf("GORM: Successfully connected to PostgreSQL database %s", cfg.PGDBName)
	return db, nil
}

// InitLegacyMSSQL initializes the SQL Server database connection (legacy)
func InitLegacyMSSQL(cfg *config.AppConfig) (*gorm.DB, error) {
	if cfg.MSSQLHost == "" {
		return nil, nil // Not configured, skip
	}

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		cfg.MSSQLUser, cfg.MSSQLPassword, cfg.MSSQLHost, cfg.MSSQLPort, cfg.MSSQLDBName)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Legacy MSSQL: %v", err)
	}

	log.Printf("GORM: Successfully connected to Legacy MSSQL database %s", cfg.MSSQLDBName)
	return db, nil
}
