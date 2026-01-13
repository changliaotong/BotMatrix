package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	// DB Config from config.json
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

	// 1. Migrate Tables
	log.Println("Migrating tables...")

	// DROP tables first to ensure clean state (dev only)
	db.Migrator().DropTable(
		&models.DigitalJob{},
		&models.DigitalCapability{},
		&models.JobCapabilityRelation{},
		&models.EmployeeSkillRelation{},
		&models.EmployeeJobRelation{},
		&models.ABExperiment{},
		&models.ABVariant{},
		&models.EmployeeExperimentRelation{},
		&models.MemorySnapshot{},
	)

	err = db.AutoMigrate(
		&models.AIProvider{},
		&models.AIModel{}, // Prerequisite for AIAgent
		&models.AIAgent{},
		&models.DigitalEmployee{},
		&models.DigitalJob{},
		&models.DigitalCapability{},
		&models.JobCapabilityRelation{},
		&models.EmployeeSkillRelation{},
		&models.EmployeeJobRelation{},
		&models.ABExperiment{},
		&models.ABVariant{},
		&models.EmployeeExperimentRelation{},
		&models.MemorySnapshot{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 1.5 Seed AI Provider and Model
	var provider models.AIProvider
	if err := db.FirstOrCreate(&provider, models.AIProvider{
		Name:      "OpenAI",
		Type:      "openai",
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "sk-mock-key",
		IsEnabled: true,
	}).Error; err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	var aiModel models.AIModel
	if err := db.FirstOrCreate(&aiModel, models.AIModel{
		ProviderID:   provider.ID,
		ModelID:      "gpt-3.5-turbo",
		ModelName:    "GPT-3.5 Turbo",
		Capabilities: `["chat"]`,
		IsDefault:    true,
	}).Error; err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}
	log.Printf("Seeded AI Model: %s (ID: %d)", aiModel.ModelName, aiModel.ID)

	// 2. Create Job: Code Repair Expert
	jobName := "Code Repair Expert"
	var job models.DigitalJob
	if err := db.Where("name = ?", jobName).First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			job = models.DigitalJob{
				Name:        jobName,
				Description: "Specialized in analyzing and fixing Go/Python code bugs.",
				Level:       3,
				Department:  "DevTeam",
				BaseSalary:  500,
				IsActive:    true,
			}
			if err := db.Create(&job).Error; err != nil {
				log.Fatalf("Failed to create job: %v", err)
			}
			log.Printf("Created Job: %s (ID: %d)", job.Name, job.ID)
		} else {
			log.Fatalf("Error finding job: %v", err)
		}
	} else {
		log.Printf("Job exists: %s (ID: %d)", job.Name, job.ID)
	}

	// 3. Create Experiment: EXP-2026-001
	expName := "EXP-2026-001: Developer Persona Test"
	var exp models.ABExperiment
	if err := db.Where("name = ?", expName).First(&exp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			startTime := time.Now()
			endTime := startTime.AddDate(0, 1, 0)
			exp = models.ABExperiment{
				Name:        expName,
				Description: "Compare Conservative vs Aggressive personas for code repair tasks.",
				TargetJobID: job.ID,
				Status:      "running",
				Metric:      "acceptance_rate",
				StartDate:   &startTime,
				EndDate:     &endTime,
			}
			if err := db.Create(&exp).Error; err != nil {
				log.Fatalf("Failed to create experiment: %v", err)
			}
			log.Printf("Created Experiment: %s (ID: %d)", exp.Name, exp.ID)
		}
	} else {
		log.Printf("Experiment exists: %s (ID: %d)", exp.Name, exp.ID)
	}

	// 4. Create Variants
	createVariant(db, exp.ID, "Variant A (Conservative)", "Standard, careful persona.", map[string]interface{}{
		"temperature": 0.4,
		"prompt_base": "You are a careful Code Repair Expert. Prioritize safety and backward compatibility.",
		"model_id":    1, // Assuming 1 is GPT-4 or similar
	})

	createVariant(db, exp.ID, "Variant B (Aggressive)", "Performance-focused, bold persona.", map[string]interface{}{
		"temperature": 0.9,
		"prompt_base": "You are a ruthless Code Repair Expert. Refuse any code that is suboptimal. Rewrite aggressively for performance.",
		"model_id":    1,
	})

	log.Println("Setup completed successfully.")
}

func createVariant(db *gorm.DB, expID uint, name, desc string, config map[string]interface{}) {
	var v models.ABVariant
	if err := db.Where("experiment_id = ? AND name = ?", expID, name).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			configBytes, _ := json.Marshal(config)
			v = models.ABVariant{
				ExperimentID:   expID,
				Name:           name,
				Description:    desc,
				ConfigOverride: string(configBytes),
				Allocation:     50,
			}
			if err := db.Create(&v).Error; err != nil {
				log.Printf("Failed to create variant %s: %v", name, err)
			} else {
				log.Printf("Created Variant: %s (ID: %d)", v.Name, v.ID)
			}
		}
	} else {
		log.Printf("Variant exists: %s (ID: %d)", v.Name, v.ID)
	}
}
