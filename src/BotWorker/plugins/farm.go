package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// FarmPlugin 开心农场插件
type FarmPlugin struct {
	db *sql.DB
	// 存储用户农场信息，key为用户ID，value为农场数据
	farms map[string]*Farm
	// 命令解析器
	cmdParser *CommandParser
}

// Farm 农场数据结构
type Farm struct {
	UserID    string
	Level     int
	Exp       int
	Coins     int
	Fields    [9]*Crop // 9块田地
	LastWater time.Time
}

// Crop 作物数据结构
type Crop struct {
	Type         string
	PlantTime    time.Time
	GrowthTime   time.Duration
	HarvestCoins int
	HarvestExp   int
}

// NewFarmPlugin 创建开心农场插件实例
func NewFarmPlugin(database *sql.DB) *FarmPlugin {
	rand.Seed(time.Now().UnixNano())
	return &FarmPlugin{
		db:        database,
		farms:     make(map[string]*Farm),
		cmdParser: NewCommandParser(),
	}
}

func (p *FarmPlugin) Name() string {
	return "farm"
}

func (p *FarmPlugin) Description() string {
	return common.T("", "farm_plugin_description|开心农场插件，支持种植、收获作物和农场升级")
}

func (p *FarmPlugin) Version() string {
	return "1.0.0"
}

func (p *FarmPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "farm_plugin_no_db|农场插件初始化失败：数据库未连接"))
		return
	}
	log.Println(common.T("", "farm_plugin_loading|农场插件正在加载..."))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理农场命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "farm") {
				HandleFeatureDisabled(robot, event, "farm")
				return nil
			}
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)

		// 检查是否为查看农场信息命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "farm_cmd_show|查看农场"), event.RawMessage); match {
			resp, err := p.doShowFarmInfo(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为种植命令
		match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "farm_cmd_plant|种植"), "(\\d+)\\s+(\\w+)", event.RawMessage)
		if match {
			if len(params) != 2 {
				p.sendMessage(robot, event, common.T("", "farm_plant_usage|种植命令用法：种植 <田地编号(1-9)> <作物名称>"))
				return nil
			}

			fieldIndex, err := strconv.Atoi(params[0])
			if err != nil || fieldIndex < 1 || fieldIndex > 9 {
				p.sendMessage(robot, event, common.T("", "farm_invalid_field|无效的田地编号。请输入1-9之间的数字。"))
				return nil
			}

			cropType := params[1]
			resp, err := p.doPlantCrop(userIDStr, fieldIndex-1, cropType)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为收获命令
		match, _, params = p.cmdParser.MatchCommandWithParams(common.T("", "farm_cmd_harvest|收获"), "(\\d+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, common.T("", "farm_harvest_usage|收获命令用法：收获 <田地编号(1-9)>"))
				return nil
			}

			fieldIndex, err := strconv.Atoi(params[0])
			if err != nil || fieldIndex < 1 || fieldIndex > 9 {
				p.sendMessage(robot, event, common.T("", "farm_invalid_field|无效的田地编号。请输入1-9之间的数字。"))
				return nil
			}

			resp, err := p.doHarvestCrop(userIDStr, fieldIndex-1)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为浇水命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "farm_cmd_water|浇水"), event.RawMessage); match {
			resp, err := p.doWaterCrops(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为购买升级命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "farm_cmd_upgrade|升级农场"), event.RawMessage); match {
			resp, err := p.doUpgradeFarm(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		return nil
	})
}

// GetSkills 报备插件技能
func (p *FarmPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "get_farm_info",
			Description: common.T("", "farm_skill_info_desc|获取用户的农场状态，包括等级、金币、经验和土地作物信息"),
			Usage:       "get_farm_info",
			Params:      map[string]string{},
		},
		{
			Name:        "plant_crop",
			Description: common.T("", "farm_skill_plant_desc|在指定的土地上种植指定的作物"),
			Usage:       "plant_crop field_index=1 crop_type=小麦",
			Params: map[string]string{
				"field_index": common.T("", "farm_skill_param_field_index|土地编号（1-9）"),
				"crop_type":   common.T("", "farm_skill_param_crop_type|作物类型（小麦、玉米、水稻、蔬菜、水果）"),
			},
		},
		{
			Name:        "harvest_crop",
			Description: common.T("", "farm_skill_harvest_desc|收获指定土地上已成熟的作物"),
			Usage:       "harvest_crop field_index=1",
			Params: map[string]string{
				"field_index": common.T("", "farm_skill_param_field_index|土地编号（1-9）"),
			},
		},
		{
			Name:        "water_crops",
			Description: common.T("", "farm_skill_water_desc|为所有正在生长的作物浇水，缩短10%的生长时间（每小时限一次）"),
			Usage:       "water_crops",
			Params:      map[string]string{},
		},
		{
			Name:        "upgrade_farm",
			Description: common.T("", "farm_skill_upgrade_desc|消耗金币提升农场等级"),
			Usage:       "upgrade_farm",
			Params:      map[string]string{},
		},
	}
}

// HandleSkill 处理技能调用
func (p *FarmPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userID := ""
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if uid, ok := params["user_id"]; ok {
		userID = uid
	}

	if userID == "" {
		return "", fmt.Errorf(common.T("", "farm_missing_user_id|缺少用户ID参数"))
	}

	switch skillName {
	case "get_farm_info":
		return p.doShowFarmInfo(userID)
	case "plant_crop":
		fieldIndex := 0
		if fiStr, ok := params["field_index"]; ok {
			fi, _ := strconv.Atoi(fiStr)
			fieldIndex = fi - 1
		} else {
			return "", fmt.Errorf(common.T("", "farm_missing_field_index|缺少土地编号参数"))
		}

		cropType, ok := params["crop_type"]
		if !ok {
			return "", fmt.Errorf(common.T("", "farm_missing_crop_type|缺少作物类型参数"))
		}
		return p.doPlantCrop(userID, fieldIndex, cropType)
	case "harvest_crop":
		fieldIndex := 0
		if fiStr, ok := params["field_index"]; ok {
			fi, _ := strconv.Atoi(fiStr)
			fieldIndex = fi - 1
		} else {
			return "", fmt.Errorf(common.T("", "farm_missing_field_index|缺少土地编号参数"))
		}
		return p.doHarvestCrop(userID, fieldIndex)
	case "water_crops":
		return p.doWaterCrops(userID)
	case "upgrade_farm":
		return p.doUpgradeFarm(userID)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doShowFarmInfo 显示农场信息逻辑
func (p *FarmPlugin) doShowFarmInfo(userIDStr string) (string, error) {
	farm := p.getOrCreateFarm(userIDStr)
	message := fmt.Sprintf(common.T("", "farm_info_title|=== 👨‍🌾 你的农场 (Lv.%d) ===\n"), farm.Level)
	message += fmt.Sprintf(common.T("", "farm_info_stats|💰 金币: %d | ✨ 经验: %d/%d\n"), farm.Coins, farm.Exp, farm.Level*100)
	message += common.T("", "farm_info_fields_title|🚜 土地状态：\n")

	for i, field := range farm.Fields {
		if field != nil {
			elapsed := time.Since(field.PlantTime)
			if elapsed >= field.GrowthTime {
				message += fmt.Sprintf(common.T("", "farm_field_harvestable|[%d] 🌾 %s (✅ 可收获)\n"), i+1, common.T("", "farm_crop_"+p.getCropKey(field.Type)+"|"+p.getCropChinese(field.Type)))
			} else {
				remaining := field.GrowthTime - elapsed
				message += fmt.Sprintf(common.T("", "farm_field_growing|[%d] 🌱 %s (⏳ 剩余 %.1f 分钟)\n"), i+1, common.T("", "farm_crop_"+p.getCropKey(field.Type)+"|"+p.getCropChinese(field.Type)), remaining.Minutes())
			}
		} else {
			message += fmt.Sprintf(common.T("", "farm_field_empty|[%d] 🕳️ 空闲\n"), i+1)
		}
	}

	return message, nil
}

// getCropChinese 获取作物的中文名称
func (p *FarmPlugin) getCropChinese(cropType string) string {
	switch cropType {
	case "小麦":
		return "小麦"
	case "玉米":
		return "玉米"
	case "水稻":
		return "水稻"
	case "蔬菜":
		return "蔬菜"
	case "水果":
		return "水果"
	default:
		return "未知"
	}
}

// getCropKey 获取作物对应的i18n key后缀
func (p *FarmPlugin) getCropKey(cropType string) string {
	switch cropType {
	case "小麦":
		return "wheat"
	case "玉米":
		return "corn"
	case "水稻":
		return "rice"
	case "蔬菜":
		return "vegetable"
	case "水果":
		return "fruit"
	default:
		return "unknown"
	}
}

// doPlantCrop 种植作物逻辑
func (p *FarmPlugin) doPlantCrop(userIDStr string, fieldIndex int, cropType string) (string, error) {
	farm := p.getOrCreateFarm(userIDStr)

	// 检查田地编号是否合法
	if fieldIndex < 0 || fieldIndex >= 9 {
		return common.T("", "farm_field_index_invalid|无效的土地编号，请输入1-9。"), nil
	}

	// 检查田地是否为空
	if farm.Fields[fieldIndex] != nil {
		return fmt.Sprintf(common.T("", "farm_field_occupied|土地 %d 已经种植了作物！"), fieldIndex+1), nil
	}

	// 作物配置
	cropConfig := map[string]struct {
		GrowthTime   time.Duration
		PlantCost    int
		HarvestCoins int
		HarvestExp   int
	}{
		"小麦": {30 * time.Minute, 50, 100, 10},
		"玉米": {60 * time.Minute, 100, 200, 20},
		"水稻": {90 * time.Minute, 150, 300, 30},
		"蔬菜": {45 * time.Minute, 75, 150, 15},
		"水果": {120 * time.Minute, 200, 400, 40},
	}

	config, exists := cropConfig[cropType]
	if !exists {
		return fmt.Sprintf(common.T("", "farm_crop_unknown|未知的作物类型：%s。可选：小麦、玉米、水稻、蔬菜、水果。"), cropType), nil
	}

	// 检查金币是否足够
	if farm.Coins < config.PlantCost {
		return fmt.Sprintf(common.T("", "farm_coins_insufficient_plant|金币不足！种植 %s 需要 %d 金币。"), cropType, config.PlantCost), nil
	}

	// 扣除金币
	farm.Coins -= config.PlantCost

	// 种植作物
	farm.Fields[fieldIndex] = &Crop{
		Type:         cropType,
		PlantTime:    time.Now(),
		GrowthTime:   config.GrowthTime,
		HarvestCoins: config.HarvestCoins,
		HarvestExp:   config.HarvestExp,
	}

	return fmt.Sprintf(common.T("", "farm_plant_success|成功在 %d 号土地种植了 %s！预计 %d 分钟后成熟。"), fieldIndex+1, cropType, int(config.GrowthTime.Minutes())), nil
}

// doHarvestCrop 收获作物逻辑
func (p *FarmPlugin) doHarvestCrop(userIDStr string, fieldIndex int) (string, error) {
	farm := p.getOrCreateFarm(userIDStr)

	// 检查田地编号是否合法
	if fieldIndex < 0 || fieldIndex >= 9 {
		return common.T("", "farm_field_index_invalid|无效的土地编号，请输入1-9。"), nil
	}

	// 检查田地是否有作物
	if farm.Fields[fieldIndex] == nil {
		return fmt.Sprintf(common.T("", "farm_field_no_crop|土地 %d 是空的，没有可以收获的作物。"), fieldIndex+1), nil
	}

	crop := farm.Fields[fieldIndex]

	// 检查作物是否成熟
	if time.Since(crop.PlantTime) < crop.GrowthTime {
		remaining := crop.GrowthTime - time.Since(crop.PlantTime)
		return fmt.Sprintf(common.T("", "farm_crop_not_mature|土地 %d 的 %s 还没有成熟，还需要 %d 分钟。"), fieldIndex+1, crop.Type, int(remaining.Minutes())), nil
	}

	// 收获作物
	farm.Coins += crop.HarvestCoins
	farm.Exp += crop.HarvestExp

	levelUpMsg := ""
	// 检查是否升级
	if farm.Exp >= farm.Level*100 {
		farm.Exp -= farm.Level * 100
		farm.Level++
		levelUpMsg = fmt.Sprintf(common.T("", "farm_level_up|\n🎊 恭喜！你的农场升级了！当前等级：Lv.%d"), farm.Level)
	}

	// 清空田地
	farm.Fields[fieldIndex] = nil

	return fmt.Sprintf(common.T("", "farm_harvest_success|成功收获了 %s！获得金币：%d，经验：%d。%s"), crop.Type, crop.HarvestCoins, crop.HarvestExp, levelUpMsg), nil
}

// doWaterCrops 浇水逻辑
func (p *FarmPlugin) doWaterCrops(userIDStr string) (string, error) {
	farm := p.getOrCreateFarm(userIDStr)

	// 检查浇水冷却时间（每小时只能浇水一次）
	if time.Since(farm.LastWater) < 1*time.Hour {
		remaining := 1*time.Hour - time.Since(farm.LastWater)
		return fmt.Sprintf(common.T("", "farm_water_cooldown|土地还很湿润，请在 %.1f 分钟后再浇水。"), remaining.Minutes()), nil
	}

	// 浇水（加速作物生长10%）
	wateredCount := 0
	for _, field := range farm.Fields {
		if field != nil {
			field.GrowthTime = field.GrowthTime * 9 / 10 // 减少10%生长时间
			wateredCount++
		}
	}

	if wateredCount == 0 {
		return common.T("", "farm_water_no_crops|农场里没有正在生长的作物，不需要浇水。"), nil
	}

	// 更新浇水时间
	farm.LastWater = time.Now()

	return fmt.Sprintf(common.T("", "farm_water_success|浇水成功！加速了 %d 处作物的生长。"), wateredCount), nil
}

// doUpgradeFarm 升级农场逻辑
func (p *FarmPlugin) doUpgradeFarm(userIDStr string) (string, error) {
	farm := p.getOrCreateFarm(userIDStr)

	upgradeCost := farm.Level * 500
	if farm.Coins < upgradeCost {
		return fmt.Sprintf(common.T("", "farm_coins_insufficient_upgrade|升级到 Lv.%d 需要 %d 金币，你的金币不足。"), farm.Level+1, upgradeCost), nil
	}

	// 扣除金币
	farm.Coins -= upgradeCost

	// 升级农场
	farm.Level++

	return fmt.Sprintf(common.T("", "farm_upgrade_success|农场升级成功！当前等级：Lv.%d"), farm.Level), nil
}

// sendMessage 发送消息
func (p *FarmPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "farm_send_failed_log|农场消息发送失败: %v"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "farm_send_failed_log|农场消息发送失败: %v"), err)
	}
}

// getOrCreateFarm 获取或创建用户农场信息
func (p *FarmPlugin) getOrCreateFarm(userID string) *Farm {
	if farm, ok := p.farms[userID]; ok {
		return farm
	}

	farm := &Farm{
		UserID:    userID,
		Level:     1,
		Coins:     500,
		LastWater: time.Now().Add(-1 * time.Hour),
	}
	p.farms[userID] = farm
	return farm
}
