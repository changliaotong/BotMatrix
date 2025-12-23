package plugins

import (
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
	cmdParser *CommandParser
	db        *gorm.DB
	redisClient *redis.Client
}

// Achievement 成就结构体
type Achievement struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Points      int       `json:"points"`
	Type        string    `json:"type"`
	Condition   string    `json:"condition"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserAchievement 用户成就结构体
type UserAchievement struct {
	UserID        string    `json:"user_id" gorm:"primaryKey"`
	AchievementID string    `json:"achievement_id" gorm:"primaryKey"`
	UnlockedAt    time.Time `json:"unlocked_at"`
	Progress      int       `json:"progress"`
	IsCompleted   bool      `json:"is_completed"`
}

func (p *AchievementPlugin) Name() string {
	return "achievement"
}

func (p *AchievementPlugin) Description() string {
	return "成就系统插件，管理用户成就"
}

func (p *AchievementPlugin) Version() string {
	return "1.0.0"
}

// NewAchievementPlugin 创建成就系统插件实例
func NewAchievementPlugin() *AchievementPlugin {
	return &AchievementPlugin{
		cmdParser: NewCommandParser(),
	}
}

func (p *AchievementPlugin) Init(robot plugin.Robot) {
	log.Println("加载成就系统插件")

	// 处理成就系统命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为成就命令
		if match, _ := p.cmdParser.MatchCommand("成就|achievement|achieve", event.RawMessage); match {
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
		usage := "🏆 成就系统命令使用说明:\n"
		usage += "====================\n"
		usage += "/成就 列表 - 查看所有成就\n"
		usage += "/成就 我的 - 查看已获得的成就\n"
		usage += "/成就 进度 - 查看成就进度\n"
		usage += "/成就 排行 - 查看成就排行榜\n"
		p.sendMessage(robot, event, usage)
		return
	}

	// 处理子命令
	subCmd := args[1]
	switch subCmd {
	case "列表", "list":
		p.showAllAchievements(robot, event)
	case "我的", "my":
		p.showMyAchievements(robot, event, userIDStr)
	case "进度", "progress":
		p.showAchievementProgress(robot, event, userIDStr)
	case "排行", "rank":
		p.showAchievementRank(robot, event)
	default:
		p.sendMessage(robot, event, "❌ 未知子命令，请使用/成就查看帮助")
	}
}

// showAllAchievements 显示所有成就
func (p *AchievementPlugin) showAllAchievements(robot plugin.Robot, event *onebot.Event) {
	var achievements []Achievement
	if err := p.db.Find(&achievements).Error; err != nil {
		log.Printf("[Achievement] 查询成就列表失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询成就列表失败")
		return
	}

	var msg string
	msg += "🏆 所有成就列表:\n"
	msg += "====================\n\n"

	for _, achievement := range achievements {
		msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
		msg += fmt.Sprintf("📝 %s\n", achievement.Description)
		msg += fmt.Sprintf("💎 奖励: %d 积分\n\n", achievement.Points)
	}

	if len(achievements) == 0 {
		msg += "暂无成就"
	}

	p.sendMessage(robot, event, msg)
}

// showMyAchievements 显示用户已获得的成就
func (p *AchievementPlugin) showMyAchievements(robot plugin.Robot, event *onebot.Event, userID string) {
	var userAchievements []UserAchievement
	if err := p.db.Where("user_id = ? AND is_completed = ?", userID, true).Find(&userAchievements).Error; err != nil {
		log.Printf("[Achievement] 查询用户成就失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询用户成就失败")
		return
	}

	var msg string
	msg += "🏆 我的成就:\n"
	msg += "====================\n\n"

	for _, ua := range userAchievements {
		var achievement Achievement
		if err := p.db.First(&achievement, "id = ?", ua.AchievementID).Error; err == nil {
			msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
			msg += fmt.Sprintf("📅 获得时间: %s\n\n", ua.UnlockedAt.Format("2006-01-02 15:04:05"))
		}
	}

	if len(userAchievements) == 0 {
		msg += "暂无获得的成就"
	}

	p.sendMessage(robot, event, msg)
}

// showAchievementProgress 显示成就进度
func (p *AchievementPlugin) showAchievementProgress(robot plugin.Robot, event *onebot.Event, userID string) {
	var userAchievements []UserAchievement
	if err := p.db.Where("user_id = ? AND is_completed = ?", userID, false).Find(&userAchievements).Error; err != nil {
		log.Printf("[Achievement] 查询成就进度失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询成就进度失败")
		return
	}

	var msg string
	msg += "📊 成就进度:\n"
	msg += "====================\n\n"

	for _, ua := range userAchievements {
		var achievement Achievement
		if err := p.db.First(&achievement, "id = ?", ua.AchievementID).Error; err == nil {
			msg += fmt.Sprintf("%s %s\n", achievement.Icon, achievement.Name)
			msg += fmt.Sprintf("📝 %s\n", achievement.Description)
			msg += fmt.Sprintf("📊 进度: %d%%\n\n", ua.Progress)
		}
	}

	if len(userAchievements) == 0 {
		msg += "暂无进行中的成就"
	}

	p.sendMessage(robot, event, msg)
}

// showAchievementRank 显示成就排行榜
func (p *AchievementPlugin) showAchievementRank(robot plugin.Robot, event *onebot.Event) {
	// 查询用户成就数量排行榜
	var rankData []struct {
		UserID string
		Count  int
	}

	query := `SELECT user_id, COUNT(*) as count FROM user_achievements WHERE is_completed = true GROUP BY user_id ORDER BY count DESC LIMIT 10`
	if err := p.db.Raw(query).Scan(&rankData).Error; err != nil {
		log.Printf("[Achievement] 查询成就排行失败: %v", err)
		p.sendMessage(robot, event, "❌ 查询成就排行失败")
		return
	}

	var msg string
	msg += "🏆 成就排行榜:\n"
	msg += "====================\n\n"

	for i, item := range rankData {
		msg += fmt.Sprintf("%d. 用户 %s: %d 个成就\n", i+1, item.UserID, item.Count)
	}

	if len(rankData) == 0 {
		msg += "暂无成就数据"
	}

	p.sendMessage(robot, event, msg)
}

// sendMessage 发送消息
func (p *AchievementPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
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