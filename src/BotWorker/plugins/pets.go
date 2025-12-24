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
	return common.T("", "pet_plugin_desc|宠物系统插件，提供宠物领养、喂养、玩耍等功能")
}

func (p *PetPlugin) Version() string {
	return "1.1.0"
}

// GetSkills 实现 SkillCapable 接口
func (p *PetPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "pet_adopt",
			Description: common.T("", "pet_skill_adopt_desc|领养一只可爱的宠物"),
			Usage:       "pet_adopt",
		},
		{
			Name:        "pet_list",
			Description: common.T("", "pet_skill_list_desc|查看你拥有的所有宠物"),
			Usage:       "pet_list",
		},
		{
			Name:        "pet_feed",
			Description: common.T("", "pet_skill_feed_desc|给你的宠物喂食"),
			Usage:       "pet_feed [pet_index]",
			Params: map[string]string{
				"pet_index": common.T("", "pet_skill_feed_param_index|宠物编号"),
			},
		},
		{
			Name:        "pet_play",
			Description: common.T("", "pet_skill_play_desc|和你的宠物一起玩耍"),
			Usage:       "pet_play [pet_index]",
			Params: map[string]string{
				"pet_index": common.T("", "pet_skill_play_param_index|宠物编号"),
			},
		},
		{
			Name:        "pet_wash",
			Description: common.T("", "pet_skill_wash_desc|给你的宠物洗澡"),
			Usage:       "pet_wash [pet_index]",
			Params: map[string]string{
				"pet_index": common.T("", "pet_skill_wash_param_index|宠物编号"),
			},
		},
		{
			Name:        "pet_rename",
			Description: common.T("", "pet_skill_rename_desc|给你的宠物改个新名字"),
			Usage:       "pet_rename <pet_index> <new_name>",
			Params: map[string]string{
				"pet_index": common.T("", "pet_skill_rename_param_index|宠物编号"),
				"new_name":  common.T("", "pet_skill_rename_param_name|新名字"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *PetPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) error {
	switch skillName {
	case "pet_adopt":
		return p.handleAdoptLogic(robot, event)
	case "pet_list":
		return p.handleListLogic(robot, event)
	case "pet_feed":
		petIndexStr := params["pet_index"]
		return p.handleFeedLogic(robot, event, petIndexStr)
	case "pet_play":
		petIndexStr := params["pet_index"]
		return p.handlePlayLogic(robot, event, petIndexStr)
	case "pet_wash":
		petIndexStr := params["pet_index"]
		return p.handleWashLogic(robot, event, petIndexStr)
	case "pet_rename":
		petIndexStr := params["pet_index"]
		newName := params["new_name"]
		if petIndexStr == "" || newName == "" {
			p.sendMessage(robot, event, common.T("", "pet_rename_usage_brief|使用方法: /改名 <宠物编号> <新名字>"))
			return nil
		}
		return p.handleRenameLogic(robot, event, petIndexStr, newName)
	}
	return nil
}

func (p *PetPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "pet_plugin_loaded|加载宠物系统插件"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return "", p.HandleSkill(robot, nil, skillName, params)
		})
	}

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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "pet_cmd_adopt|领养宠物"), event.RawMessage); !match {
			return nil
		}

		return p.handleAdoptLogic(robot, event)
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
		if match, _ := p.cmdParser.MatchCommand(common.T("", "pet_cmd_list|查看宠物"), event.RawMessage); !match {
			return nil
		}

		return p.handleListLogic(robot, event)
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_feed|喂食"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		petIndexStr := ""
		if len(params) > 0 {
			petIndexStr = params[0]
		}

		return p.handleFeedLogic(robot, event, petIndexStr)
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_play|玩耍"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		petIndexStr := ""
		if len(params) > 0 {
			petIndexStr = params[0]
		}

		return p.handlePlayLogic(robot, event, petIndexStr)
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_wash|洗澡"), `(\d*)`, event.RawMessage)
		if !match {
			return nil
		}

		petIndexStr := ""
		if len(params) > 0 {
			petIndexStr = params[0]
		}

		return p.handleWashLogic(robot, event, petIndexStr)
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
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "pet_cmd_rename|改名"), `(\d+)\s+(\S+)`, event.RawMessage)
		if !match {
			return nil
		}

		petIndexStr := params[0]
		newName := params[1]

		return p.handleRenameLogic(robot, event, petIndexStr, newName)
	})

	// 定时更新宠物状态（每小时）
	go p.updatePetStatus()
}

func (p *PetPlugin) handleAdoptLogic(robot plugin.Robot, event *onebot.Event) error {
	userID := event.UserID
	if userID == 0 {
		p.sendMessage(robot, event, common.T("", "pet_adopt_no_userid|❌ 无法获取用户信息，领养失败"))
		return nil
	}

	userIDStr := fmt.Sprintf("%d", userID)

	// 检查积分是否足够 (领养需要 50 积分)
	adoptCost := 50
	if p.points != nil && p.points.GetPoints(userIDStr) < adoptCost {
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_adopt_insufficient_points|❌ 领养宠物需要 %d 积分，你的积分不足"), adoptCost))
		return nil
	}

	// 检查用户是否已经有宠物 (从数据库查)
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil {
		log.Printf(common.T("", "pet_query_failed_log|[Pets] 查询用户宠物失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_query_failed|❌ 查询宠物信息失败，请稍后再试"))
		return nil
	}

	if len(userPets) >= 3 {
		p.sendMessage(robot, event, common.T("", "pet_adopt_limit|❌ 你已经领养了 3 只宠物，无法领养更多了"))
		return nil
	}

	// 随机生成宠物类型
	petTypes := []string{
		common.T("", "pet_type_cat|猫咪"),
		common.T("", "pet_type_dog|小狗"),
		common.T("", "pet_type_rabbit|兔子"),
		common.T("", "pet_type_hamster|仓鼠"),
		common.T("", "pet_type_bear|小熊"),
		common.T("", "pet_type_panda|熊猫"),
		common.T("", "pet_type_tiger|老虎"),
		common.T("", "pet_type_lion|狮子"),
	}
	petType := petTypes[rand.Intn(len(petTypes))]

	// 生成宠物ID
	petID := fmt.Sprintf("pet_%d_%d", time.Now().Unix(), userID)

	// 创建新宠物模型
	petModel := &db.PetModel{
		PetID:     petID,
		UserID:    userIDStr,
		Name:      fmt.Sprintf(common.T("", "pet_default_name|%d 的%s"), userID, petType),
		Type:      petType,
		Level:     1,
		Exp:       0,
		Hunger:    80,
		Happiness: 80,
		Health:    100,
	}

	// 存储宠物到数据库
	if err := db.CreatePet(p.db, petModel); err != nil {
		log.Printf(common.T("", "pet_save_failed_log|[Pets] 存储宠物失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_adopt_failed|❌ 领养失败，请稍后再试"))
		return nil
	}

	// 扣除积分
	if p.points != nil {
		p.points.AddPoints(userIDStr, -adoptCost, common.T("", "pet_adopt_action|领养宠物"), "pet_adopt")
	}

	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_adopt_success|🎉 领养成功！扣除 %d 积分\n宠物类型: %s\n名字: %s\n等级: %d\n经验: %d\n饥饿度: %d\n快乐度: %d\n健康度: %d"),
		adoptCost, petType, petModel.Name, petModel.Level, petModel.Exp, petModel.Hunger, petModel.Happiness, petModel.Health))

	return nil
}

func (p *PetPlugin) handleListLogic(robot plugin.Robot, event *onebot.Event) error {
	userID := event.UserID
	if userID == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_userid|❌ 无法获取用户信息"))
		return nil
	}

	// 获取用户的宠物
	userIDStr := fmt.Sprintf("%d", userID)
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil {
		log.Printf(common.T("", "pet_query_failed_log|[Pets] 查询用户宠物失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_query_failed_brief|❌ 查询失败"))
		return nil
	}

	if len(userPets) == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_pets|❌ 你还没有领养任何宠物，使用“领养宠物”来领养一只吧！"))
		return nil
	}

	// 发送宠物列表
	msg := common.T("", "pet_list_header|🐾 你的宠物列表:\n")
	msg += common.T("", "pet_list_separator|------------------\n")
	for i, pet := range userPets {
		msg += fmt.Sprintf("%d. %s\n", i+1, pet.Name)
		msg += fmt.Sprintf(common.T("", "pet_info_type|类型: %s\n"), pet.Type)
		msg += fmt.Sprintf(common.T("", "pet_info_level|等级: %d\n"), pet.Level)
		msg += fmt.Sprintf(common.T("", "pet_info_exp|经验: %d/%d\n"), pet.Exp, pet.Level*100)
		msg += fmt.Sprintf(common.T("", "pet_info_hunger|饥饿度: %d\n"), pet.Hunger)
		msg += fmt.Sprintf(common.T("", "pet_info_happiness|快乐度: %d\n"), pet.Happiness)
		msg += fmt.Sprintf(common.T("", "pet_info_health|健康度: %d\n"), pet.Health)
		msg += common.T("", "pet_list_separator|------------------\n")
	}

	p.sendMessage(robot, event, msg)

	return nil
}

func (p *PetPlugin) handleFeedLogic(robot plugin.Robot, event *onebot.Event, petIndexStr string) error {
	userID := event.UserID
	if userID == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_userid|❌ 无法获取用户信息"))
		return nil
	}

	userIDStr := fmt.Sprintf("%d", userID)

	// 喂食消耗 5 积分
	feedCost := 5
	if p.points != nil && p.points.GetPoints(userIDStr) < feedCost {
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_feed_insufficient_points|❌ 喂食需要 %d 积分，你的积分不足"), feedCost))
		return nil
	}

	// 获取用户的宠物
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil || len(userPets) == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_pets|❌ 你还没有领养任何宠物，使用“领养宠物”来领养一只吧！"))
		return nil
	}

	// 解析宠物编号
	petIndex := 0
	if petIndexStr != "" {
		index, err := strconv.Atoi(petIndexStr)
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
		log.Printf(common.T("", "pet_update_failed_log|[Pets] 更新宠物信息失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_op_failed|❌ 操作失败，请稍后再试"))
		return nil
	}

	// 扣除积分
	if p.points != nil {
		p.points.AddPoints(userIDStr, -feedCost, common.T("", "pet_feed_action|喂食宠物"), "pet_feed")
	}

	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_feed_success|🥣 你喂食了 %s，消耗了 %d 积分\n饥饿度: %d -> %d\n快乐度: %d -> %d\n经验: %d -> %d"),
		feedCost, pet.Name, oldHunger, pet.Hunger, oldHappiness, pet.Happiness, oldExp, pet.Exp))

	return nil
}

func (p *PetPlugin) handlePlayLogic(robot plugin.Robot, event *onebot.Event, petIndexStr string) error {
	userID := event.UserID
	if userID == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_userid|❌ 无法获取用户信息"))
		return nil
	}

	userIDStr := fmt.Sprintf("%d", userID)

	// 获取用户的宠物
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil || len(userPets) == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_pets|❌ 你还没有领养任何宠物，使用“领养宠物”来领养一只吧！"))
		return nil
	}

	// 解析宠物编号
	petIndex := 0
	if petIndexStr != "" {
		index, err := strconv.Atoi(petIndexStr)
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
		log.Printf(common.T("", "pet_update_failed_log|[Pets] 更新宠物信息失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_op_failed|❌ 操作失败，请稍后再试"))
		return nil
	}

	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_play_success|🎾 你和 %s 玩耍了一会\n快乐度: %d -> %d\n饥饿度: %d -> %d\n经验: %d -> %d"),
		pet.Name, oldHappiness, pet.Happiness, oldHunger, pet.Hunger, oldExp, pet.Exp))

	return nil
}

func (p *PetPlugin) handleWashLogic(robot plugin.Robot, event *onebot.Event, petIndexStr string) error {
	userID := event.UserID
	if userID == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_userid|❌ 无法获取用户信息"))
		return nil
	}

	userIDStr := fmt.Sprintf("%d", userID)

	// 获取用户的宠物
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil || len(userPets) == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_pets|❌ 你还没有领养任何宠物，使用“领养宠物”来领养一只吧！"))
		return nil
	}

	// 解析宠物编号
	petIndex := 0
	if petIndexStr != "" {
		index, err := strconv.Atoi(petIndexStr)
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
		log.Printf(common.T("", "pet_update_failed_log|[Pets] 更新宠物信息失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_op_failed|❌ 操作失败，请稍后再试"))
		return nil
	}

	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_wash_success|🧼 你给 %s 洗了个澡\n健康度: %d -> %d\n快乐度: %d -> %d\n经验: %d -> %d"),
		pet.Name, oldHealth, pet.Health, oldHappiness, pet.Happiness, oldExp, pet.Exp))

	return nil
}

func (p *PetPlugin) handleRenameLogic(robot plugin.Robot, event *onebot.Event, petIndexStr string, newName string) error {
	userID := event.UserID
	userIDStr := fmt.Sprintf("%d", userID)

	// 获取用户的宠物
	userPets, err := db.GetPetsByUserID(p.db, userIDStr)
	if err != nil || len(userPets) == 0 {
		p.sendMessage(robot, event, common.T("", "pet_no_pets_brief|❌ 你还没有领养任何宠物"))
		return nil
	}

	index, _ := strconv.Atoi(petIndexStr)
	if index <= 0 || index > len(userPets) {
		p.sendMessage(robot, event, common.T("", "pet_invalid_index|❌ 无效的宠物编号"))
		return nil
	}

	pet := userPets[index-1]
	oldName := pet.Name
	pet.Name = newName

	if err := db.UpdatePet(p.db, pet); err != nil {
		log.Printf(common.T("", "pet_rename_failed_log|[Pets] 重命名宠物失败: %v"), err)
		p.sendMessage(robot, event, common.T("", "pet_rename_failed|❌ 重命名失败，请稍后再试"))
		return nil
	}

	p.sendMessage(robot, event, fmt.Sprintf(common.T("", "pet_rename_success|✅ 成功将 %s 重命名为 %s"), oldName, newName))
	return nil
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

		log.Printf(common.T("", "pet_levelup_log|[Pets] 宠物 %s 升级了！当前等级: %d"), pet.Name, pet.Level)
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
			log.Printf(common.T("", "pet_cron_load_failed_log|[Pets] 定时任务加载宠物失败: %v"), err)
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
				log.Printf(common.T("", "pet_cron_update_failed_log|[Pets] 定时任务更新宠物 %s 失败: %v"), pet.PetID, err)
			}
		}
	}
}

// sendMessage 发送消息
func (p *PetPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if robot == nil || event == nil || msg == "" {
		return
	}
	_, _ = SendTextReply(robot, event, msg)
}
