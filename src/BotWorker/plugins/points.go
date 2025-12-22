package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

// PointsPlugin 积分系统插件
type PointsPlugin struct {
	db *sql.DB
	// 存储用户上次签到时间，key为用户ID，value为签到时间
	lastSignInTime map[string]time.Time
	// 存储用户上次领积分时间，key为用户ID，value为领积分时间
	lastGetPointsTime map[string]time.Time
	// 命令解析器
	cmdParser *CommandParser
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin(database *sql.DB) *PointsPlugin {
	return &PointsPlugin{
		db:                database,
		lastSignInTime:    make(map[string]time.Time),
		lastGetPointsTime: make(map[string]time.Time),
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
	if p.db == nil {
		log.Println("积分系统插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载积分系统插件")

	// 处理积分查询命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 检查是否为积分查询命令
		if match, _ := p.cmdParser.MatchCommand("points|积分", event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，查询失败")
			return nil
		}

		// 从数据库获取用户积分
		userIDStr := fmt.Sprintf("%d", userID)
		userPoints, err := db.GetPoints(p.db, userIDStr)
		if err != nil {
			log.Printf("获取积分失败: %v", err)
			p.sendMessage(robot, event, "查询积分失败，请稍后再试")
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("你当前的积分为：%d", userPoints))
		return nil
	})

	// 处理签到积分命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 检查是否为签到命令
		match, msg := p.cmdParser.MatchCommand("signpoints|签到积分|签到|早安|晚安", event.RawMessage)
		if !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，签到失败")
			return nil
		}

		// 检查是否已经签到
		now := time.Now()
		userIDStr := fmt.Sprintf("%d", userID)
		if lastSignIn, ok := p.lastSignInTime[userIDStr]; ok {
			// 检查是否在同一天
			if isSameDay(lastSignIn, now) {
				p.sendMessage(robot, event, fmt.Sprintf("你今天已经签到过了！上次签到时间：%s", lastSignIn.Format("15:04:05")))
				return nil
			}
		}

		// 增加积分（签到奖励10积分）
		err := db.AddPoints(p.db, userIDStr, 10, "签到奖励", "sign_in")
		if err != nil {
			log.Printf("签到积分增加失败: %v", err)
			p.sendMessage(robot, event, "签到失败，请稍后再试")
			return nil
		}
		p.lastSignInTime[userIDStr] = now

		// 获取更新后的积分
		userPoints, _ := db.GetPoints(p.db, userIDStr)

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

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			return nil
		}

		// 检查是否为命令消息（不奖励积分）
		if p.cmdParser.IsCommand("points|积分|signpoints|签到积分|签到|早安|晚安|rank|排行榜|积分榜|打赏|reward|转账|transfer|领积分|getpoints|存积分|存款|取积分|取款|冻结积分|冻结|解冻积分|解冻", event.RawMessage) {
			return nil
		}

		// 发言奖励1积分
		userIDStr := fmt.Sprintf("%d", userID)
		_ = db.AddPoints(p.db, userIDStr, 1, "发言奖励", "message_reward")

		return nil
	})

	// 处理积分排行榜命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 检查是否为排行榜命令
		if match, _ := p.cmdParser.MatchCommand("rank|排行榜|积分榜", event.RawMessage); !match {
			return nil
		}

		// 从数据库获取积分排行榜
		rank, err := p.getPointsRankFromDB()
		if err != nil {
			log.Printf("获取积分排行榜失败: %v", err)
			p.sendMessage(robot, event, "获取排行榜失败")
			return nil
		}

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

		p.sendMessage(robot, event, msg)
		return nil
	})

	// 处理打赏/转账功能
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 检查是否为打赏或转账命令
		match, cmd, params := p.cmdParser.MatchCommandWithParams("打赏|reward|转账|transfer", "(\\d+)\\s+(\\d+)", event.RawMessage)
		if !match || len(params) != 2 {
			if match {
				p.sendMessage(robot, event, fmt.Sprintf("%s命令格式：%s <用户ID> <积分数量>", cmd, cmd))
			}
			return nil
		}

		// 解析转账信息
		toUserID := params[0]
		pointsStr := params[1]
		points, err := strconv.Atoi(pointsStr)
		if err != nil || points <= 0 {
			p.sendMessage(robot, event, "积分数量必须为正整数")
			return nil
		}

		// 获取操作者ID
		fromUserID := event.UserID
		fromUserIDStr := fmt.Sprintf("%d", fromUserID)

		if fromUserIDStr == toUserID {
			p.sendMessage(robot, event, "不能给自己转账哦")
			return nil
		}

		// 执行转账（使用数据库事务）
		reason := "主动转账"
		if cmd == "打赏" || cmd == "reward" {
			reason = "打赏"
		}

		err = db.TransferPoints(p.db, fromUserIDStr, toUserID, points, reason, "transfer")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("操作失败: %v", err))
			return nil
		}

		// 发送成功消息
		p.sendMessage(robot, event, fmt.Sprintf("✅ %s成功！你给用户 %s %s了 %d 积分", reason, toUserID, reason, points))
		return nil
	})

	// 处理领积分功能
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				HandleFeatureDisabled(robot, event, "points")
				return nil
			}
		}

		// 检查是否为领积分命令
		if match, _ := p.cmdParser.MatchCommand("领积分|getpoints", event.RawMessage); !match {
			return nil
		}

		// 获取用户ID
		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，领积分失败")
			return nil
		}

		// 检查是否已经领取过
		userIDStr := fmt.Sprintf("%d", userID)
		lastGetTime, ok := p.lastGetPointsTime[userIDStr]
		now := time.Now()
		if ok && isSameDay(lastGetTime, now) {
			p.sendMessage(robot, event, "你今天已经领取过积分了！")
			return nil
		}

		// 领取5积分
		err := db.AddPoints(p.db, userIDStr, 5, "每日领积分", "daily_bonus")
		if err != nil {
			p.sendMessage(robot, event, "领取失败，请稍后再试")
			return nil
		}
		p.lastGetPointsTime[userIDStr] = now

		// 获取更新后的积分
		userPoints, _ := db.GetPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf("领取成功！获得5积分\n当前积分：%d", userPoints))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				return nil
			}
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，存积分失败")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		matchDep, _, depParams := p.cmdParser.MatchCommandWithParams("存积分|存款", `(\\d+)`, event.RawMessage)
		if matchDep && len(depParams) == 1 {
			amount, err := strconv.Atoi(depParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, "存入的积分数量必须为正整数")
				return nil
			}

			err = db.DepositPointsToSavings(p.db, userIDStr, amount)
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf("存积分失败: %v", err))
				return nil
			}

			saving, _ := db.GetSavingsPoints(p.db, userIDStr)
			p.sendMessage(robot, event, fmt.Sprintf("已存入 %d 积分\n当前存积分余额：%d", amount, saving))
			return nil
		}

		matchQuery, _ := p.cmdParser.MatchCommand("存积分|存款", event.RawMessage)
		if !matchQuery {
			return nil
		}

		saving, err := db.GetSavingsPoints(p.db, userIDStr)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("查询存积分失败: %v", err))
			return nil
		}

		points, err := db.GetPoints(p.db, userIDStr)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("查询积分失败: %v", err))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf("当前可用积分：%d\n当前存积分余额：%d", points, saving))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				return nil
			}
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，取积分失败")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("取积分|取款", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, "取出的积分数量必须为正整数")
			return nil
		}

		err = db.WithdrawPointsFromSavings(p.db, userIDStr, amount)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("取积分失败: %v", err))
			return nil
		}

		saving, _ := db.GetSavingsPoints(p.db, userIDStr)
		points, _ := db.GetPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf("已取出 %d 积分\n当前可用积分：%d\n当前存积分余额：%d", amount, points, saving))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "points") {
				return nil
			}
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，冻结积分失败")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("冻结积分|冻结", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, "冻结的积分数量必须为正整数")
			return nil
		}

		err = db.FreezePoints(p.db, userIDStr, amount, "手动冻结积分")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("冻结积分失败: %v", err))
			return nil
		}

		frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf("已冻结 %d 积分\n当前冻结积分：%d", amount, frozen))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, "无法获取用户ID，解冻积分失败")
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("解冻积分|解冻", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, "解冻的积分数量必须为正整数")
			return nil
		}

		err = db.UnfreezePoints(p.db, userIDStr, amount, "手动解冻积分")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf("解冻积分失败: %v", err))
			return nil
		}

		frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf("已解冻 %d 积分\n当前冻结积分：%d", amount, frozen))

		return nil
	})
}

// sendMessage 发送消息
func (p *PointsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

func (p *PointsPlugin) AddPoints(userID string, points int, reason string, category string) {
	if p.db == nil {
		return
	}
	_ = db.AddPoints(p.db, userID, points, reason, category)
}

func (p *PointsPlugin) GetPoints(userID string) int {
	if p.db == nil {
		return 0
	}
	points, err := db.GetPoints(p.db, userID)
	if err != nil {
		return 0
	}
	return points
}

type PointsRankItem struct {
	UserID string
	Points int
}

func (p *PointsPlugin) getPointsRankFromDB() ([]PointsRankItem, error) {
	if p.db == nil {
		return nil, nil
	}

	rows, err := p.db.Query("SELECT user_id, points FROM users ORDER BY points DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rank []PointsRankItem
	for rows.Next() {
		var item PointsRankItem
		if err := rows.Scan(&item.UserID, &item.Points); err != nil {
			return nil, err
		}
		rank = append(rank, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rank, nil
}
