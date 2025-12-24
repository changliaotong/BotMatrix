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
	lastSignInTime map[int64]time.Time
	// 存储用户上次领积分时间，key为用户ID，value为领积分时间
	lastGetPointsTime map[int64]time.Time
	// 命令解析器
	cmdParser *CommandParser
}

// NewPointsPlugin 创建积分系统插件实例
func NewPointsPlugin(database *sql.DB) *PointsPlugin {
	return &PointsPlugin{
		db:                database,
		lastSignInTime:    make(map[int64]time.Time),
		lastGetPointsTime: make(map[int64]time.Time),
		cmdParser:         NewCommandParser(),
	}
}

func (p *PointsPlugin) Name() string {
	return "points"
}

func (p *PointsPlugin) Description() string {
	return common.T("", "points_plugin_desc|积分系统插件，支持积分获取、签到、转账、排行榜等功能")
}

func (p *PointsPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 返回插件提供的技能列表
func (p *PointsPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "get_points",
			Description: common.T("", "points_skill_get_points_desc|获取用户当前积分"),
			Usage:       "get_points user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "sign_in_points",
			Description: common.T("", "points_skill_sign_in_points_desc|签到获取积分"),
			Usage:       "sign_in_points user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "get_points_rank",
			Description: common.T("", "points_skill_get_points_rank_desc|获取积分排行榜"),
			Usage:       "get_points_rank",
			Params:      map[string]string{},
		},
		{
			Name:        "transfer_points",
			Description: common.T("", "points_skill_transfer_points_desc|转账积分"),
			Usage:       "transfer_points from_user_id=123 to_user_id=456 amount=100",
			Params: map[string]string{
				"from_user_id": common.T("", "points_skill_param_from_user_id|转出用户ID"),
				"to_user_id":   common.T("", "points_skill_param_to_user_id|转入用户ID"),
				"amount":       common.T("", "points_skill_param_amount|积分数量"),
			},
		},
		{
			Name:        "get_daily_bonus",
			Description: common.T("", "points_skill_get_daily_bonus_desc|领取每日福利积分"),
			Usage:       "get_daily_bonus user_id=123456",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
			},
		},
		{
			Name:        "deposit_points",
			Description: common.T("", "points_skill_deposit_points_desc|存入积分到小金库"),
			Usage:       "deposit_points user_id=123456 amount=100",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
				"amount":  common.T("", "points_skill_param_amount|积分数量"),
			},
		},
		{
			Name:        "withdraw_points",
			Description: common.T("", "points_skill_withdraw_points_desc|从小金库取出积分"),
			Usage:       "withdraw_points user_id=123456 amount=100",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
				"amount":  common.T("", "points_skill_param_amount|积分数量"),
			},
		},
		{
			Name:        "freeze_points",
			Description: common.T("", "points_skill_freeze_points_desc|冻结用户积分"),
			Usage:       "freeze_points user_id=123456 amount=100",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
				"amount":  common.T("", "points_skill_param_amount|积分数量"),
			},
		},
		{
			Name:        "unfreeze_points",
			Description: common.T("", "points_skill_unfreeze_points_desc|解冻用户积分"),
			Usage:       "unfreeze_points user_id=123456 amount=100",
			Params: map[string]string{
				"user_id": common.T("", "points_skill_param_user_id|用户ID"),
				"amount":  common.T("", "points_skill_param_amount|积分数量"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *PointsPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	var userID int64
	if event != nil {
		userID = event.UserID.Int64()
	} else if params["user_id"] != "" {
		val, err := strconv.ParseInt(params["user_id"], 10, 64)
		if err == nil {
			userID = val
		}
	}

	switch skillName {
	case "get_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		return p.doGetPoints(userID)
	case "sign_in_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		return p.doSignInPoints(userID, "")
	case "get_points_rank":
		return p.doGetPointsRank()
	case "transfer_points":
		fromUserIDStr := params["from_user_id"]
		var fromUserID int64
		if fromUserIDStr == "" {
			fromUserID = userID
		} else {
			fromUserID, _ = strconv.ParseInt(fromUserIDStr, 10, 64)
		}

		toUserIDStr := params["to_user_id"]
		toUserID, _ := strconv.ParseInt(toUserIDStr, 10, 64)

		amountStr := params["amount"]
		if fromUserID == 0 || toUserID == 0 || amountStr == "" {
			return "", fmt.Errorf(common.T("", "points_missing_params|缺少必要参数"))
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount <= 0 {
			return "", fmt.Errorf(common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
		}
		return p.doTransferPoints(fromUserID, toUserID, amount, common.T("", "points_reason_transfer|转账"))
	case "get_daily_bonus":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		return p.doGetDailyBonus(userID)
	case "deposit_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		amountStr := params["amount"]
		if amountStr == "" {
			return "", fmt.Errorf(common.T("", "points_missing_amount|缺少积分数量参数"))
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount <= 0 {
			return "", fmt.Errorf(common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
		}
		return p.doDepositPoints(userID, amount)
	case "withdraw_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		amountStr := params["amount"]
		if amountStr == "" {
			return "", fmt.Errorf(common.T("", "points_missing_amount|缺少积分数量参数"))
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount <= 0 {
			return "", fmt.Errorf(common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
		}
		return p.doWithdrawPoints(userID, amount)
	case "freeze_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		amountStr := params["amount"]
		if amountStr == "" {
			return "", fmt.Errorf(common.T("", "points_missing_amount|缺少积分数量参数"))
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount <= 0 {
			return "", fmt.Errorf(common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
		}
		return p.doFreezePoints(userID, amount)
	case "unfreeze_points":
		if userID == 0 {
			return "", fmt.Errorf(common.T("", "points_missing_user_id|缺少用户ID参数"))
		}
		amountStr := params["amount"]
		if amountStr == "" {
			return "", fmt.Errorf(common.T("", "points_missing_amount|缺少积分数量参数"))
		}
		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount <= 0 {
			return "", fmt.Errorf(common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
		}
		return p.doUnfreezePoints(userID, amount)
	default:
		return "", fmt.Errorf(common.T("", "points_skill_not_found|未知技能: %s"), skillName)
	}
}

func (p *PointsPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "points_db_not_configured|积分插件初始化失败：数据库未配置"))
		return
	}
	log.Println(common.T("", "points_plugin_loaded|积分系统插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 统一处理积分相关命令
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

		userID := event.UserID
		if userID == 0 {
			return nil
		}
		userIDStr := fmt.Sprintf("%d", userID)

		// 1. 积分查询
		if match, _ := p.cmdParser.MatchCommand(common.T("", "points_cmd_get|积分|点数|balance"), event.RawMessage); match {
			msg, err := p.doGetPoints(userIDStr)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 2. 签到积分 (含早安/晚安)
		matchSign, signMsg := p.cmdParser.MatchCommand(common.T("", "points_cmd_sign|签到|早安|晚安|signin"), event.RawMessage)
		if matchSign {
			var trigger string
			if signMsg == "早安" || signMsg == "晚安" {
				trigger = signMsg
			}
			rewardMsg, err := p.doSignInPoints(userIDStr, trigger)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, rewardMsg)
			return nil
		}

		// 3. 排行榜
		if match, _ := p.cmdParser.MatchCommand(common.T("", "points_cmd_rank|积分榜|排行榜|rank"), event.RawMessage); match {
			msg, err := p.doGetPointsRank()
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 4. 打赏/转账
		matchTransfer, cmd, params := p.cmdParser.MatchCommandWithParams(common.T("", "points_cmd_transfer|转账|打赏|transfer|reward"), "(\\d+)\\s+(\\d+)", event.RawMessage)
		if matchTransfer {
			if len(params) != 2 {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_transfer_usage|%s 命令用法：%s <用户ID> <积分数量>"), cmd, cmd))
				return nil
			}
			toUserID := params[0]
			points, err := strconv.Atoi(params[1])
			if err != nil || points <= 0 {
				p.sendMessage(robot, event, common.T("", "points_amount_invalid|❌ 积分数量无效，请输入大于0的整数"))
				return nil
			}
			reason := common.T("", "points_reason_transfer|转账")
			if cmd == "打赏" || cmd == "reward" {
				reason = common.T("", "points_reason_reward|打赏")
			}
			msg, err := p.doTransferPoints(userIDStr, toUserID, points, reason)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 5. 领积分
		if match, _ := p.cmdParser.MatchCommand(common.T("", "points_cmd_bonus|领积分|领福利|bonus"), event.RawMessage); match {
			msg, err := p.doGetDailyBonus(userIDStr)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 6. 存积分
		matchDep, _, depParams := p.cmdParser.MatchCommandWithParams(common.T("", "points_cmd_deposit|存积分|存款|deposit"), `(\\d+)`, event.RawMessage)
		if matchDep && len(depParams) == 1 {
			amount, err := strconv.Atoi(depParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, common.T("", "points_deposit_amount_invalid|❌ 存款金额无效，请输入大于0的整数"))
				return nil
			}
			msg, err := p.doDepositPoints(userIDStr, amount)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}
		// 存积分余额查询
		if match, _ := p.cmdParser.MatchCommand(common.T("", "points_cmd_deposit|存积分|存款|deposit"), event.RawMessage); match {
			saving, _ := db.GetSavingsPoints(p.db, userIDStr)
			points, _ := db.GetPoints(p.db, userIDStr)
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_balance_summary|💰 资产概览：\n可用积分：%d\n小金库余额：%d"), points, saving))
			return nil
		}

		// 7. 取积分
		matchWithdraw, _, drawParams := p.cmdParser.MatchCommandWithParams(common.T("", "points_cmd_withdraw|取积分|取款|withdraw"), `(\\d+)`, event.RawMessage)
		if matchWithdraw && len(drawParams) == 1 {
			amount, err := strconv.Atoi(drawParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, common.T("", "points_withdraw_amount_invalid|❌ 取款金额无效，请输入大于0的整数"))
				return nil
			}
			msg, err := p.doWithdrawPoints(userIDStr, amount)
			if err != nil {
				p.sendMessage(robot, event, err.Error())
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 8. 冻结积分
		matchFreeze, _, freezeParams := p.cmdParser.MatchCommandWithParams(common.T("", "points_cmd_freeze|冻结积分|freeze"), `(\\d+)`, event.RawMessage)
		if matchFreeze && len(freezeParams) == 1 {
			amount, err := strconv.Atoi(freezeParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, common.T("", "points_freeze_amount_invalid|❌ 冻结金额无效，请输入大于0的整数"))
				return nil
			}
			err = db.FreezePoints(p.db, userIDStr, amount, common.T("", "points_reason_manual_freeze|手动冻结"))
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_freeze_failed|❌ 冻结失败：%v"), err))
				return nil
			}
			frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_freeze_success|✅ 成功冻结 %d 积分，当前已冻结总额：%d"), amount, frozen))
			return nil
		}

		// 9. 解冻积分
		matchUnfreeze, _, unfreezeParams := p.cmdParser.MatchCommandWithParams(common.T("", "points_cmd_unfreeze|解冻积分|unfreeze"), `(\\d+)`, event.RawMessage)
		if matchUnfreeze && len(unfreezeParams) == 1 {
			amount, err := strconv.Atoi(unfreezeParams[0])
			if err != nil || amount <= 0 {
				p.sendMessage(robot, event, common.T("", "points_unfreeze_amount_invalid|❌ 解冻金额无效，请输入大于0的整数"))
				return nil
			}
			err = db.UnfreezePoints(p.db, userIDStr, amount, common.T("", "points_reason_manual_unfreeze|手动解冻"))
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_unfreeze_failed|❌ 解冻失败：%v"), err))
				return nil
			}
			frozen, _ := db.GetFrozenPoints(p.db, userIDStr)
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "points_unfreeze_success|✅ 成功解冻 %d 积分，当前已冻结总额：%d"), amount, frozen))
			return nil
		}

		// 10. 发言奖励 (排除命令消息)
		if !p.cmdParser.IsCommand(common.T("", "points_cmd_get|积分|点数|balance")+"|"+common.T("", "points_cmd_sign|签到|早安|晚安|signin")+"|"+common.T("", "points_cmd_rank|积分榜|排行榜|rank")+"|"+common.T("", "points_cmd_transfer|转账|打赏|transfer|reward")+"|"+common.T("", "points_cmd_bonus|领积分|领福利|bonus")+"|"+common.T("", "points_cmd_deposit|存积分|存款|deposit")+"|"+common.T("", "points_cmd_withdraw|取积分|取款|withdraw")+"|"+common.T("", "points_cmd_freeze|冻结积分|freeze")+"|"+common.T("", "points_cmd_unfreeze|解冻积分|unfreeze"), event.RawMessage) {
			_ = db.AddPoints(p.db, userIDStr, 1, common.T("", "points_reason_message|发言奖励"), "message_reward")
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *PointsPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "points_send_failed_log|发送积分消息失败: %v"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "points_send_failed_log|发送积分消息失败: %v"), err)
	}
}

func (p *PointsPlugin) AddPoints(userID int64, points int, reason string, category string) {
	if p.db == nil {
		return
	}
	_ = db.AddPoints(p.db, userID, points, reason, category)
}

// doGetPoints 执行获取积分逻辑
func (p *PointsPlugin) doGetPoints(userID int64) (string, error) {
	userPoints, err := db.GetPoints(p.db, userID)
	if err != nil {
		log.Printf(common.T("", "points_query_log_failed|查询积分失败")+": %v", err)
		return "", fmt.Errorf(common.T("", "points_query_failed|查询积分失败，请稍后再试"))
	}
	return fmt.Sprintf(common.T("", "points_current_balance|您当前的积分为：%d"), userPoints), nil
}

// doSignInPoints 执行签到积分逻辑
func (p *PointsPlugin) doSignInPoints(userID int64, trigger string) (string, error) {
	now := time.Now()
	if lastSignIn, ok := p.lastSignInTime[userID]; ok {
		if isSameDay(lastSignIn, now) {
			return "", fmt.Errorf(common.T("", "points_sign_already|您今天已经签到过了 (签到时间: %s)"), lastSignIn.Format("15:04:05"))
		}
	}

	err := db.AddPoints(p.db, userID, 10, common.T("", "points_reason_signin|每日签到"), "sign_in")
	if err != nil {
		log.Printf(common.T("", "points_sign_log_failed|签到积分奖励失败")+": %v", err)
		return "", fmt.Errorf(common.T("", "points_sign_failed|签到失败，请稍后再试"))
	}
	p.lastSignInTime[userID] = now

	userPoints, _ := db.GetPoints(p.db, userID)

	var rewardMsg string
	switch trigger {
	case "早安":
		rewardMsg = fmt.Sprintf(common.T("", "points_sign_morning|☀️ 早安！签到成功，获得 10 积分，当前总积分：%d"), userPoints)
	case "晚安":
		rewardMsg = fmt.Sprintf(common.T("", "points_sign_night|🌙 晚安！签到成功，获得 10 积分，当前总积分：%d"), userPoints)
	default:
		rewardMsg = fmt.Sprintf(common.T("", "points_sign_success|✅ 签到成功，获得 10 积分，当前总积分：%d"), userPoints)
	}
	return rewardMsg, nil
}

// doGetPointsRank 执行获取排行榜逻辑
func (p *PointsPlugin) doGetPointsRank() (string, error) {
	rank, err := p.getPointsRankFromDB()
	if err != nil {
		log.Printf(common.T("", "points_rank_log_failed|获取积分排行榜失败")+": %v", err)
		return "", fmt.Errorf(common.T("", "points_rank_failed|获取排行榜失败，请稍后再试"))
	}

	if len(rank) == 0 {
		return common.T("", "points_rank_empty|暂无积分排行数据"), nil
	}

	msg := common.T("", "points_rank_title|🏆 积分排行榜 (Top 10)") + "\n"
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
		msg += fmt.Sprintf(common.T("", "points_rank_item|%s 用户(%d): %d 积分"), medal, item.UserID, item.Points) + "\n"
	}
	msg += "------------------------\n"
	return msg, nil
}

// doTransferPoints 执行转账积分逻辑
func (p *PointsPlugin) doTransferPoints(fromUserID, toUserID int64, points int, reason string) (string, error) {
	if fromUserID == toUserID {
		return "", fmt.Errorf(common.T("", "points_transfer_self|不能给自己转账哦"))
	}

	err := db.TransferPoints(p.db, fromUserID, toUserID, points, reason, "transfer")
	if err != nil {
		return "", fmt.Errorf(common.T("", "points_op_failed|操作失败: %v"), err)
	}

	return fmt.Sprintf(common.T("", "points_transfer_success|✅ %s成功！\n转账给: %d\n类型: %s\n金额: %d 积分"), reason, toUserID, reason, points), nil
}

// doGetDailyBonus 执行领取每日福利逻辑
func (p *PointsPlugin) doGetDailyBonus(userID int64) (string, error) {
	lastGetTime, ok := p.lastGetPointsTime[userID]
	now := time.Now()
	if ok && isSameDay(lastGetTime, now) {
		return "", fmt.Errorf(common.T("", "points_get_already|您今天已经领取过福利了"))
	}

	err := db.AddPoints(p.db, userID, 20, common.T("", "points_reason_daily|每日福利"), "daily_bonus")
	if err != nil {
		log.Printf(common.T("", "points_daily_log_failed|领取每日福利失败")+": %v", err)
		return "", fmt.Errorf(common.T("", "points_daily_failed|领取失败，请稍后再试"))
	}
	p.lastGetPointsTime[userID] = now

	userPoints, _ := db.GetPoints(p.db, userID)
	return fmt.Sprintf(common.T("", "points_daily_success|✅ 领取成功，获得 20 积分，当前总积分：%d"), userPoints), nil
}

// doDepositPoints 执行存积分逻辑
func (p *PointsPlugin) doDepositPoints(userID string, amount int) (string, error) {
	err := db.DepositPointsToSavings(p.db, userID, amount)
	if err != nil {
		return "", fmt.Errorf(common.T("", "points_deposit_failed|存款失败")+" (%v)", err)
	}

	saving, _ := db.GetSavingsPoints(p.db, userID)
	return fmt.Sprintf(common.T("", "points_deposit_success|✅ 成功存入 %d 积分，小金库当前余额：%d"), amount, saving), nil
}

// doWithdrawPoints 执行取积分逻辑
func (p *PointsPlugin) doWithdrawPoints(userID string, amount int) (string, error) {
	err := db.WithdrawPointsFromSavings(p.db, userID, amount)
	if err != nil {
		return "", fmt.Errorf(common.T("", "points_withdraw_failed|取款失败")+" (%v)", err)
	}

	saving, _ := db.GetSavingsPoints(p.db, userID)
	points, _ := db.GetPoints(p.db, userID)
	return fmt.Sprintf(common.T("", "points_withdraw_success|✅ 成功取出 %d 积分，当前可用：%d，小金库当前余额：%d"), amount, points, saving), nil
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
