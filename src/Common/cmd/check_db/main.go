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

	var empCount int64
	db.Model(&models.DigitalEmployee{}).Count(&empCount)
	log.Printf("Employees: %d", empCount)

	var activeEmpCount int64
	db.Model(&models.DigitalEmployee{}).Where("\"Status\" = ?", "active").Count(&activeEmpCount)
	log.Printf("Active Employees: %d", activeEmpCount)

	var goalCount int64
	db.Model(&models.DigitalFactoryGoal{}).Count(&goalCount)
	log.Printf("Goals: %d", goalCount)

	var taskCount int64
	db.Model(&models.DigitalEmployeeTask{}).Count(&taskCount)
	log.Printf("Tasks: %d", taskCount)
}
