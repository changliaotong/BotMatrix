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
	"time"
)

// FishingPlugin 钓鱼系统插件
type FishingPlugin struct {
	db *sql.DB
	// 存储用户上次钓鱼时间，key为用户ID，value为钓鱼时间
	lastFishingTime map[string]time.Time
	// 存储用户钓鱼技能等级，key为用户ID，value为等级
	fishingLevel map[string]int
	// 命令解析器
	cmdParser *CommandParser
}

// NewFishingPlugin 创建钓鱼系统插件实例
func NewFishingPlugin(database *sql.DB) *FishingPlugin {
	rand.Seed(time.Now().UnixNano())
	return &FishingPlugin{
		db:              database,
		lastFishingTime: make(map[string]time.Time),
		fishingLevel:    make(map[string]int),
		cmdParser:       NewCommandParser(),
	}
}

func (p *FishingPlugin) Name() string {
	return "fishing"
}

func (p *FishingPlugin) Description() string {
	return common.T("", "fishing_plugin_description|钓鱼系统插件，支持在水边垂钓获取积分，提升钓鱼等级。")
}

func (p *FishingPlugin) Version() string {
	return "1.0.0"
}

func (p *FishingPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "fishing_plugin_no_db|数据库连接未初始化，钓鱼插件无法正常运行"))
		return
	}
	log.Println(common.T("", "fishing_plugin_loading|钓鱼系统插件正在加载..."))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理钓鱼命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "fishing") {
				HandleFeatureDisabled(robot, event, "fishing")
				return nil
			}
		}

		// 检查是否为钓鱼命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fishing_cmd_fish|钓鱼|垂钓|go fishing"), event.RawMessage); match {
			userIDStr := fmt.Sprintf("%d", event.UserID)
			resp, err := p.doFish(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为查看钓鱼等级命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fishing_cmd_level|钓鱼等级|我的钓鱼等级"), event.RawMessage); match {
			userIDStr := fmt.Sprintf("%d", event.UserID)
			resp, err := p.doGetFishingStatus(userIDStr)
			if err != nil {
				return err
			}
			p.sendMessage(robot, event, resp)
			return nil
		}

		// 检查是否为查看钓鱼冷却命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "fishing_cmd_cooldown|钓鱼冷却|钓鱼等待时间"), event.RawMessage); match {
			userIDStr := fmt.Sprintf("%d", event.UserID)
			resp, err := p.doGetFishingCooldown(userIDStr)
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
func (p *FishingPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "start_fishing",
			Description: common.T("", "fishing_skill_start_desc|开始一次钓鱼活动"),
			Usage:       "start_fishing",
			Params:      map[string]string{},
		},
		{
			Name:        "get_fishing_status",
			Description: common.T("", "fishing_skill_status_desc|获取当前钓鱼等级和状态"),
			Usage:       "get_fishing_status",
			Params:      map[string]string{},
		},
		{
			Name:        "get_fishing_cooldown",
			Description: common.T("", "fishing_skill_cooldown_desc|查询下一次可以钓鱼的剩余时间"),
			Usage:       "get_fishing_cooldown",
			Params:      map[string]string{},
		},
	}
}

// HandleSkill 处理技能调用
func (p *FishingPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userID := ""
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if uid, ok := params["user_id"]; ok {
		userID = uid
	}

	if userID == "" {
		return common.T("", "fishing_missing_user_id|未找到用户ID，无法执行钓鱼操作"), nil
	}

	switch skillName {
	case "start_fishing":
		return p.doFish(userID)
	case "get_fishing_status":
		return p.doGetFishingStatus(userID)
	case "get_fishing_cooldown":
		return p.doGetFishingCooldown(userID)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doFish 钓鱼逻辑
func (p *FishingPlugin) doFish(userIDStr string) (string, error) {
	now := time.Now()

	// 检查钓鱼冷却时间（每10分钟只能钓鱼一次）
	if lastFishing, ok := p.lastFishingTime[userIDStr]; ok {
		if now.Sub(lastFishing) < 10*time.Minute {
			remainingTime := 10*time.Minute - now.Sub(lastFishing)
			return fmt.Sprintf(common.T("", "fishing_cooldown|你刚才钓得太累了，先休息一会儿吧。还需要等待 %.1f 分钟。"), remainingTime.Minutes()), nil
		}
	}

	// 获取用户钓鱼等级
	level := p.getFishingLevel(userIDStr)

	// 钓鱼成功率（根据等级提升）
	successRate := 0.5 + float64(level)*0.05
	if successRate > 0.95 {
		successRate = 0.95
	}

	// 判断是否钓鱼成功
	if rand.Float64() > successRate {
		// 钓鱼失败
		p.lastFishingTime[userIDStr] = now
		return common.T("", "fishing_fail|唉，鱼儿太聪明了，咬了钩又跑了。你一无所获。"), nil
	}

	// 钓鱼成功，随机获得积分
	basePoints := 10 + level*5
	bonusPoints := rand.Intn(20)
	totalPoints := basePoints + bonusPoints

	// 增加积分
	uid, _ := strconv.ParseInt(userIDStr, 10, 64)
	err := db.AddPoints(p.db, uid, totalPoints, common.T("", "fishing_points_reason|钓鱼获得奖励"), "fishing")
	if err != nil {
		return common.T("", "fishing_points_add_fail|糟糕，系统在为你发放积分时出错了。"), nil
	}

	// 提升钓鱼技能经验
	expGain := rand.Intn(5) + 1
	newLevel := p.addFishingExperience(userIDStr, expGain)

	// 更新钓鱼时间
	p.lastFishingTime[userIDStr] = now

	// 构建成功消息
	message := fmt.Sprintf(common.T("", "fishing_success|太棒了！你钓到了一条大鱼，获得了 %d 积分！"), totalPoints)
	if newLevel > level {
		message += "\n" + fmt.Sprintf(common.T("", "fishing_level_up|恭喜！你的钓鱼技术精进了，当前钓鱼等级提升至 Lv.%d！"), newLevel)
	}
	return message, nil
}

// doGetFishingStatus 查看钓鱼状态逻辑
func (p *FishingPlugin) doGetFishingStatus(userIDStr string) (string, error) {
	level := p.getFishingLevel(userIDStr)
	return fmt.Sprintf(common.T("", "fishing_status|你当前的钓鱼等级为：Lv.%d"), level), nil
}

// doGetFishingCooldown 查看钓鱼冷却逻辑
func (p *FishingPlugin) doGetFishingCooldown(userIDStr string) (string, error) {
	now := time.Now()

	if lastFishing, ok := p.lastFishingTime[userIDStr]; ok {
		if now.Sub(lastFishing) < 10*time.Minute {
			remainingTime := 10*time.Minute - now.Sub(lastFishing)
			return fmt.Sprintf(common.T("", "fishing_cooldown|你刚才钓得太累了，先休息一会儿吧。还需要等待 %.1f 分钟。"), remainingTime.Minutes()), nil
		} else {
			return common.T("", "fishing_cooldown_finished|你已经休息好了，随时可以再次挥杆垂钓！"), nil
		}
	} else {
		return common.T("", "fishing_never_fished|你还没有钓过鱼，快去水边试试运气吧！"), nil
	}
}

// getFishingLevel 获取用户钓鱼等级
func (p *FishingPlugin) getFishingLevel(userIDStr string) int {
	if level, ok := p.fishingLevel[userIDStr]; ok {
		return level
	}
	return 1 // 默认等级1
}

// addFishingExperience 增加钓鱼经验并提升等级
func (p *FishingPlugin) addFishingExperience(userIDStr string, exp int) int {
	level := p.getFishingLevel(userIDStr)
	expNeeded := level * 10 // 升级所需经验

	// 这里简化处理，实际应该存储经验值
	if rand.Intn(expNeeded) < exp {
		newLevel := level + 1
		p.fishingLevel[userIDStr] = newLevel
		return newLevel
	}

	return level
}

// sendMessage 发送消息
func (p *FishingPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "fishing_send_failed_log|发送钓鱼消息失败，机器人或事件为空"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "fishing_send_failed_log|发送钓鱼消息失败")+": %v", err)
	}
}
