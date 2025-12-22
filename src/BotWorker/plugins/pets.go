package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// PetPlugin 宠物系统插件
type PetPlugin struct {
	db        *sql.DB
	points    *PointsPlugin
	cmdParser *CommandParser // 命令解析器
}

// NewPetPlugin 创建宠物系统插件实例
func NewPetPlugin(database *sql.DB, pointsPlugin *PointsPlugin) *PetPlugin {
	return &PetPlugin{
		db:        database,
		points:    pointsPlugin,
		cmdParser: NewCommandParser(),
	}
}

func (p *PetPlugin) Name() string {
	return "pets"
}

func (p *PetPlugin) Description() string {
	return "宠物系统插件，支持领养宠物、喂食、玩耍、升级等功能（集成积分系统）"
}

func (p *PetPlugin) Version() string {
	return "1.1.0"
}

func (p *PetPlugin) Init(robot plugin.Robot) {
	log.Println("加载宠物系统插件")

	// 处理领养宠物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				HandleFeatureDisabled(robot, event, "pets")
				return nil
			}
		}

		// 检查是否为领养宠物命令
		if match, _ := p.cmdParser.MatchCommand("adopt|领养宠物|领养", event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，领养失败")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 检查积分是否足够 (领养需要 50 积分)
		adoptCost := 50
		if p.points != nil && p.points.GetPoints(userIDStr) < adoptCost {
			p.sendMessage(robot, event, fmt.Sprintf("领养宠物需要 %d 积分，你当前的积分不足", adoptCost))
			return nil
		}

		// 检查用户是否已经有宠物 (从数据库查)
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil {
			log.Printf("查询用户宠物失败: %v", err)
			p.sendMessage(robot, event, "查询宠物信息失败，请稍后再试")
			return nil
		}

		if len(userPets) >= 3 {
			p.sendMessage(robot, event, "你最多只能领养3只宠物")
			return nil
		}

		// 随机生成宠物类型
		petTypes := []string{"🐱 猫咪", "🐶 狗狗", "🐰 兔子", "🐹 仓鼠", "🐻 小熊", "🐼 熊猫", "🐯 老虎", "🦁 狮子"}
		petType := petTypes[rand.Intn(len(petTypes))]

		// 生成宠物ID
		petID := fmt.Sprintf("pet_%d_%d", time.Now().Unix(), userID)

		// 创建新宠物模型
		petModel := &db.PetModel{
			PetID:     petID,
			UserID:    userIDStr,
			Name:      fmt.Sprintf("%d的%s", userID, petType),
			Type:      petType,
			Level:     1,
			Exp:       0,
			Hunger:    80,
			Happiness: 80,
			Health:    100,
		}

		// 存储宠物到数据库
		if err := db.CreatePet(p.db, petModel); err != nil {
			log.Printf("保存宠物到数据库失败: %v", err)
			p.sendMessage(robot, event, "领养失败，请联系管理员")
			return nil
		}

		// 扣除积分
		if p.points != nil {
			p.points.AddPoints(userIDStr, -adoptCost, "领养宠物", "pet_adopt")
		}

		p.sendMessage(robot, event, fmt.Sprintf("🎉 恭喜你花费 %d 积分领养了一只%s！\n宠物名字：%s\n等级：%d\n经验：%d\n饥饿值：%d\n快乐值：%d\n健康值：%d",
			adoptCost, petType, petModel.Name, petModel.Level, petModel.Exp, petModel.Hunger, petModel.Happiness, petModel.Health))

		return nil
	})

	// 处理查看宠物命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				HandleFeatureDisabled(robot, event, "pets")
				return nil
			}
		}

		// 检查是否为查看宠物命令
		if match, _ := p.cmdParser.MatchCommand("pets|我的宠物|宠物", event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		// 获取用户的宠物
		userIDStr := fmt.Sprintf("%d", userID)
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil {
			log.Printf("查询用户宠物失败: %v", err)
			p.sendMessage(robot, event, "查询宠物信息失败")
			return nil
		}

		if len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 发送宠物列表
		msg := "🐾 你的宠物 🐾\n"
		msg += "------------------------\n"
		for i, pet := range userPets {
			msg += fmt.Sprintf("%d. %s\n", i+1, pet.Name)
			msg += fmt.Sprintf("   类型：%s\n", pet.Type)
			msg += fmt.Sprintf("   等级：%d\n", pet.Level)
			msg += fmt.Sprintf("   经验：%d/%d\n", pet.Exp, pet.Level*100)
			msg += fmt.Sprintf("   饥饿值：%d/100\n", pet.Hunger)
			msg += fmt.Sprintf("   快乐值：%d/100\n", pet.Happiness)
			msg += fmt.Sprintf("   健康值：%d/100\n", pet.Health)
			msg += "------------------------\n"
		}

		p.sendMessage(robot, event, msg)

		return nil
	})

	// 处理喂食命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				HandleFeatureDisabled(robot, event, "pets")
				return nil
			}
		}

		// 检查是否为喂食命令
		match, _, params := p.cmdParser.MatchCommandWithParams("feed|喂食", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 喂食消耗 5 积分
		feedCost := 5
		if p.points != nil && p.points.GetPoints(userIDStr) < feedCost {
			p.sendMessage(robot, event, fmt.Sprintf("喂食需要 %d 积分，你当前的积分不足", feedCost))
			return nil
		}

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := strconv.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 喂食
		oldHunger := pet.Hunger
		pet.Hunger += 20
		if pet.Hunger > 100 {
			pet.Hunger = 100
		}
		oldHappiness := pet.Happiness
		pet.Happiness += 5
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		oldExp := pet.Exp
		pet.Exp += 10

		// 检查升级
		p.checkLevelUp(pet)

		// 更新到数据库
		if err := db.UpdatePet(p.db, pet); err != nil {
			log.Printf("更新宠物失败: %v", err)
			p.sendMessage(robot, event, "操作失败，请重试")
			return nil
		}

		// 扣除积分
		if p.points != nil {
			p.points.AddPoints(userIDStr, -feedCost, "喂食宠物", "pet_feed")
		}

		p.sendMessage(robot, event, fmt.Sprintf("🍖 你花费 %d 积分给%s喂食了！\n饥饿值：%d → %d\n快乐值：%d → %d\n经验值：%d → %d",
			feedCost, pet.Name, oldHunger, pet.Hunger, oldHappiness, pet.Happiness, oldExp, pet.Exp))

		return nil
	})

	// 处理玩耍命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				HandleFeatureDisabled(robot, event, "pets")
				return nil
			}
		}

		// 检查是否为玩耍命令
		match, _, params := p.cmdParser.MatchCommandWithParams("play|玩耍", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := strconv.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 玩耍
		oldHappiness := pet.Happiness
		pet.Happiness += 20
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		oldHunger := pet.Hunger
		pet.Hunger -= 10
		if pet.Hunger < 0 {
			pet.Hunger = 0
		}
		oldExp := pet.Exp
		pet.Exp += 15

		// 检查升级
		p.checkLevelUp(pet)

		// 更新到数据库
		if err := db.UpdatePet(p.db, pet); err != nil {
			log.Printf("更新宠物失败: %v", err)
			p.sendMessage(robot, event, "操作失败，请重试")
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("🎮 你和%s玩耍了！\n快乐值：%d → %d\n饥饿值：%d → %d\n经验值：%d → %d",
			pet.Name, oldHappiness, pet.Happiness, oldHunger, pet.Hunger, oldExp, pet.Exp))

		return nil
	})

	// 处理洗澡命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				HandleFeatureDisabled(robot, event, "pets")
				return nil
			}
		}

		// 检查是否为洗澡命令
		match, _, params := p.cmdParser.MatchCommandWithParams("wash|洗澡", `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物，使用/领养命令领养一只吧")
			return nil
		}

		// 解析宠物编号
		petIndex := 0
		if len(params) > 0 && params[0] != "" {
			index, err := strconv.Atoi(params[0])
			if err == nil && index > 0 && index <= len(userPets) {
				petIndex = index - 1
			}
		}

		pet := userPets[petIndex]

		// 洗澡
		oldHealth := pet.Health
		pet.Health += 15
		if pet.Health > 100 {
			pet.Health = 100
		}
		oldHappiness := pet.Happiness
		pet.Happiness += 10
		if pet.Happiness > 100 {
			pet.Happiness = 100
		}
		oldExp := pet.Exp
		pet.Exp += 5

		// 检查升级
		p.checkLevelUp(pet)

		// 更新到数据库
		if err := db.UpdatePet(p.db, pet); err != nil {
			log.Printf("更新宠物失败: %v", err)
			p.sendMessage(robot, event, "操作失败，请重试")
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("🛁 你给%s洗澡了！\n健康值：%d → %d\n快乐值：%d → %d\n经验值：%d → %d",
			pet.Name, oldHealth, pet.Health, oldHappiness, pet.Happiness, oldExp, pet.Exp))

		return nil
	})

	// 处理改名命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "pets") {
				return nil
			}
		}

		// 检查是否为改名命令
		match, _, params := p.cmdParser.MatchCommandWithParams("rename|改名", `(\d+)\s+(\S+)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, "你还没有宠物")
			return nil
		}

		index, _ := strconv.Atoi(params[0])
		if index <= 0 || index > len(userPets) {
			p.sendMessage(robot, event, "宠物编号不正确")
			return nil
		}

		newName := params[1]
		pet := userPets[index-1]
		oldName := pet.Name
		pet.Name = newName

		if err := db.UpdatePet(p.db, pet); err != nil {
			log.Printf("改名失败: %v", err)
			p.sendMessage(robot, event, "改名失败")
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("🏷️ 成功将宠物 %s 改名为 %s", oldName, newName))
		return nil
	})

	// 定时更新宠物状态（每小时）
	go p.updatePetStatus()
}

// checkLevelUp 检查宠物是否升级
func (p *PetPlugin) checkLevelUp(pet *db.PetModel) {
	requiredExp := pet.Level * 100
	if pet.Exp >= requiredExp {
		pet.Level++
		pet.Exp -= requiredExp
		pet.Health = 100
		pet.Happiness = 100
		pet.Hunger = 80

		log.Printf("宠物%s升级到%d级", pet.Name, pet.Level)
	}
}

// updatePetStatus 定时更新宠物状态
func (p *PetPlugin) updatePetStatus() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// 从数据库加载所有宠物
		allPets, err := db.GetAllPets(p.db)
		if err != nil {
			log.Printf("定时任务：加载所有宠物失败: %v", err)
			continue
		}

		for _, pet := range allPets {
			// 每小时减少饥饿值和快乐值
			pet.Hunger -= 5
			if pet.Hunger < 0 {
				pet.Hunger = 0
			}

			pet.Happiness -= 5
			if pet.Happiness < 0 {
				pet.Happiness = 0
			}

			// 饥饿值或快乐值过低会影响健康
			if pet.Hunger < 20 || pet.Happiness < 20 {
				pet.Health -= 10
				if pet.Health < 0 {
					pet.Health = 0
				}
			}

			// 更新到数据库
			if err := db.UpdatePet(p.db, pet); err != nil {
				log.Printf("定时任务：更新宠物 %s 失败: %v", pet.PetID, err)
			}
		}
	}
}

// sendMessage 发送消息
func (p *PetPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	_, _ = SendTextReply(robot, event, msg)
}
