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
	Type       string
	PlantTime  time.Time
	GrowthTime time.Duration
	HarvestCoins int
	HarvestExp  int
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
	return "开心农场插件，支持种植作物、收获作物、升级农场"
}

func (p *FarmPlugin) Version() string {
	return "1.0.0"
}

func (p *FarmPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("开心农场插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载开心农场插件")

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
		farm := p.getOrCreateFarm(userIDStr)

		// 检查是否为查看农场命令
		if match, _ := p.cmdParser.MatchCommand("农场|我的农场", event.RawMessage); match {
			p.showFarmInfo(robot, event, farm)
			return nil
		}

		// 检查是否为种植命令
		match, _, params := p.cmdParser.MatchCommandWithParams("种植", "(\d+)\s+(\w+)", event.RawMessage)
		if match {
			if len(params) != 2 {
				p.sendMessage(robot, event, "种植命令格式：种植 <田地编号> <作物名称>")
				return nil
			}

			fieldIndex, err := strconv.Atoi(params[0])
			if err != nil || fieldIndex < 1 || fieldIndex > 9 {
				p.sendMessage(robot, event, "田地编号必须在1-9之间")
				return nil
			}

			cropType := params[1]
			p.plantCrop(robot, event, farm, fieldIndex-1, cropType)
			return nil
		}

		// 检查是否为收获命令
		match, _, params = p.cmdParser.MatchCommandWithParams("收获", "(\d+)", event.RawMessage)
		if match {
			if len(params) != 1 {
				p.sendMessage(robot, event, "收获命令格式：收获 <田地编号>")
				return nil
			}

			fieldIndex, err := strconv.Atoi(params[0])
			if err != nil || fieldIndex < 1 || fieldIndex > 9 {
				p.sendMessage(robot, event, "田地编号必须在1-9之间")
				return nil
			}

			p.harvestCrop(robot, event, farm, fieldIndex-1)
			return nil
		}

		// 检查是否为浇水命令
		if match, _ := p.cmdParser.MatchCommand("浇水", event.RawMessage); match {
			p.waterCrops(robot, event, farm)
			return nil
		}

		// 检查是否为购买升级命令
		match, _, params = p.cmdParser.MatchCommandWithParams("升级农场", event.RawMessage)
		if match {
			p.upgradeFarm(robot, event, farm)
			return nil
		}

		return nil
	})
}

// getOrCreateFarm 获取或创建用户农场
func (p *FarmPlugin) getOrCreateFarm(userIDStr string) *Farm {
	if farm, ok := p.farms[userIDStr]; ok {
		return farm
	}

	// 创建新农场
	farm := &Farm{
		UserID:    userIDStr,
		Level:     1,
		Exp:       0,
		Coins:     1000,
		Fields:    [9]*Crop{},
		LastWater: time.Now(),
	}

	p.farms[userIDStr] = farm
	return farm
}

// showFarmInfo 显示农场信息
func (p *FarmPlugin) showFarmInfo(robot plugin.Robot, event *onebot.Event, farm *Farm) {
	message := fmt.Sprintf("🌾 开心农场 - 等级 %d\n", farm.Level)
	message += fmt.Sprintf("金币: %d | 经验: %d/%d\n", farm.Coins, farm.Exp, farm.Level*100)
	message += "田地状态：\n"

	for i, field := range farm.Fields {
		if field != nil {
			elapsed := time.Since(field.PlantTime)
			if elapsed >= field.GrowthTime {
				message += fmt.Sprintf("%d号地: %s (可收获)\n", i+1, field.Type)
			} else {
				remaining := field.GrowthTime - elapsed
				message += fmt.Sprintf("%d号地: %s (剩余 %.0f 分钟)\n", i+1, field.Type, remaining.Minutes())
			}
		} else {
			message += fmt.Sprintf("%d号地: 空\n", i+1)
		}
	}

	p.sendMessage(robot, event, message)
}

// plantCrop 种植作物
func (p *FarmPlugin) plantCrop(robot plugin.Robot, event *onebot.Event, farm *Farm, fieldIndex int, cropType string) {
	// 检查田地是否为空
	if farm.Fields[fieldIndex] != nil {
		p.sendMessage(robot, event, fmt.Sprintf("%d号地已经种植了作物，无法再次种植", fieldIndex+1))
		return
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
		p.sendMessage(robot, event, fmt.Sprintf("未知的作物类型：%s", cropType))
		return
	}

	// 检查金币是否足够
	if farm.Coins < config.PlantCost {
		p.sendMessage(robot, event, fmt.Sprintf("金币不足，种植%s需要%d金币", cropType, config.PlantCost))
		return
	}

	// 扣除金币
	farm.Coins -= config.PlantCost

	// 种植作物
	farm.Fields[fieldIndex] = &Crop{
		Type:       cropType,
		PlantTime:  time.Now(),
		GrowthTime: config.GrowthTime,
		HarvestCoins: config.HarvestCoins,
		HarvestExp: config.HarvestExp,
	}

	p.sendMessage(robot, event, fmt.Sprintf("🌱 成功在%d号地种植了%s\n需要%d分钟成熟", fieldIndex+1, cropType, int(config.GrowthTime.Minutes())))
}

// harvestCrop 收获作物
func (p *FarmPlugin) harvestCrop(robot plugin.Robot, event *onebot.Event, farm *Farm, fieldIndex int) {
	// 检查田地是否有作物
	if farm.Fields[fieldIndex] == nil {
		p.sendMessage(robot, event, fmt.Sprintf("%d号地没有种植作物，无法收获", fieldIndex+1))
		return
	}

	crop := farm.Fields[fieldIndex]

	// 检查作物是否成熟
	if time.Since(crop.PlantTime) < crop.GrowthTime {
		remaining := crop.GrowthTime - time.Since(crop.PlantTime)
		p.sendMessage(robot, event, fmt.Sprintf("%d号地的%s还未成熟，还需要%d分钟", fieldIndex+1, crop.Type, int(remaining.Minutes())))
		return
	}

	// 收获作物
	farm.Coins += crop.HarvestCoins
	farm.Exp += crop.HarvestExp

	// 检查是否升级
	if farm.Exp >= farm.Level*100 {
		farm.Exp -= farm.Level * 100
		farm.Level++
		p.sendMessage(robot, event, fmt.Sprintf("🎉 农场升级到%d级！\n", farm.Level))
	}

	// 清空田地
	farm.Fields[fieldIndex] = nil

	p.sendMessage(robot, event, fmt.Sprintf("💰 成功收获了%s\n获得%d金币和%d经验", crop.Type, crop.HarvestCoins, crop.HarvestExp))
}

// waterCrops 浇水
func (p *FarmPlugin) waterCrops(robot plugin.Robot, event *onebot.Event, farm *Farm) {
	// 检查浇水冷却时间（每小时只能浇水一次）
	if time.Since(farm.LastWater) < 1*time.Hour {
		remaining := 1*time.Hour - time.Since(farm.LastWater)
		p.sendMessage(robot, event, fmt.Sprintf("浇水冷却中，还需等待%.0f分钟", remaining.Minutes()))
		return
	}

	// 浇水（加速作物生长10%）
	wateredCount := 0
	for _, field := range farm.Fields {
		if field != nil {
			field.GrowthTime = field.GrowthTime * 9 / 10 // 减少10%生长时间
			wateredCount++
		}
	}

	// 更新浇水时间
	farm.LastWater = time.Now()

	p.sendMessage(robot, event, fmt.Sprintf("💧 浇水完成！为%d块田地的作物加速生长", wateredCount))
}

// upgradeFarm 升级农场
func (p *FarmPlugin) upgradeFarm(robot plugin.Robot, event *onebot.Event, farm *Farm) {
	upgradeCost := farm.Level * 500
	if farm.Coins < upgradeCost {
		p.sendMessage(robot, event, fmt.Sprintf("金币不足，升级到%d级需要%d金币", farm.Level+1, upgradeCost))
		return
	}

	// 扣除金币
	farm.Coins -= upgradeCost

	// 升级农场
	farm.Level++

	p.sendMessage(robot, event, fmt.Sprintf("🏠 农场升级成功！现在是%d级\n", farm.Level))
}

// sendMessage 发送消息
func (p *FarmPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}