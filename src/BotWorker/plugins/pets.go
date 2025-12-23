package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "pet_plugin_desc")
}

func (p *PetPlugin) Version() string {
	return "1.1.0"
}

func (p *PetPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "pet_plugin_loaded"))

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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "pet_cmd_adopt"), event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "pet_adopt_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 检查积分是否足够 (领养需要 50 积分)
		adoptCost := 50
		if p.points != nil && p.points.GetPoints(userIDStr) < adoptCost {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_adopt_insufficient_points"), adoptCost))
			return nil
		}

		// 检查用户是否已经有宠物 (从数据库查)
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil {
			log.Printf(common.T("", "pet_query_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_query_failed"))
			return nil
		}

		if len(userPets) >= 3 {
			p.sendMessage(robot, event, common.T("", "pet_adopt_limit"))
			return nil
		}

		// 随机生成宠物类型
		petTypes := []string{
			common.T("", "pet_type_cat"),
			common.T("", "pet_type_dog"),
			common.T("", "pet_type_rabbit"),
			common.T("", "pet_type_hamster"),
			common.T("", "pet_type_bear"),
			common.T("", "pet_type_panda"),
			common.T("", "pet_type_tiger"),
			common.T("", "pet_type_lion"),
		}
		petType := petTypes[rand.Intn(len(petTypes))]

		// 生成宠物ID
		petID := fmt.Sprintf("pet_%d_%d", time.Now().Unix(), userID)

		// 创建新宠物模型
		petModel := &db.PetModel{
			PetID:     petID,
			UserID:    userIDStr,
			Name:      fmt.Sprintf(common.T("", "pet_default_name"), userID, petType),
			Type:      petType,
			Level:     1,
			Exp:       0,
			Hunger:    80,
			Happiness: 80,
			Health:    100,
		}

		// 存储宠物到数据库
		if err := db.CreatePet(p.db, petModel); err != nil {
			log.Printf(common.T("", "pet_save_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_adopt_failed"))
			return nil
		}

		// 扣除积分
		if p.points != nil {
			p.points.AddPoints(userIDStr, -adoptCost, common.T("", "pet_adopt_action"), "pet_adopt")
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_adopt_success"),
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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "pet_cmd_list"), event.RawMessage); !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_userid"))
			return nil
		}

		// 获取用户的宠物
		userIDStr := fmt.Sprintf("%d", userID)
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil {
			log.Printf(common.T("", "pet_query_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_query_failed_brief"))
			return nil
		}

		if len(userPets) == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_pets"))
			return nil
		}

		// 发送宠物列表
		msg := common.T("", "pet_list_header")
		msg += common.T("", "pet_list_separator")
		for i, pet := range userPets {
			msg += fmt.Sprintf("%d. %s\n", i+1, pet.Name)
			msg += fmt.Sprintf(common.T("", "pet_info_type"), pet.Type)
			msg += fmt.Sprintf(common.T("", "pet_info_level"), pet.Level)
			msg += fmt.Sprintf(common.T("", "pet_info_exp"), pet.Exp, pet.Level*100)
			msg += fmt.Sprintf(common.T("", "pet_info_hunger"), pet.Hunger)
			msg += fmt.Sprintf(common.T("", "pet_info_happiness"), pet.Happiness)
			msg += fmt.Sprintf(common.T("", "pet_info_health"), pet.Health)
			msg += common.T("", "pet_list_separator")
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_feed"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 喂食消耗 5 积分
		feedCost := 5
		if p.points != nil && p.points.GetPoints(userIDStr) < feedCost {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_feed_insufficient_points"), feedCost))
			return nil
		}

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_pets"))
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
			log.Printf(common.T("", "pet_update_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_op_failed"))
			return nil
		}

		// 扣除积分
		if p.points != nil {
			p.points.AddPoints(userIDStr, -feedCost, common.T("", "pet_feed_action"), "pet_feed")
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_feed_success"),
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_play"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_pets"))
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
			log.Printf(common.T("", "pet_update_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_op_failed"))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_play_success"),
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_wash"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_pets"))
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
			log.Printf(common.T("", "pet_update_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_op_failed"))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_wash_success"),
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_rename"), `(\d+)\s+(\S+)`, event.RawMessage)
		if !match {
			return nil
		}

		userID := event.UserID
		userIDStr := fmt.Sprintf("%d", userID)

		// 获取用户的宠物
		userPets, err := db.GetPetsByUserID(p.db, userIDStr)
		if err != nil || len(userPets) == 0 {
			p.sendMessage(robot, event, common.T("", "pet_no_pets_brief"))
			return nil
		}

		index, _ := strconv.Atoi(params[0])
		if index <= 0 || index > len(userPets) {
			p.sendMessage(robot, event, common.T("", "pet_invalid_index"))
			return nil
		}

		newName := params[1]
		pet := userPets[index-1]
		oldName := pet.Name
		pet.Name = newName

		if err := db.UpdatePet(p.db, pet); err != nil {
			log.Printf(common.T("", "pet_rename_failed_log"), err)
			p.sendMessage(robot, event, common.T("", "pet_rename_failed"))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_rename_success"), oldName, newName))
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

		log.Printf(common.T("", "pet_levelup_log"), pet.Name, pet.Level)
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
			log.Printf(common.T("", "pet_cron_load_failed_log"), err)
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
				log.Printf(common.T("", "pet_cron_update_failed_log"), pet.PetID, err)
			}
		}
	}
}

// sendMessage 发送消息
func (p *PetPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	_, _ = SendTextReply(robot, event, msg)
}
