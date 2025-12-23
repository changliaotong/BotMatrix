package plugins

import (
	"BotMatrix/common"
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
	return common.T("", "points_plugin_desc")
}

func (p *PointsPlugin) Version() string {
	return "1.0.0"
}

func (p *PointsPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "points_db_not_configured"))
		return
	}
	log.Println(common.T("", "points_plugin_loaded"))

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
			p.sendMessage(robot, event, common.T("", "points_query_no_userid"))
			return nil
		}

		// 从数据库获取用户积分
		userIDStr := fmt.Sprintf("%d", userID)
		userPoints, err := db.GetPoints(p.db, userIDStr)
		if err != nil {
			log.Printf(common.T("", "points_query_log_failed")+": %v", err)
			p.sendMessage(robot, event, common.T("", "points_query_failed"))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_current_balance"), userPoints))
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
			p.sendMessage(robot, event, common.T("", "points_sign_no_userid"))
			return nil
		}

		// 检查是否已经签到
		now := time.Now()
		userIDStr := fmt.Sprintf("%d", userID)
		if lastSignIn, ok := p.lastSignInTime[userIDStr]; ok {
			// 检查是否在同一天
			if isSameDay(lastSignIn, now) {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_sign_already"), lastSignIn.Format("15:04:05")))
				return nil
			}
		}

		// 增加积分（签到奖励10积分）
		err := db.AddPoints(p.db, userIDStr, 10, common.T("", "points_reason_signin"), "sign_in")
		if err != nil {
			log.Printf(common.T("", "points_sign_log_failed")+": %v", err)
			p.sendMessage(robot, event, common.T("", "points_sign_failed"))
			return nil
		}
		p.lastSignInTime[userIDStr] = now

		// 获取更新后的积分
		userPoints, _ := db.GetPoints(p.db, userIDStr)

		var rewardMsg string
		switch msg {
		case "早安":
			rewardMsg = fmt.Sprintf(common.T("", "points_sign_morning"), userPoints)
		case "晚安":
			rewardMsg = fmt.Sprintf(common.T("", "points_sign_night"), userPoints)
		default:
			rewardMsg = fmt.Sprintf(common.T("", "points_sign_success"), userPoints)
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
		_ = db.AddPoints(p.db, userIDStr, 1, common.T("", "points_reason_message"), "message_reward")

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
			log.Printf(common.T("", "points_rank_log_failed")+": %v", err)
			p.sendMessage(robot, event, common.T("", "points_rank_failed"))
			return nil
		}

		if len(rank) == 0 {
			p.sendMessage(robot, event, common.T("", "points_rank_empty"))
			return nil
		}

		msg := common.T("", "points_rank_title") + "\n"
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
			msg += fmt.Sprintf(common.T("", "points_rank_item"), medal, item.UserID, item.Points) + "\n"
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
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_transfer_usage"), cmd, cmd))
			}
			return nil
		}

		// 解析转账信息
		toUserID := params[0]
		pointsStr := params[1]
		points, err := strconv.Atoi(pointsStr)
		if err != nil || points <= 0 {
			p.sendMessage(robot, event, common.T("", "points_amount_invalid"))
			return nil
		}

		// 获取操作者ID
		fromUserID := event.UserID
		fromUserIDStr := fmt.Sprintf("%d", fromUserID)

		if fromUserIDStr == toUserID {
			p.sendMessage(robot, event, common.T("", "points_transfer_self"))
			return nil
		}

		// 执行转账（使用数据库事务）
		reason := common.T("", "points_reason_transfer")
		if cmd == "打赏" || cmd == "reward" {
			reason = common.T("", "points_reason_reward")
		}

		err = db.TransferPoints(p.db, fromUserIDStr, toUserID, points, reason, "transfer")
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_op_failed"), err))
			return nil
		}

		// 发送成功消息
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_transfer_success"), reason, toUserID, reason, points))
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
			p.sendMessage(robot, event, common.T("", "points_get_no_userid"))
			return nil
		}

		// 检查是否已经领取过
		userIDStr := fmt.Sprintf("%d", userID)
		lastGetTime, ok := p.lastGetPointsTime[userIDStr]
		now := time.Now()
		if ok && isSameDay(lastGetTime, now) {
			p.sendMessage(robot, event, common.T("", "points_get_already"))
			return nil
		}

		// 领取5积分
		err := db.AddPoints(p.db, userIDStr, 5, common.T("", "points_reason_daily_bonus"), "daily_bonus")
		if err != nil {
			p.sendMessage(robot, event, common.T("", "points_get_failed"))
			return nil
		}
		p.lastGetPointsTime[userIDStr] = now

		// 获取更新后的积分
		userPoints, _ := db.GetPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_get_success"), userPoints))

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
			p.sendMessage(robot, event, common.T("", "points_deposit_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		matchDep, _, depParams := p.cmdParser.MatchCommandWithParams("存积分|存款", `(\\d+)`, event.RawMessage)
		if matchDep && len(depParams) == 1 {
			amount, err := strconv.Atoi(depParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, common.T("", "points_deposit_amount_invalid"))
				return nil
			}

			err = db.DepositPointsToSavings(p.db, userIDStr, amount)
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_deposit_failed"), err))
				return nil
			}

			saving, _ := db.GetSavingsPoints(p.db, userIDStr)
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_deposit_success"), amount, saving))
			return nil
		}

		matchQuery, _ := p.cmdParser.MatchCommand("存积分|存款", event.RawMessage)
		if !matchQuery {
			return nil
		}

		saving, err := db.GetSavingsPoints(p.db, userIDStr)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_deposit_query_failed"), err))
			return nil
		}

		points, err := db.GetPoints(p.db, userIDStr)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_query_failed_with_err"), err))
			return nil
		}

		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_balance_summary"), points, saving))

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
			p.sendMessage(robot, event, common.T("", "points_withdraw_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("取积分|取款", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, common.T("", "points_withdraw_amount_invalid"))
			return nil
		}

		err = db.WithdrawPointsFromSavings(p.db, userIDStr, amount)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_withdraw_failed"), err))
			return nil
		}

		saving, _ := db.GetSavingsPoints(p.db, userIDStr)
		points, _ := db.GetPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_withdraw_success"), amount, points, saving))

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
			p.sendMessage(robot, event, common.T("", "points_freeze_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("冻结积分|冻结", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, common.T("", "points_freeze_amount_invalid"))
			return nil
		}

		err = db.FreezePoints(p.db, userIDStr, amount, common.T("", "points_reason_manual_freeze"))
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_freeze_failed"), err))
			return nil
		}

		frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_freeze_success"), amount, frozen))

		return nil
	})

	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		userID := event.UserID
		if userID == 0 {
			p.sendMessage(robot, event, common.T("", "points_unfreeze_no_userid"))
			return nil
		}

		userIDStr := fmt.Sprintf("%d", userID)

		match, _, params := p.cmdParser.MatchCommandWithParams("解冻积分|解冻", `(\\d+)`, event.RawMessage)
		if !match || len(params) != 1 {
			return nil
		}

		amount, err := strconv.Atoi(params[0])
		if err != nil || amount <= 0 {
			p.sendMessage(robot, event, common.T("", "points_unfreeze_amount_invalid"))
			return nil
		}

		err = db.UnfreezePoints(p.db, userIDStr, amount, common.T("", "points_reason_manual_unfreeze"))
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_unfreeze_failed"), err))
			return nil
		}

		frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_unfreeze_success"), amount, frozen))

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
