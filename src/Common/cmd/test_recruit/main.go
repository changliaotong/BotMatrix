package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"BotMatrix/common/ai/employee"
	"BotMatrix/common/config"
	"BotMatrix/common/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Config
	if err := config.InitConfig("config.json"); err != nil {
		log.Printf("Warning: Failed to load config file: %v", err)
	}

	// DB Config
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.GlobalConfig.PGHost,
		config.GlobalConfig.PGPort,
		config.GlobalConfig.PGUser,
		config.GlobalConfig.PGPassword,
		config.GlobalConfig.PGDBName,
		config.GlobalConfig.PGSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	factory := employee.NewDigitalEmployeeFactory(db)
	ctx := context.Background()

	// 1. Get Job ID
	var job models.DigitalJob
	if err := db.Where("name = ?", "Code Repair Expert").First(&job).Error; err != nil {
		log.Fatalf("Job not found: %v", err)
	}

	// 2. Get Variants
	var varA, varB models.ABVariant
	db.Where("name LIKE ?", "%Conservative%").First(&varA)
	db.Where("name LIKE ?", "%Aggressive%").First(&varB)

	// 3. Recruit Variant A
	fmt.Println("--- Recruiting Variant A (Conservative) ---")
	empA, err := factory.Recruit(ctx, employee.RecruitParams{
		JobID:        job.ID,
		EnterpriseID: 1, // Mock Ent ID
		VariantID:    &varA.ID,
		BotID:        fmt.Sprintf("bot_a_%d", time.Now().UnixNano()),
	})
	if err != nil {
		log.Fatalf("Failed to recruit A: %v", err)
	}
	printEmployeeDetails(db, empA)

	// 4. Recruit Variant B
	fmt.Println("\n--- Recruiting Variant B (Aggressive) ---")
	empB, err := factory.Recruit(ctx, employee.RecruitParams{
		JobID:        job.ID,
		EnterpriseID: 1,
		VariantID:    &varB.ID,
		BotID:        fmt.Sprintf("bot_b_%d", time.Now().UnixNano()),
	})
	if err != nil {
		log.Fatalf("Failed to recruit B: %v", err)
	}
	printEmployeeDetails(db, empB)
}

func printEmployeeDetails(db *gorm.DB, emp *models.DigitalEmployee) {
	fmt.Printf("Employee: %s (ID: %d)\n", emp.Name, emp.ID)

	// Fetch Agent to see Prompt
	var agent models.AIAgent
	if err := db.First(&agent, emp.AgentID).Error; err != nil {
		fmt.Printf("  Error fetching agent: %v\n", err)
		return
	}

	fmt.Printf("  Agent: %s\n", agent.Name)
	fmt.Printf("  Temperature: %.2f\n", agent.Temperature)
	fmt.Printf("  System Prompt Preview: %.100s...\n", agent.Prompt)
}
