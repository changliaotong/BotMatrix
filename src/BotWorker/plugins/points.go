package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"strconv"
	"time"
)

// PointsPlugin 积分系统插件
type PointsPlugin struct {
	// 存储用户积分，key为用户ID，value为积分数量
	points map[string]int
	// 存储用户上次签到时间，key为用户ID，value为签到时间
	lastSignInTime map[string]time.Time
	// 存储用户上次领积分时间，key为用户ID，value为领积分时间
	lastGetPointsTime map[string]time.Time
	// 存储用户积分记录，key为用户ID，value为积分记录列表
	pointsRecords map[string][]PointsRecord
	// 命令解析器
	cmdParser *CommandParser
}

// PointsRecord 积分记录
type PointsRecord struct {
	Points    int       // 积分数量
	Reason    string    // 积分变动原因
	Timestamp time.Time // 变动时间
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin() *PointsPlugin {
	return &PointsPlugin{
		points:            make(map[string]int),
		lastSignInTime:    make(map[string]time.Time),
		lastGetPointsTime: make(map[string]time.Time),
		pointsRecords:     make(map[string][]PointsRecord),
		cmdParser:         NewCommandParser(),
	}
}

func (p *PointsPlugin) Name() string {
	return "points"
}

func (p *PointsPlugin) Description() string {
	return "积分系统插件，支持签到积分、发言积分、查询积分等功能"
}

func (p *PointsPlugin) Version() string {
	return "1.0.0"
}

func (p *PointsPlugin) Init(robot plugin.Robot) {
	log.Println("加载积分系统插件")

	// 处理积分查询命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为积分查询命令
		if match, _ := p.cmdParser.MatchCommand("points|积分", event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，查询失败")
			return nil
		}

		// 获取用户积分
		userPoints := p.points[userID]
		if userPoints == 0 {
			p.sendMessage(robot, event, fmt.Sprintf("你当前的积分为：0"))
		} else {
			p.sendMessage(robot, event, fmt.Sprintf("你当前的积分为：%d", userPoints))
		}

		return nil
	})

	// 处理签到积分命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为签到命令
		match, msg := p.cmdParser.MatchCommand("signpoints|签到积分|签到|早安|晚安", event.RawMessage)
		if !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，签到失败")
			return nil
		}

		// 检查是否已经签到
		now := time.Now()
		if lastSignIn, ok := p.lastSignInTime[userID]; ok {
			// 检查是否在同一天
			if isSameDay(lastSignIn, now) {
				p.sendMessage(robot, event, fmt.Sprintf("你今天已经签到过了！上次签到时间：%s", lastSignIn.Format("15:04:05")))
				return nil
			}
		}

		// 增加积分（签到奖励10积分）
		p.addPoints(userID, 10, "签到奖励")
		p.lastSignInTime[userID] = now

		// 发送签到成功消息
		userPoints := p.points[userID]
		var rewardMsg string
		switch msg {
		case "早安":
			rewardMsg = fmt.Sprintf("☀️ 早安！签到成功！获得10积分\n当前积分：%d", userPoints)
		case "晚安":
			rewardMsg = fmt.Sprintf("🌙 晚安！签到成功！获得10积分\n当前积分：%d", userPoints)
		default:
			rewardMsg = fmt.Sprintf("签到成功！获得10积分\n当前积分：%d", userPoints)
		}
		p.sendMessage(robot, event, rewardMsg)

		return nil
	})

	// 处理发言积分
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			return nil
		}

		// 检查是否为命令消息（不奖励积分）
		// 检查所有插件的命令模式
		if p.cmdParser.IsCommand("points|积分|signpoints|签到积分|签到|早安|晚安|rank|排行榜|积分榜|打赏|reward|领积分|getpoints", event.RawMessage) {
			return nil
		}

		// 发言奖励1积分
		p.addPoints(userID, 1, "发言奖励")

		return nil
	})

	// 处理积分排行榜命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为排行榜命令
		if match, _ := p.cmdParser.MatchCommand("rank|排行榜|积分榜", event.RawMessage); !match {
			return nil
		}

		// 获取积分排行榜
		rank := p.getPointsRank()

		// 发送排行榜消息
		if len(rank) == 0 {
			p.sendMessage(robot, event, "暂无积分记录")
			return nil
		}

		msg := "🏆 积分排行榜 🏆\n"
		msg += "------------------------\n"
		for i, item := range rank {
			var medal string
			switch i {
			case 0:
				medal = "🥇"
			case 1:
				medal = "🥈"
			case 2:
				medal = "🥉"
			default:
				medal = fmt.Sprintf("%d.", i+1)
			}
			msg += fmt.Sprintf("%s 用户%s：%d积分\n", medal, item.UserID, item.Points)
		}
		msg += "------------------------\n"
		msg += fmt.Sprintf("总参与人数：%d人\n", len(p.points))

		p.sendMessage(robot, event, msg)

		return nil
	})

	// 处理打赏功能
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为打赏命令
		match, _, params := p.cmdParser.MatchCommandWithParams("打赏|reward", "(\\S+)\\s+(\\S+)", event.RawMessage)
		if !match || len(params) != 2 {
			p.sendMessage(robot, event, "打赏命令格式：/打赏 <用户ID> <积分数量>")
			return nil
		}

		// 解析打赏信息
		toUserID := params[0]
		pointsStr := params[1]
		points, err := strconv.Atoi(pointsStr)
		if err != nil || points <= 0 {
			p.sendMessage(robot, event, "积分数量必须为正整数")
			return nil
		}

		// 获取打赏者ID
		fromUserID := event.UserID
		if fromUserID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，打赏失败")
			return nil
		}

		// 检查打赏者积分是否足够
		if p.points[fromUserID] < points {
			p.sendMessage(robot, event, "积分不足，打赏失败")
			return nil
		}

		// 执行打赏
		p.addPoints(fromUserID, -points, fmt.Sprintf("打赏用户%s", toUserID))
		p.addPoints(toUserID, points, fmt.Sprintf("收到用户%s打赏", fromUserID))

		// 发送打赏成功消息
		rewardMsg := fmt.Sprintf("打赏成功！用户%s 打赏用户%s %d积分", fromUserID, toUserID, points)
		p.sendMessage(robot, event, rewardMsg)

		return nil
	})

	// 处理领积分功能
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为领积分命令
		if match, _ := p.cmdParser.MatchCommand("领积分|getpoints", event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == "" {
			p.sendMessage(robot, event, "无法获取用户ID，领积分失败")
			return nil
		}

		// 检查是否已经领取过
		lastGetTime, ok := p.lastGetPointsTime[userID]
		now := time.Now()
		if ok && isSameDay(lastGetTime, now) {
			p.sendMessage(robot, event, "你今天已经领取过积分了！")
			return nil
		}

		// 领取5积分
		p.addPoints(userID, 5, "每日领积分")
		p.lastGetPointsTime[userID] = now

		// 发送领取成功消息
		userPoints := p.points[userID]
		msg := fmt.Sprintf("领取成功！获得5积分\n当前积分：%d", userPoints)
		p.sendMessage(robot, event, msg)

		return nil
	})
}

// sendMessage 发送消息
func (p *PointsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// addPoints 增加用户积分
func (p *PointsPlugin) addPoints(userID string, points int, reason string) {
	// 增加积分
	p.points[userID] += points

	// 记录积分变动
	record := PointsRecord{
		Points:    points,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	p.pointsRecords[userID] = append(p.pointsRecords[userID], record)
}

// getPointsRank 获取积分排行榜
func (p *PointsPlugin) getPointsRank() []PointsRankItem {
	// 转换为排行榜项列表
	var rank []PointsRankItem
	for userID, points := range p.points {
		if points > 0 {
			rank = append(rank, PointsRankItem{UserID: userID, Points: points})
		}
	}

	// 按积分降序排序
	for i := 0; i < len(rank); i++ {
		for j := i + 1; j < len(rank); j++ {
			if rank[j].Points > rank[i].Points {
				rank[i], rank[j] = rank[j], rank[i]
			}
		}
	}

	// 返回前10名
	if len(rank) > 10 {
		return rank[:10]
	}
	return rank
}

// PointsRankItem 排行榜项
type PointsRankItem struct {
	UserID string // 用户ID
	Points int    // 积分数量
}

// isSameDay 检查两个时间是否在同一天
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
