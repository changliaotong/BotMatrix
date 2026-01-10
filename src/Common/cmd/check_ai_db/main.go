package main

import (
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"log"
)

func main() {
	config.InitConfig("d:/projects/BotMatrix/config.json")
	db, err := database.InitGORM(config.GlobalConfig)
	if err != nil {
		log.Fatal(err)
	}

	var providers []models.AIProvider
	if err := db.Find(&providers).Error; err != nil {
		log.Printf("Error querying AI providers: %v", err)
	} else {
		log.Printf("Found %d AI providers", len(providers))
		for _, p := range providers {
			log.Printf("- ID: %d, Name: %s, Type: %s, BaseURL: %s", p.ID, p.Name, p.Type, p.BaseURL)
		}
	}

	var modelsList []models.AIModel
	if err := db.Find(&modelsList).Error; err != nil {
		log.Printf("Error querying AI models: %v", err)
	} else {
		log.Printf("Found %d AI models", len(modelsList))
		for _, m := range modelsList {
			log.Printf("- ID: %d, Name: %s, ApiModelID: %s, IsDefault: %v", m.ID, m.ModelName, m.ApiModelID, m.IsDefault)
		}
	}

	// Seeding
	if len(providers) == 0 {
		log.Println("No AI providers found. Seeding default Mock provider...")
		mockProvider := models.AIProvider{
			Name:      "Mock Provider",
			Type:      "mock",
			BaseURL:   "",
			APIKey:    "mock-key",
			IsEnabled: true,
		}
		if err := db.Create(&mockProvider).Error; err != nil {
			log.Printf("Failed to create mock provider: %v", err)
		} else {
			log.Printf("Created mock provider with ID: %d", mockProvider.ID)

			// Create default model for mock
			mockModel := models.AIModel{
				ProviderID:   mockProvider.ID,
				ModelName:    "Mock Model",
				ApiModelID:   "mock-gpt-3.5",
				IsDefault:    true,
				Capabilities: "[\"chat\"]",
			}
			if err := db.Create(&mockModel).Error; err != nil {
				log.Printf("Failed to create mock model: %v", err)
			} else {
				log.Printf("Created mock model with ID: %d", mockModel.ID)
			}
		}
	}
}
