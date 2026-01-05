package database

import (
	"fmt"
	"log"

	"BotMatrix/common/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GORMManager manages GORM database operations
type GORMManager struct {
	DB *gorm.DB
}

// NewGORMManager creates a new GORMManager
func NewGORMManager(db *gorm.DB) *GORMManager {
	return &GORMManager{DB: db}
}

// InitGORM initializes the GORM database connection
func InitGORM(cfg *config.AppConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.PGHost, cfg.PGUser, cfg.PGPassword, cfg.PGDBName, cfg.PGPort, cfg.PGSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GORM database: %v", err)
	}

	log.Printf("GORM: Successfully connected to database %s", cfg.PGDBName)
	return db, nil
}
