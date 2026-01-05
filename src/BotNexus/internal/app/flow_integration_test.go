package app

import (
	"BotMatrix/common"
	"BotMatrix/common/bot"
	"BotMatrix/common/config"
	"BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/tasks"
	"BotMatrix/common/types"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCoreMessageFlowIntegration(t *testing.T) {
	// Initialize logger
	log.InitLogger(log.Config{
		Level:       "debug",
		Format:      "console",
		Development: true,
	})

	// 1. Setup Manager with minimal config
	config.ENABLE_SKILL = true

	// 2. Initialize GORM with SQLite in-memory
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}

	// Auto Migrate necessary models
	err = db.AutoMigrate(
		&models.BotEntityGORM{},
		&models.RoutingRuleGORM{},
		&models.UserGORM{},
		&models.UserIdentity{},
		&models.MessageLogGORM{},
		&models.Task{},
		&models.Execution{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	sqlDB, _ := db.DB()

	// Setup MiniRedis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to run miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	m := &Manager{
		Manager: bot.NewManager(),
	}
	m.Config = config.GlobalConfig
	m.GORMDB = db
	m.DB = sqlDB
	m.Rdb = rdb

	// Mock TaskManager and AI for regex matching
	m.TaskManager = tasks.NewTaskManager(db, nil, m, "nexus")
	// We need to initialize AI and add a skill with regex
	// types.Capability is the common type
	testSkill := types.Capability{
		Name:        "test_skill",
		Description: "Test Skill",
		Regex:       "^test (.+)$",
	}

	m.TaskManager.AI.UpdateSkills([]types.Capability{testSkill})

	// Mock Core Plugin
	m.Core = common.NewCorePlugin(m.Manager)

	// 2. Mock Bot
	botID := "test_bot_123"
	bot := &types.BotClient{
		SelfID:    botID,
		Platform:  "qq",
		Connected: time.Now(),
		Protocol:  "v11",
	}
	m.Mutex.Lock()
	m.Bots[botID] = bot
	m.Mutex.Unlock()

	// 3. Mock Worker
	workerID := "test_worker_456"
	worker := &types.WorkerClient{
		ID:            workerID,
		LastHeartbeat: time.Now(),
		Capabilities: []types.WorkerCapability{
			{
				Name: "test_skill",
			},
		},
	}
	m.Mutex.Lock()
	m.Workers = append(m.Workers, worker)
	m.Mutex.Unlock()

	// 4. Intercept command sent to worker
	var capturedCmd types.WorkerCommand
	var wg sync.WaitGroup
	wg.Add(1)
	m.OnCommandSent = func(wID string, cmd types.WorkerCommand) {
		if wID == workerID {
			capturedCmd = cmd
			wg.Done()
		}
	}

	// 5. Send a message from Bot to Nexus
	msg := types.InternalMessage{
		SelfID:      botID,
		UserID:      "user_1",
		PostType:    "message",
		MessageType: "private",
		RawMessage:  "test hello world",
		Time:        time.Now().Unix(),
	}

	// Trigger the flow
	m.handleBotMessage(bot, msg)

	// Wait for the command to be sent (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for command to be sent to worker")
	}

	// 6. Verify skill call was triggered
	assert.Equal(t, "skill_call", capturedCmd.Type)
	assert.Equal(t, "test_skill", capturedCmd.Skill)

	// 7. Simulate Worker processing and returning result
	correlationID := capturedCmd.CorrelationID
	resultMsg := "Processed: hello world"

	// Mock HandleSkillResult
	// In a real flow, the worker would send a skill_result via Redis or WS.
	// Nexus's HandleSkillResult handles the result.

	m.HandleSkillResult(types.SkillResult{
		CorrelationID: correlationID,
		WorkerID:      workerID,
		Status:        "success",
		Result:        resultMsg,
	})

	// 8. Verify the result flow (optional, if we want to check if it reached the bot)
	// Since bot.Conn is nil, it would have broadcasted an event.
	// We could mock the event broadcaster if needed, but the current test already proves the core loop.

	t.Log("Core Message Flow Integration Test Passed!")
}

func TestMultiWorkerRouting(t *testing.T) {
	// Initialize logger
	log.InitLogger(log.Config{
		Level:       "debug",
		Format:      "console",
		Development: true,
	})

	// Setup Manager
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&models.BotEntityGORM{}, &models.RoutingRuleGORM{}, &models.UserIdentity{}, &models.MessageLogGORM{}, &models.Task{}, &models.Execution{})
	sqlDB, _ := db.DB()
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	m := &Manager{Manager: bot.NewManager()}
	m.GORMDB = db
	m.DB = sqlDB
	m.Rdb = rdb
	m.TaskManager = tasks.NewTaskManager(db, rdb, m, "nexus")

	// 1. Setup two workers with different skills
	w1 := &types.WorkerClient{ID: "worker_weather", LastHeartbeat: time.Now(), Capabilities: []types.WorkerCapability{{Name: "weather", Regex: "天气"}}}
	w2 := &types.WorkerClient{ID: "worker_calc", LastHeartbeat: time.Now(), Capabilities: []types.WorkerCapability{{Name: "calc", Regex: "计算"}}}
	m.Mutex.Lock()
	m.Workers = append(m.Workers, w1, w2)
	m.Mutex.Unlock()

	// Sync skills to AI
	m.SyncWorkerSkills()

	// 2. Test routing to worker_weather
	msg1 := types.InternalMessage{RawMessage: "查询北京天气", SelfID: "bot1", UserID: "user1"}
	target1 := m.getTargetWorkerID(msg1)
	assert.Equal(t, "worker_weather", target1)

	// 3. Test routing to worker_calc
	msg2 := types.InternalMessage{RawMessage: "帮我计算 1+1", SelfID: "bot1", UserID: "user1"}
	target2 := m.getTargetWorkerID(msg2)
	assert.Equal(t, "worker_calc", target2)

	t.Log("Multi-Worker Routing Test Passed!")
}
