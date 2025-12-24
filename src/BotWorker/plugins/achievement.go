package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// AchievementPlugin 成就系统插件
type AchievementPlugin struct {
	cmdParser   *CommandParser
	db          *gorm.DB
	redisClient *redis.Client
}

// Achievement 成就定义
type Achievement struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Description string
	Icon        string
	Points      int
	Condition   string
	Type        string
}

// UserAchievement 用户成就获得情况
type UserAchievement struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        string `gorm:"index"`
	AchievementID string `gorm:"index"`
	IsCompleted   bool
	Progress      int
	UnlockedAt    time.Time
}

func (p *AchievementPlugin) initDatabase() {
	if p.db == nil {
		return
	}
	p.db.AutoMigrate(&Achievement{}, &UserAchievement{})
}

func (p *AchievementPlugin) Name() string {
	return "achievement"
}

func (p *AchievementPlugin) Description() string {
	return common.T("", "achievement_plugin_desc|成就系统插件，记录并展示用户的各种成就和荣誉")
}

func (p *AchievementPlugin) Version() string {
	return "1.1.0"
}

// GetSkills 报备插件技能
func (p *AchievementPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "list_achievements",
			Description: common.T("", "achievement_skill_list_desc|查看所有成就列表"),
			Usage:       "list_achievements",
			Params:      map[string]string{},
		},
		{
			Name:        "my_achievements",
			Description: common.T("", "achievement_skill_my_desc|查看我的成就"),
			Usage:       "my_achievements user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "achievement_param_user_id|用户ID"),
			},
		},
		{
			Name:        "achievement_progress",
			Description: common.T("", "achievement_skill_progress_desc|查看我的成就进度"),
			Usage:       "achievement_progress user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "achievement_param_user_id|用户ID"),
			},
		},
		{
			Name:        "achievement_rank",
			Description: common.T("", "achievement_skill_rank_desc|查看成就排行榜"),
			Usage:       "achievement_rank",
			Params:      map[string]string{},
		},
	}
}

// NewAchievementPlugin 创建成就系统插件实例
func NewAchievementPlugin() *AchievementPlugin {
	return &AchievementPlugin{
		cmdParser: NewCommandParser(),
	}
}

// HandleSkill 实现 SkillCapable 接口
func (p *AchievementPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	userID := ""
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if uid, ok := params["user_id"]; ok {
		userID = uid
	}

	switch skillName {
	case "list_achievements":
		return p.doShowAllAchievements(), nil
	case "my_achievements":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "achievement_missing_user_id|缺少用户ID参数"))
		}
		return p.doShowMyAchievements(userID), nil
	case "achievement_progress":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "achievement_missing_user_id|缺少用户ID参数"))
		}
		return p.doShowAchievementProgress(userID), nil
	case "achievement_rank":
		return p.doShowAchievementRank(), nil
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

func (p *AchievementPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "achievement_plugin_loaded|成就系统插件加载成功"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 初始化数据库
	p.initDatabase()

	// 处理成就系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为成就命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "achievement_cmd|成就"), event.RawMessage); match {
			// 处理成就命令
			p.handleAchievementCommand(robot, event)
		}

		return nil
	})
}

// handleAchievementCommand 处理成就命令
func (p *AchievementPlugin) handleAchievementCommand(robot plugin.Robot, event *onebot.Event) {
	userIDStr := fmt.Sprintf("%d", event.UserID)

	// 检查命令参数
	args := p.cmdParser.ParseArgs(event.RawMessage)
	if len(args) == 1 {
		// 发送成就系统使用说明
		p.sendMessage(robot, event, common.T("", "achievement_usage|成就系统使用说明：\n- 成就 列表：查看所有成就\n- 成就 我的：查看已获得成就\n- 成就 进度：查看进行中成就\n- 成就 排行：查看成就点数排行"))
		return
	}

	// 处理子命令
	subCmd := args[1]
	if match, _ := p.cmdParser.MatchCommand(common.T("", "achievement_subcmd_list|列表"), subCmd); match {
		p.sendMessage(robot, event, p.doShowAllAchievements())
	} else if match, _ := p.cmdParser.MatchCommand(common.T("", "achievement_subcmd_my|我的"), subCmd); match {
		p.sendMessage(robot, event, p.doShowMyAchievements(userIDStr))
	} else if match, _ := p.cmdParser.MatchCommand(common.T("", "achievement_subcmd_progress|进度"), subCmd); match {
		p.sendMessage(robot, event, p.doShowAchievementProgress(userIDStr))
	} else if match, _ := p.cmdParser.MatchCommand(common.T("", "achievement_subcmd_rank|排行"), subCmd); match {
		p.sendMessage(robot, event, p.doShowAchievementRank())
	} else {
		p.sendMessage(robot, event, common.T("", "achievement_unknown_subcmd|未知的子命令。请输入'成就'查看使用说明。"))
	}
}

// doShowAllAchievements 显示所有成就
func (p *AchievementPlugin) doShowAllAchievements() string {
	if p.db == nil {
		return common.T("", "achievement_db_conn_failed|❌ 数据库连接失败")
	}
	var achievements []Achievement
	if err := p.db.Find(&achievements).Error; err != nil {
		log.Printf("[Achievement] %s: %v", common.T("", "achievement_query_list_failed_log|查询成就列表失败"), err)
		return common.T("", "achievement_query_list_failed|❌ 查询成就列表失败")
	}

	var msg string
	msg += common.T("", "achievement_list_title|🏆 所有成就列表") + "\n"
	msg += "====================\n\n"

	for _, achievement := range achievements {
		msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
		msg += fmt.Sprintf("📝 %s\n", achievement.Description)
		msg += fmt.Sprintf(common.T("", "achievement_reward_item|💰 奖励： %d 积分"), achievement.Points) + "\n\n"
	}

	if len(achievements) == 0 {
		msg += common.T("", "achievement_no_achievements|暂无任何成就数据")
	}

	return msg
}

// doShowMyAchievements 显示用户已获得的成就
func (p *AchievementPlugin) doShowMyAchievements(userID string) string {
	if p.db == nil {
		return common.T("", "achievement_db_conn_failed|❌ 数据库连接失败")
	}
	var userAchievements []UserAchievement
	if err := p.db.Where("user_id = ? AND is_completed = ?", userID, true).Find(&userAchievements).Error; err != nil {
		log.Printf("[Achievement] %s: %v", common.T("", "achievement_query_user_failed_log|查询用户成就失败"), err)
		return common.T("", "achievement_query_user_failed|❌ 查询用户成就失败")
	}

	var msg string
	msg += common.T("", "achievement_my_title|🏅 我的成就") + "\n"
	msg += "====================\n\n"

	for _, ua := range userAchievements {
		var achievement Achievement
		if err := p.db.First(&achievement, "id = ?", ua.AchievementID).Error; err == nil {
			msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
			msg += fmt.Sprintf(common.T("", "achievement_unlocked_at|🔓 解锁时间： %s"), ua.UnlockedAt.Format("2006-01-02 15:04:05")) + "\n\n"
		}
	}

	if len(userAchievements) == 0 {
		msg += common.T("", "achievement_no_unlocked|你还没有获得任何成就哦，继续努力吧！")
	}

	return msg
}

// doShowAchievementProgress 显示成就进度
func (p *AchievementPlugin) doShowAchievementProgress(userID string) string {
	if p.db == nil {
		return common.T("", "achievement_db_conn_failed|❌ 数据库连接失败")
	}
	var userAchievements []UserAchievement
	if err := p.db.Where("user_id = ? AND is_completed = ?", userID, false).Find(&userAchievements).Error; err != nil {
		log.Printf("[Achievement] %s: %v", common.T("", "achievement_query_progress_failed_log|查询成就进度失败"), err)
		return common.T("", "achievement_query_progress_failed|❌ 查询成就进度失败")
	}

	var msg string
	msg += common.T("", "achievement_progress_title|📈 成就进度") + "\n"
	msg += "====================\n\n"

	for _, ua := range userAchievements {
		var achievement Achievement
		if err := p.db.First(&achievement, "id = ?", ua.AchievementID).Error; err == nil {
			msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
			msg += fmt.Sprintf("📝 %s\n", achievement.Description)
			msg += fmt.Sprintf(common.T("", "achievement_progress_item|📊 当前进度： %d"), ua.Progress) + "\n\n"
		}
	}

	if len(userAchievements) == 0 {
		msg += common.T("", "achievement_no_in_progress|暂无进行中的成就")
	}

	return msg
}

// doShowAchievementRank 显示成就排行榜
func (p *AchievementPlugin) doShowAchievementRank() string {
	if p.db == nil {
		return common.T("", "achievement_db_conn_failed|❌ 数据库连接失败")
	}
	// 查询用户成就数量排行榜
	var rankData []struct {
		UserID string
		Count  int
	}

	query := `SELECT user_id, COUNT(*) as count FROM user_achievements WHERE is_completed = true GROUP BY user_id ORDER BY count DESC LIMIT 10`
	if err := p.db.Raw(query).Scan(&rankData).Error; err != nil {
		log.Printf("[Achievement] %s: %v", common.T("", "achievement_query_rank_failed_log|查询成就排行榜失败"), err)
		return common.T("", "achievement_query_rank_failed|❌ 查询成就排行榜失败")
	}

	var msg string
	msg += common.T("", "achievement_rank_title|📊 成就排行榜") + "\n"
	msg += "====================\n\n"

	for i, item := range rankData {
		msg += fmt.Sprintf(common.T("", "achievement_rank_item|第 %d 名： 用户 %s (成就数：%d)"), i+1, item.UserID, item.Count) + "\n"
	}

	if len(rankData) == 0 {
		msg += common.T("", "achievement_no_rank_data|暂无排行数据")
	}

	return msg
}

// sendMessage 发送消息
func (p *AchievementPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "achievement_send_failed_log|发送消息失败: %v"), err)
	}
}

// InitializeAchievements 初始化成就数据
func (p *AchievementPlugin) InitializeAchievements() error {
	achievements := []Achievement{
		{
			ID:          "first_sign_in",
			Name:        "首次签到",
			Description: "完成第一次签到",
			Icon:        "📅",
			Points:      100,
			Type:        "beginner",
			Condition:   "sign_in_count >= 1",
		},
		{
			ID:          "daily_streak",
			Name:        "连续签到",
			Description: "连续签到7天",
			Icon:        "🔥",
			Points:      500,
			Type:        "streak",
			Condition:   "sign_in_streak >= 7",
		},
		{
			ID:          "gift_master",
			Name:        "送礼达人",
			Description: "累计送出100份礼物",
			Icon:        "🎁",
			Points:      1000,
			Type:        "social",
			Condition:   "gift_sent_count >= 100",
		},
		{
			ID:          "love_owner",
			Name:        "群主真爱粉",
			Description: "累计发送100次爱群主",
			Icon:        "💖",
			Points:      2000,
			Type:        "social",
			Condition:   "love_owner_count >= 100",
		},
		{
			ID:          "chatty",
			Name:        "话唠之王",
			Description: "累计发送1000条消息",
			Icon:        "💬",
			Points:      5000,
			Type:        "activity",
			Condition:   "message_count >= 1000",
		},
	}

	for _, achievement := range achievements {
		if err := p.db.FirstOrCreate(&achievement, "id = ?", achievement.ID).Error; err != nil {
			return err
		}
	}

	return nil
}
