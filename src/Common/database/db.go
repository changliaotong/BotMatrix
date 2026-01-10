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

	// 预先检查并修复一些关键表的列，防止 AutoMigrate 失败
	m.fixCommonColumns(db)

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

	m.DB = db
	log.Printf("Successfully initialized GORM with database %s", cfg.PGDBName)

	// Sync call_count from ai_usage_logs if necessary
	go func() {
		log.Printf("[DB] Starting call_count sync from ai_usage_logs...")
		var counts []struct {
			AgentID uint `gorm:"column:AgentId"`
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

// fixCommonColumns 预先检查并修复一些关键表的列，防止 AutoMigrate 失败
func (m *GORMManager) fixCommonColumns(db *gorm.DB) {
	// 1. 修复 users 表
	if db.Migrator().HasTable("users") {
		// 检查 deleted_at
		if !db.Migrator().HasColumn("users", "deleted_at") {
			log.Println("[DB] Adding missing column users.deleted_at")
			db.Exec("ALTER TABLE users ADD COLUMN deleted_at timestamptz")
			db.Exec("CREATE INDEX idx_users_deleted_at ON users(deleted_at)")
		}
		// 检查 platform
		if !db.Migrator().HasColumn("users", "platform") {
			log.Println("[DB] Adding missing column users.platform")
			db.Exec("ALTER TABLE users ADD COLUMN platform varchar(32)")
		}
		// 检查 platform_id
		if !db.Migrator().HasColumn("users", "platform_id") {
			log.Println("[DB] Adding missing column users.platform_id")
			db.Exec("ALTER TABLE users ADD COLUMN platform_id varchar(64)")
		}
	}

	// 2. 修复 BotEntity 表 (注意表名可能是 BotEntity 或 bot_entities)
	botTable := "BotEntity"
	if db.Migrator().HasTable(botTable) {
		if !db.Migrator().HasColumn(botTable, "platform") {
			log.Println("[DB] Adding missing column BotEntity.platform")
			db.Exec("ALTER TABLE \"BotEntity\" ADD COLUMN platform varchar(32)")
		}
	}

	// 3. 修复 DigitalRoleTemplate 表
	templateTable := "DigitalRoleTemplate"
	if db.Migrator().HasTable(templateTable) {
		if !db.Migrator().HasColumn(templateTable, "name") {
			log.Println("[DB] Adding missing column DigitalRoleTemplate.name")
			db.Exec("ALTER TABLE \"DigitalRoleTemplate\" ADD COLUMN name varchar(100)")
		}
	}
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
