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

	fmt.Println("--- AI Providers ---")
	var providers []models.AIProviderGORM
	db.Find(&providers)
	for _, p := range providers {
		keyLen := len(p.APIKey)
		maskedKey := ""
		if keyLen > 8 {
			maskedKey = p.APIKey[:4] + "..." + p.APIKey[keyLen-4:]
		} else if keyLen > 0 {
			maskedKey = "****"
		}
		fmt.Printf("ID: %d, Name: %s, Type: %s, BaseURL: %s, APIKey: %s (len: %d)\n", p.ID, p.Name, p.Type, p.BaseURL, maskedKey, keyLen)
	}

	fmt.Println("\n--- AI Models ---")
	var aiModels []models.AIModelGORM
	db.Find(&aiModels)
	for _, m := range aiModels {
		fmt.Printf("ID: %d, ModelID: %s, Name: %s, ProviderID: %d, IsDefault: %v\n", m.ID, m.ModelID, m.ModelName, m.ProviderID, m.IsDefault)
	}

	fmt.Println("\n--- Searching for '早喵' Agent ---")
	var zaoMiao models.AIAgentGORM
	result := db.Where("name = ?", "早喵").First(&zaoMiao)
	if result.Error != nil {
		fmt.Printf("Agent '早喵' not found: %v\n", result.Error)
	} else {
		fmt.Printf("ID: %d, Name: %s, ModelID: %d\n", zaoMiao.ID, zaoMiao.Name, zaoMiao.ModelID)
		if zaoMiao.ModelID != 0 {
			var model models.AIModelGORM
			db.First(&model, zaoMiao.ModelID)
			fmt.Printf("Model ID: %d, ModelName: %s, APIModelID: %s, ProviderID: %d\n", model.ID, model.ModelName, model.ModelID, model.ProviderID)

			var provider models.AIProviderGORM
			db.First(&provider, model.ProviderID)
			fmt.Printf("Provider ID: %d, ProviderName: %s, APIKey length: %d\n", provider.ID, provider.Name, len(provider.APIKey))
		}
	}
	return
}
