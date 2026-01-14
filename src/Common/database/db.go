package database

import (
	"database/sql"
	"fmt"
	"log"

	"BotMatrix/common/config"
	"BotMatrix/common/models"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitGORM initializes the GORM database connection
func (m *GORMManager) InitGORM(cfg *config.AppConfig) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDBName, cfg.PGSSLMode)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect gorm: %v", err)
	}

	// Manual migration: Ensure users table has qq column
	if err := db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS qq VARCHAR(20)").Error; err != nil {
		log.Printf("[DB] Warning: Failed to add qq column to users table: %v", err)
	}

	// Auto migrate in batches to handle potential failures better
	modelsToMigrate := []any{
		&models.AIProviderGORM{},
		&models.AIModelGORM{},
		&models.AIAgentGORM{},
		&models.AISessionGORM{},
		&models.AIChatMessageGORM{},
		&models.AIUsageLogGORM{},
		&models.DigitalEmployeeGORM{},
		&models.DigitalEmployeeKpiGORM{},
		&models.BotEntityGORM{},
		&models.MessageLogGORM{},
		&models.UserGORM{},
		&models.RoutingRuleGORM{},
		&models.GroupCacheGORM{},
		&models.MemberCacheGORM{},
		&models.FriendCacheGORM{},
		&models.DigitalRoleTemplateGORM{},
		&models.CognitiveMemoryGORM{},
		&models.AIAgentTraceGORM{},
		&models.BotSkillPermissionGORM{},
		&models.MCPServerGORM{},
		&models.MCPToolGORM{},
		&models.MessageStatGORM{},
		&models.UserLoginTokenGORM{},
		&models.Member{},
		&models.Sz84Group{},
	}

	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("GORM AutoMigrate failed for model %T: %v", model, err)
		}
	}

	// Migrate other models
	if err := db.AutoMigrate(
		&models.FissionConfigGORM{},
		&models.InvitationGORM{},
		&models.FissionTaskGORM{},
		&models.UserFissionRecordGORM{},
		&models.FissionRewardLogGORM{},
		&models.AIPromptTemplateGORM{},
		&models.AIKnowledgeBaseGORM{},
		&models.AISkillGORM{},
		&models.AITrainingDataGORM{},
		&models.AIIntentGORM{},
		&models.AIIntentRoutingGORM{},
		&models.GroupBotRoleGORM{},
		&models.EnterpriseGORM{},
		&models.EnterpriseMemberGORM{},
		&models.PlatformAccountGORM{},
		&models.B2BConnectionGORM{},
		&models.B2BSkillSharingGORM{},
		&models.DigitalEmployeeDispatchGORM{},
		&models.DigitalEmployeeTodoGORM{},
		&models.DigitalEmployeeTaskGORM{},
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

	m.DB = db
	log.Printf("Successfully initialized GORM with database %s", cfg.PGDBName)

	// Sync call_count from ai_usage_logs if necessary
	go func() {
		log.Printf("[DB] Starting call_count sync from ai_usage_logs...")
		var counts []struct {
			AgentID uint
			Total   int
		}
		if err := db.Model(&models.AIUsageLogGORM{}).Select("agent_id, count(*) as total").Group("agent_id").Scan(&counts).Error; err == nil {
			for _, c := range counts {
				if c.AgentID > 0 {
					db.Model(&models.AIAgentGORM{}).Where("id = ?", c.AgentID).Update("call_count", c.Total)
				}
			}
			log.Printf("[DB] Successfully synced call_count for %d agents", len(counts))
		} else {
			log.Printf("[DB] Failed to sync call_count: %v", err)
		}
	}()

	return nil
}

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
