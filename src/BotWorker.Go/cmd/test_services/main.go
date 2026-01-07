package main

import (
	"fmt"
	"log"

	"botworker/internal/services"
	"botworker/internal/store"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Dummy connection string - we just want to verify compilation and struct initialization
	// This will fail to connect but that's expected. We are checking code structure.
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DryRun: true, // Don't actually connect
	})
	if err != nil {
		log.Printf("DB Init failed (expected): %v", err)
	}

	// Initialize Store
	s := store.NewStore(db)
	fmt.Println("Store initialized successfully")

	// Initialize Services
	userService := services.NewUserService(s)
	groupService := services.NewGroupService(s)
	economyService := services.NewEconomyService(s)
	robotService := services.NewRobotService(s)

	fmt.Printf("Services initialized: User=%v, Group=%v, Economy=%v, Robot=%v\n",
		userService != nil, groupService != nil, economyService != nil, robotService != nil)

	// Verify Interfaces
	var _ services.IUserService = userService
	var _ services.IGroupService = groupService
	var _ services.IEconomyService = economyService
	var _ services.IRobotService = robotService

	fmt.Println("Interface verification passed")
}
