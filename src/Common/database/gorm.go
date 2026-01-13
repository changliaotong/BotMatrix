package database

import (
	"fmt"
	"log"

	"BotMatrix/common/config"
	"BotMatrix/common/models"

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

// InitGORM initializes the GORM database connection based on config
func (gm *GORMManager) InitGORM(cfg *config.AppConfig) error {
	db, err := ConnectGORM(cfg)
	if err != nil {
		return err
	}

	// Auto migrate in batches to handle potential failures better
	modelsToMigrate := []any{
		&models.AIProvider{},
		&models.AIModel{},
		&models.AIAgent{},
		&models.AISession{},
		&models.AIChatMessage{},
		&models.AIUsageLog{},
		&models.DigitalEmployee{},
		&models.DigitalEmployeeKpi{},
		&models.Member{},
		&models.MessageLog{},
		&models.UserInfo{},
		&models.RoutingRule{},
		&models.GroupCache{},
		&models.MemberCache{},
		&models.FriendCache{},
		&models.DigitalRoleTemplate{},
		&models.CognitiveMemory{},
		&models.BotSkillPermission{},
		&models.MCPServer{},
		&models.MCPTool{},
		&models.MessageStat{},
		&models.UserLoginToken{},
		&models.User{},
		&models.BotEntity{},
	}

	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("GORM AutoMigrate failed for model %T: %v", model, err)
		}
	}

	// Migrate other models
	if err := db.AutoMigrate(
		&models.AIPromptTemplate{},
		&models.AIKnowledgeBase{},
		&models.AISkill{},
		&models.AITrainingData{},
		&models.AIIntent{},
		&models.AIIntentRouting{},
		&models.GroupBotRole{},
		&models.Enterprise{},
		&models.EnterpriseMember{},
		&models.PlatformAccount{},
		&models.B2BConnection{},
		&models.B2BSkillSharing{},
		&models.DigitalEmployeeDispatch{},
		&models.DigitalEmployeeTodo{},
		&models.DigitalEmployeeTask{},
		&models.Task{},
		&models.Execution{},
		&models.Tag{},
		&models.Strategy{},
		&models.AIDraft{},
		&models.UserIdentity{},
		&models.ShadowRule{},
		&models.TaskTag{},
	); err != nil {
		log.Printf("GORM AutoMigrate failed (remaining models): %v", err)
	}

	gm.DB = db
	log.Printf("Successfully initialized GORM")

	// Sync call_count from ai_usage_logs if necessary
	go func() {
		log.Printf("[DB] Starting call_count sync from ai_usage_logs...")
		var counts []struct {
			AgentID uint
			Total   int
		}
		if err := db.Model(&models.AIUsageLog{}).Select("\"AgentId\" as agent_id, count(*) as total").Group("\"AgentId\"").Scan(&counts).Error; err == nil {
			for _, c := range counts {
				if c.AgentID > 0 {
					db.Model(&models.AIAgent{}).Where("id = ?", c.AgentID).Update("call_count", c.Total)
				}
			}
			log.Printf("[DB] Successfully synced call_count for %d agents", len(counts))
		} else {
			log.Printf("[DB] Failed to sync call_count: %v", err)
		}
	}()

	return nil
}

// ConnectGORM connects to the database and returns a GORM DB instance
func ConnectGORM(cfg *config.AppConfig) (*gorm.DB, error) {
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
