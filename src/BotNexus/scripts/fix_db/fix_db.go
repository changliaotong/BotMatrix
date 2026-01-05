package main

import (
	"fmt"
	"log"

	"BotMatrix/common/config"
	"BotMatrix/common/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config.InitConfig("config.json")
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.GlobalConfig.PGHost, config.GlobalConfig.PGPort, config.GlobalConfig.PGUser, config.GlobalConfig.PGPassword, config.GlobalConfig.PGDBName, config.GlobalConfig.PGSSLMode)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	fmt.Println("Updating models...")
	// Update all Doubao models to set APIModelID = ModelName
	var doubaoModels []models.AIModelGORM
	db.Where("provider_id = ?", 7).Find(&doubaoModels)
	for _, m := range doubaoModels {
		if m.ModelID == "" {
			fmt.Printf("Setting ModelID for %s (ID: %d) to %s\n", m.ModelName, m.ID, m.ModelName)
			db.Model(&m).Update("api_model_id", m.ModelName)
		}
	}

	fmt.Println("\nUpdating '早喵' agent...")
	// Update '早喵' to use a model from Doubao provider
	// Let's use ID: 1 which is "doubao-1.5-vision-lite-250315"
	var zaoMiao models.AIAgentGORM
	result := db.Where("name = ?", "早喵").First(&zaoMiao)
	if result.Error == nil {
		fmt.Printf("Setting '早喵' (ID: %d) ModelID to 1\n", zaoMiao.ID)
		db.Model(&zaoMiao).Update("model_id", 1)
	} else {
		fmt.Printf("Agent '早喵' not found: %v\n", result.Error)
	}

	fmt.Println("\nDone.")
}
