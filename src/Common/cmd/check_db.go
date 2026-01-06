package main

import (
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	config.InitConfig("config.json")
	db, err := database.InitGORM(config.GlobalConfig)
	if err != nil {
		log.Fatal(err)
	}

	var tasks []models.DigitalEmployeeTask
	db.Find(&tasks)

	fmt.Printf("Total tasks found: %d\n", len(tasks))
	for _, t := range tasks {
		data, _ := json.MarshalIndent(t, "", "  ")
		fmt.Println(string(data))
	}
}
