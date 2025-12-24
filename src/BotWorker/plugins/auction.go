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

// AutoBidSetting 自动跟价设置
type AutoBidSetting struct {
	UserID       string // 用户ID
	AuctionID    string // 竞拍ID
	MaxPrice     int    // 最高出价
	BidIncrement int    // 加价幅度
	Status       string // active 激活, disabled 禁用
}

// AuctionItem 竞拍物品
type AuctionItem struct {
	ID            string
	Name          string
	Description   string
	Type          string // physical 实物, virtual 虚拟, group_name 群冠名
	StartTime     time.Time
	EndTime       time.Time
	BasePrice     int
	CurrentPrice  int
	CurrentWinner string
	Status        string // pending 待开始, active 进行中, ended 已结束
	CreatorID     string
	GroupID       string
	SponsorDate   time.Time // 群冠名生效日期（仅群冠名竞拍使用）
}

// AuctionPlugin 竞拍系统插件
type AuctionPlugin struct {
	db *sql.DB
	// 存储竞拍物品，key为竞拍ID
	actions map[string]*AuctionItem
	// 命令解析器
	cmdParser *CommandParser
	// 积分系统插件引用
	pointsPlugin *PointsPlugin
	// 自动跟价设置，key为"userID:auctionID"
	autoBids map[string]*AutoBidSetting
}

// NewAuctionPlugin 创建竞拍系统插件实例
func NewAuctionPlugin(database *sql.DB, pointsPlugin *PointsPlugin) *AuctionPlugin {
	return &AuctionPlugin{
		db:           database,
		actions:      make(map[string]*AuctionItem),
		cmdParser:    NewCommandParser(),
		pointsPlugin: pointsPlugin,
		autoBids:     make(map[string]*AutoBidSetting),
	}
}

func (p *AuctionPlugin) Name() string {
	return "auction"
}

func (p *AuctionPlugin) Description() string {
	return common.T("", "auction_plugin_desc|竞拍系统插件，支持发布物品竞拍、加价、自动跟价等功能")
}

func (p *AuctionPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *AuctionPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "create_auction",
			Description: common.T("", "auction_skill_create_auction_desc|创建一个新的竞拍物品"),
			Usage:       "create_auction name='测试物品' base_price=100 duration=60 description='这是一个测试物品' type='virtual' group_id='123456' user_id='654321'",
			Params: map[string]string{
				"name":        common.T("", "auction_skill_param_name|竞拍物品名称"),
				"base_price":  common.T("", "auction_skill_param_base_price|起拍价格"),
				"duration":    common.T("", "auction_skill_param_duration|持续时间（分钟）"),
				"description": common.T("", "auction_skill_param_description|物品详细描述"),
				"type":        common.T("", "auction_skill_param_type|物品类型（physical, virtual, group_name）"),
				"group_id":    common.T("", "auction_skill_param_group_id|所属群号"),
				"user_id":     common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
		{
			Name:        "place_bid",
			Description: common.T("", "auction_skill_place_bid_desc|对指定竞拍物品进行出价"),
			Usage:       "place_bid auction_id='auction_123' price=200 user_id='654321'",
			Params: map[string]string{
				"auction_id": common.T("", "auction_skill_param_auction_id|竞拍物品ID"),
				"price":      common.T("", "auction_skill_param_price|出价金额"),
				"user_id":    common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
		{
			Name:        "get_auction",
			Description: common.T("", "auction_skill_get_auction_desc|查询指定竞拍物品的详细信息"),
			Usage:       "get_auction auction_id='auction_123'",
			Params: map[string]string{
				"auction_id": common.T("", "auction_skill_param_auction_id|竞拍物品ID"),
			},
		},
		{
			Name:        "list_auctions",
			Description: common.T("", "auction_skill_list_auctions_desc|列出当前群内所有进行中的竞拍"),
			Usage:       "list_auctions group_id='123456'",
			Params: map[string]string{
				"group_id": common.T("", "auction_skill_param_group_id|所属群号"),
			},
		},
		{
			Name:        "set_auto_bid",
			Description: common.T("", "auction_skill_set_auto_bid_desc|设置自动跟价"),
			Usage:       "set_auto_bid auction_id='auction_123' max_price=1000 increment=10 user_id='654321'",
			Params: map[string]string{
				"auction_id": common.T("", "auction_skill_param_auction_id|竞拍物品ID"),
				"max_price":  common.T("", "auction_skill_param_max_price|最高接受价格"),
				"increment":  common.T("", "auction_skill_param_increment|加价幅度"),
				"user_id":    common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
		{
			Name:        "cancel_auto_bid",
			Description: common.T("", "auction_skill_cancel_auto_bid_desc|取消自动跟价"),
			Usage:       "cancel_auto_bid auction_id='auction_123' user_id='654321'",
			Params: map[string]string{
				"auction_id": common.T("", "auction_skill_param_auction_id|竞拍物品ID"),
				"user_id":    common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
		{
			Name:        "get_my_auto_bids",
			Description: common.T("", "auction_skill_get_my_auto_bids_desc|查看我设置的所有自动跟价"),
			Usage:       "get_my_auto_bids user_id='654321'",
			Params: map[string]string{
				"user_id": common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
		{
			Name:        "end_auction",
			Description: common.T("", "auction_skill_end_auction_desc|手动结束竞拍（仅限创建者或管理员）"),
			Usage:       "end_auction auction_id='auction_123' user_id='654321'",
			Params: map[string]string{
				"auction_id": common.T("", "auction_skill_param_auction_id|竞拍物品ID"),
				"user_id":    common.T("", "auction_skill_param_user_id|创建者用户ID"),
			},
		},
	}
}

func (p *AuctionPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	var userID string
	if event != nil {
		userID = fmt.Sprintf("%d", event.UserID)
	} else if params["user_id"] != "" {
		userID = params["user_id"]
	}

	var groupID string
	if event != nil && event.MessageType == "group" {
		groupID = fmt.Sprintf("%d", event.GroupID)
	} else if params["group_id"] != "" {
		groupID = params["group_id"]
	}

	switch skillName {
	case "create_auction":
		name := params["name"]
		basePrice, _ := strconv.Atoi(params["base_price"])
		duration, _ := strconv.Atoi(params["duration"])
		description := params["description"]
		itemType := params["type"]
		if groupID == "" || userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_group_user|❌ 缺少群号或用户ID"))
		}
		msg, err := p.doCreateAuction(name, basePrice, duration, description, itemType, groupID, userID)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_create_failed|发布竞拍失败：%v"), err))
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "place_bid":
		auctionID := params["auction_id"]
		price, _ := strconv.Atoi(params["price"])
		if userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_user|❌ 缺少用户ID"))
		}
		msg, err := p.doPlaceBid(auctionID, price, userID)
		if err != nil {
			p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_bid_failed|出价失败：%v"), err))
			return "", err
		}
		p.sendMessage(robot, event, msg)
		if msg != "" && !errContains(err, common.T("", "auction_not_exists|竞拍不存在")) && !errContains(err, common.T("", "auction_already_ended|竞拍已结束")) {
			p.placeBidAfterHook(robot, event, auctionID)
		}
		return msg, nil

	case "get_auction":
		auctionID := params["auction_id"]
		msg, err := p.doShowAuctionStatus(auctionID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "list_auctions":
		if groupID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_group|❌ 缺少群号"))
		}
		msg, err := p.doShowAllAuctions(groupID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "set_auto_bid":
		auctionID := params["auction_id"]
		maxPrice, _ := strconv.Atoi(params["max_price"])
		increment, _ := strconv.Atoi(params["increment"])
		if userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_user|❌ 缺少用户ID"))
		}
		msg, err := p.doSetAutoBid(auctionID, maxPrice, increment, userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "cancel_auto_bid":
		auctionID := params["auction_id"]
		if userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_user|❌ 缺少用户ID"))
		}
		msg, err := p.doCancelAutoBid(auctionID, userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "get_my_auto_bids":
		if userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_user|❌ 缺少用户ID"))
		}
		msg, err := p.doShowMyAutoBids(userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	case "end_auction":
		auctionID := params["auction_id"]
		if userID == "" {
			return "", fmt.Errorf(common.T("", "auction_missing_user|❌ 缺少用户ID"))
		}
		msg, err := p.doEndAuction(auctionID, userID)
		if err != nil {
			return "", err
		}
		p.sendMessage(robot, event, msg)
		return msg, nil

	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

func (p *AuctionPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println(common.T("", "auction_db_not_configured|❌ 竞拍系统数据库未配置"))
		return
	}
	log.Println(common.T("", "auction_plugin_loaded|✅ 竞拍系统插件已加载"))

	// 启动定时检查竞拍状态的协程
	go p.checkAuctionStatus(robot)

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 统一处理竞拍相关命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "auction") {
				HandleFeatureDisabled(robot, event, "auction")
				return nil
			}
		}

		userIDStr := fmt.Sprintf("%d", event.UserID)
		groupIDStr := ""
		if event.MessageType == "group" {
			groupIDStr = fmt.Sprintf("%d", event.GroupID)
		}

		// 检查是否为创建竞拍命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_create|发布竞拍|创建竞拍|auction_create"), `(\S+)\s+(\d+)\s+(\d+)\s+(.*)`, event.RawMessage); match && len(params) == 4 {
			itemName := params[0]
			basePrice, _ := strconv.Atoi(params[1])
			duration, _ := strconv.Atoi(params[2])
			description := params[3]
			msg, err := p.doCreateAuction(itemName, basePrice, duration, description, "virtual", groupIDStr, userIDStr)
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_create_failed|发布竞拍失败：%v"), err))
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为竞拍群冠名命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_group_sponsor|竞拍群冠名|冠名竞拍|sponsor_auction"), `(\d+)\s+(\d+)\s+(.*)`, event.RawMessage); match && len(params) == 3 {
			basePrice, _ := strconv.Atoi(params[0])
			duration, _ := strconv.Atoi(params[1])
			description := params[2]
			msg, err := p.doCreateAuction("群冠名", basePrice, duration, description, "group_name", groupIDStr, userIDStr)
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_create_failed|发布竞拍失败：%v"), err))
				return nil
			}
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为出价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_bid|出价|竞价|bid"), `(\S+)\s+(\d+)`, event.RawMessage); match && len(params) == 2 {
			auctionID := params[0]
			price, _ := strconv.Atoi(params[1])
			msg, err := p.doPlaceBid(auctionID, price, userIDStr)
			if err != nil {
				p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_bid_failed|出价失败：%v"), err))
				return nil
			}
			p.sendMessage(robot, event, msg)
			if msg != "" && !errContains(err, common.T("", "auction_not_exists|竞拍不存在")) && !errContains(err, common.T("", "auction_already_ended|竞拍已结束")) {
				p.placeBidAfterHook(robot, event, auctionID)
			}
			return nil
		}

		// 检查是否为查看竞拍命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_view|查看竞拍|竞拍详情|view_auction"), `(\S+)`, event.RawMessage); match && len(params) == 1 {
			auctionID := params[0]
			msg, _ := p.doShowAuctionStatus(auctionID)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为查看所有竞拍命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "auction_cmd_view_all|所有竞拍|竞拍列表|list_auctions"), event.RawMessage); match {
			msg, _ := p.doShowAllAuctions(groupIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为设置自动跟价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_set_auto|设置自动跟价|自动出价|auto_bid"), `(\S+)\s+(\d+)\s+(\d+)`, event.RawMessage); match && len(params) == 3 {
			actionID := params[0]
			maxPrice, _ := strconv.Atoi(params[1])
			increment, _ := strconv.Atoi(params[2])
			msg, _ := p.doSetAutoBid(actionID, maxPrice, increment, userIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为取消自动跟价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_cancel_auto|取消自动跟价|停止自动出价|stop_auto_bid"), `(\S+)`, event.RawMessage); match && len(params) == 1 {
			actionID := params[0]
			msg, _ := p.doCancelAutoBid(actionID, userIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为查看我的自动跟价命令
		if match, _ := p.cmdParser.MatchCommand(common.T("", "auction_cmd_view_my_auto|我的自动跟价|我的出价|my_bids"), event.RawMessage); match {
			msg, _ := p.doShowMyAutoBids(userIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		// 检查是否为结束竞拍命令
		if match, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "auction_cmd_end|结束竞拍|停止竞拍|end_auction"), `(\S+)`, event.RawMessage); match && len(params) == 1 {
			actionID := params[0]
			msg, _ := p.doEndAuction(actionID, userIDStr)
			p.sendMessage(robot, event, msg)
			return nil
		}

		return nil
	})
}

func (p *AuctionPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, msg string) {
	if robot == nil || event == nil || msg == "" {
		return
	}
	_, _ = SendTextReply(robot, event, msg)
}

func (p *AuctionPlugin) createAuction(robot plugin.Robot, event *onebot.Event, name string, basePrice int, durationMinutes int, description string, itemType string, groupID string, creatorID string) {
	msg, err := p.doCreateAuction(name, basePrice, durationMinutes, description, itemType, groupID, creatorID)
	if err != nil {
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_create_failed|创建竞拍失败: %v"), err))
		return
	}
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doCreateAuction(name string, basePrice int, durationMinutes int, description string, itemType string, groupID string, creatorID string) (string, error) {
	if basePrice <= 0 {
		return common.T("", "auction_price_must_positive|❌ 起拍价必须大于0"), nil
	}

	if durationMinutes <= 0 {
		return common.T("", "auction_duration_must_positive|❌ 竞拍时长必须大于0分钟"), nil
	}

	// 创建竞拍ID
	auctionID := fmt.Sprintf("auction_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)

	// 群冠名竞拍特殊处理：支持每天竞拍，可以提前拍7天之内的
	var sponsorDate time.Time
	if itemType == "group_name" {
		// duration参数现在表示提前天数（1-7天）
		advanceDays := durationMinutes
		if advanceDays < 1 {
			advanceDays = 1
		} else if advanceDays > 7 {
			advanceDays = 7
		}

		// 计算冠名生效日期
		sponsorDate = startTime.AddDate(0, 0, advanceDays)
		sponsorDate = time.Date(sponsorDate.Year(), sponsorDate.Month(), sponsorDate.Day(), 0, 0, 0, 0, sponsorDate.Location())

		// 竞拍结束时间为冠名生效当天的21点
		endTime = time.Date(sponsorDate.Year(), sponsorDate.Month(), sponsorDate.Day(), 21, 0, 0, 0, sponsorDate.Location())

		// 如果当前时间已经过了当天的21点，则竞拍结束时间为明天的21点
		if time.Now().After(endTime) {
			sponsorDate = sponsorDate.AddDate(0, 0, 1)
			endTime = time.Date(sponsorDate.Year(), sponsorDate.Month(), sponsorDate.Day(), 21, 0, 0, 0, sponsorDate.Location())
		}
	}

	// 创建竞拍物品
	action := &AuctionItem{
		ID:            auctionID,
		Name:          name,
		Description:   description,
		Type:          itemType,
		StartTime:     startTime,
		EndTime:       endTime,
		BasePrice:     basePrice,
		CurrentPrice:  basePrice,
		CurrentWinner: "",
		Status:        "active", // 直接开始
		CreatorID:     creatorID,
		GroupID:       groupID,
		SponsorDate:   sponsorDate, // 设置群冠名生效日期
	}

	// 保存到内存
	p.actions[auctionID] = action

	// 保存到数据库
	if p.db != nil {
		data := map[string]interface{}{
			"name":           action.Name,
			"description":    action.Description,
			"type":           action.Type,
			"start_time":     action.StartTime.Unix(),
			"end_time":       action.EndTime.Unix(),
			"base_price":     action.BasePrice,
			"current_price":  action.CurrentPrice,
			"current_winner": action.CurrentWinner,
			"status":         action.Status,
			"creator_id":     action.CreatorID,
			"group_id":       action.GroupID,
		}
		session := &db.Session{
			SessionID: auctionID,
			UserID:    creatorID,
			GroupID:   groupID,
			State:     "auction:active",
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(p.db, session)
	}

	// 发送开始竞拍消息
	var itemTypeStr string
	switch itemType {
	case "physical":
		itemTypeStr = common.T("", "auction_item_type_physical|实物")
	case "virtual":
		itemTypeStr = common.T("", "auction_item_type_virtual|虚拟物品")
	case "group_name":
		itemTypeStr = common.T("", "auction_item_type_group_name|群冠名")
	default:
		itemTypeStr = common.T("", "auction_item_type_default|物品")
	}

	message := fmt.Sprintf(common.T("", "auction_start_announcement|📢 竞拍开始！\n竞拍ID：%s\n物品类型：%s\n物品名称：%s\n物品描述：%s\n起拍价：%d积分\n当前价格：%d积分\n开始时间：%s\n结束时间：%s\n\n使用 '出价 %s <积分>' 参与竞拍\n使用 '查看竞拍 %s' 查看竞拍详情"),
		action.ID,
		itemTypeStr,
		action.Name,
		action.Description,
		action.BasePrice,
		action.CurrentPrice,
		action.StartTime.Format("2006-01-02 15:04:05"),
		action.EndTime.Format("2006-01-02 15:04:05"),
		action.ID,
		action.ID,
	)

	return message, nil
}

// 出价
func (p *AuctionPlugin) placeBid(robot plugin.Robot, event *onebot.Event, auctionID string, price int, userID string) {
	msg, err := p.doPlaceBid(auctionID, price, userID)
	if err != nil {
		p.sendMessage(robot, event, fmt.Sprintf(common.T("", "auction_bid_failed|❌ 出价失败: %v"), err))
		return
	}
	p.sendMessage(robot, event, msg)

	if msg != "" && !errContains(err, common.T("", "auction_not_exists|竞拍不存在")) && !errContains(err, common.T("", "auction_already_ended|竞拍已结束")) {
		p.placeBidAfterHook(robot, event, auctionID)
	}
}

func errContains(err error, sub string) bool {
	if err == nil {
		return false
	}
	return fmt.Sprintf("%v", err) == sub
}

func (p *AuctionPlugin) doPlaceBid(auctionID string, price int, userID string) (string, error) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		return common.T("", "auction_not_exists|竞拍不存在"), nil
	}

	// 检查竞拍状态
	if action.Status != "active" {
		return common.T("", "auction_already_ended|竞拍已结束"), nil
	}

	// 检查是否超过结束时间
	if time.Now().After(action.EndTime) {
		// p.doEndAuction(auctionID, "system") // 这里无法直接结束，因为需要robot
		return common.T("", "auction_timeout|竞拍已超时结束"), nil
	}

	// 检查出价是否高于当前价格
	if price <= action.CurrentPrice {
		return fmt.Sprintf(common.T("", "auction_bid_higher|❌ 出价必须高于当前价格 %d 积分"), action.CurrentPrice), nil
	}

	// 检查用户积分是否足够
	userPoints := p.pointsPlugin.GetPoints(userID)
	if userPoints < price {
		return fmt.Sprintf(common.T("", "auction_points_insufficient|❌ 积分不足，当前积分：%d，需要：%d"), userPoints, price), nil
	}

	// 冻结上一位竞拍者的积分
	if action.CurrentWinner != "" {
		_ = db.UnfreezePoints(p.db, action.CurrentWinner, action.CurrentPrice, fmt.Sprintf(common.T("", "auction_reason_unfreeze_outbid|竞拍 %s 出价被超过，解冻积分"), action.Name))
	}

	// 冻结当前出价者的积分
	err := db.FreezePoints(p.db, userID, price, fmt.Sprintf(common.T("", "auction_reason_freeze_bid|参与竞拍 %s 的出价"), action.Name))
	if err != nil {
		return fmt.Sprintf(common.T("", "auction_bid_failed|❌ 出价失败: %v"), err), err
	}

	// 更新竞拍信息
	previousWinner := action.CurrentWinner
	action.CurrentPrice = price
	action.CurrentWinner = userID

	// 更新数据库
	if p.db != nil {
		data := map[string]interface{}{
			"current_price":  action.CurrentPrice,
			"current_winner": action.CurrentWinner,
		}
		session := &db.Session{
			SessionID: auctionID,
			UserID:    action.CreatorID,
			GroupID:   action.GroupID,
			State:     "auction:active",
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(p.db, session)
	}

	// 发送出价成功消息
	var winnerMsg string
	if previousWinner == "" {
		winnerMsg = common.T("", "auction_first_bid|🎉 恭喜，您是第一位出价者！")
	} else {
		winnerMsg = fmt.Sprintf(common.T("", "auction_outbid_user|🔥 您的出价已超过前一位竞拍者 %s"), previousWinner)
	}

	message := fmt.Sprintf(common.T("", "auction_bid_success_msg|✅ 出价成功！\n竞拍ID：%s\n物品名称：%s\n出价人：%s\n当前价格：%d积分\n%s\n结束时间：%s\n\n使用 '查看竞拍 %s' 查看详情"),
		action.ID,
		action.Name,
		userID,
		action.CurrentPrice,
		winnerMsg,
		action.EndTime.Format("2006-01-02 15:04:05"),
		action.ID,
	)

	return message, nil
}

// 查看竞拍状态
func (p *AuctionPlugin) showAuctionStatus(robot plugin.Robot, event *onebot.Event, auctionID string) {
	msg, _ := p.doShowAuctionStatus(auctionID)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doShowAuctionStatus(auctionID string) (string, error) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		return common.T("", "auction_not_exists|竞拍不存在"), nil
	}

	// 检查竞拍是否已结束
	if action.Status == "active" && time.Now().After(action.EndTime) {
		// p.doEndAuction(auctionID, "system")
	}

	// 构建状态消息
	var statusStr string
	switch action.Status {
	case "pending":
		statusStr = common.T("", "auction_status_pending|待开始")
	case "active":
		statusStr = common.T("", "auction_status_active|进行中")
	case "ended":
		statusStr = common.T("", "auction_status_ended|已结束")
	}

	var winnerStr string
	if action.CurrentWinner != "" {
		winnerStr = action.CurrentWinner
	} else {
		winnerStr = common.T("", "auction_none|无")
	}

	var remainingTimeStr string
	if action.Status == "active" {
		remainingTime := action.EndTime.Sub(time.Now())
		if remainingTime > 0 {
			remainingTimeStr = fmt.Sprintf(common.T("", "auction_remaining_time|%d分%d秒"), int(remainingTime.Minutes()), int(remainingTime.Seconds())%60)
		} else {
			remainingTimeStr = common.T("", "auction_timed_out|已超时")
		}
	}

	message := fmt.Sprintf(common.T("", "auction_detail_msg|📊 竞拍详情\n竞拍ID：%s\n物品名称：%s\n物品描述：%s\n起拍价：%d积分\n当前价格：%d积分\n当前领先者：%s\n当前状态：%s\n开始时间：%s\n结束时间：%s\n剩余时间：%s\n\n使用 '出价 %s <积分>' 参与竞拍"),
		action.ID,
		action.Name,
		action.Description,
		action.BasePrice,
		action.CurrentPrice,
		winnerStr,
		statusStr,
		action.StartTime.Format("2006-01-02 15:04:05"),
		action.EndTime.Format("2006-01-02 15:04:05"),
		remainingTimeStr,
		action.ID,
	)

	return message, nil
}

// 查看所有竞拍
func (p *AuctionPlugin) showAllAuctions(robot plugin.Robot, event *onebot.Event, groupID string) {
	msg, _ := p.doShowAllAuctions(groupID)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doShowAllAuctions(groupID string) (string, error) {
	// 筛选当前群的竞拍
	var activeAuctions []*AuctionItem
	var endedAuctions []*AuctionItem

	for _, action := range p.actions {
		if action.GroupID != groupID {
			continue
		}

		// 检查是否需要结束竞拍
		if action.Status == "active" && time.Now().After(action.EndTime) {
			// p.doEndAuction(action.ID, "system")
		}

		if action.Status == "active" {
			activeAuctions = append(activeAuctions, action)
		} else {
			endedAuctions = append(endedAuctions, action)
		}
	}

	// 构建消息
	message := common.T("", "auction_list_title|📜 竞拍列表\n")

	if len(activeAuctions) > 0 {
		message += common.T("", "auction_list_active|🔥 进行中的竞拍：\n")
		for _, action := range activeAuctions {
			remainingTime := action.EndTime.Sub(time.Now())
			var remainingStr string
			if remainingTime > 0 {
				remainingStr = fmt.Sprintf(common.T("", "auction_list_remaining|剩余%d分钟"), int(remainingTime.Minutes()))
			} else {
				remainingStr = common.T("", "auction_timed_out|已超时")
			}
			message += fmt.Sprintf(common.T("", "auction_list_item_active|- [%s] %s (当前价: %d, %s)\n"), action.ID, action.Name, action.CurrentPrice, remainingStr)
		}
		message += "\n"
	}

	if len(endedAuctions) > 0 {
		message += common.T("", "auction_list_ended|⌛ 最近结束的竞拍：\n")
		for i, action := range endedAuctions {
			if i >= 5 { // 最多显示5个已结束的竞拍
				message += fmt.Sprintf(common.T("", "auction_list_ended_more|... 还有 %d 个已结束的竞拍"), len(endedAuctions)-5)
				break
			}
			winner := common.T("", "auction_none|无")
			if action.CurrentWinner != "" {
				winner = action.CurrentWinner
			}
			message += fmt.Sprintf(common.T("", "auction_list_item_ended|- [%s] %s (成交价: %d, 胜出者: %s)\n"), action.ID, action.Name, action.CurrentPrice, winner)
		}
	}

	if len(activeAuctions) == 0 && len(endedAuctions) == 0 {
		message += common.T("", "auction_no_activity|暂无竞拍活动\n")
	}

	message += common.T("", "auction_list_usage|\n使用 '查看竞拍 <ID>' 查看详情\n使用 '发布竞拍 <名称> <起拍价> [描述]' 发布竞拍")

	return message, nil
}

// 结束竞拍
func (p *AuctionPlugin) endAuction(robot plugin.Robot, event *onebot.Event, auctionID string, operator string) {
	msg, _ := p.doEndAuction(auctionID, operator)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doEndAuction(auctionID string, operator string) (string, error) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		return common.T("", "auction_not_exists|竞拍不存在"), nil
	}

	// 检查是否有权限结束竞拍（创建者或系统）
	if operator != "system" && operator != action.CreatorID {
		return common.T("", "auction_only_creator_can_end|❌ 只有竞拍发起者或系统可以手动结束竞拍"), nil
	}

	// 检查竞拍是否已经结束
	if action.Status == "ended" {
		return common.T("", "auction_already_ended|竞拍已结束"), nil
	}

	// 更新竞拍状态
	action.Status = "ended"

	// 更新数据库
	if p.db != nil {
		data := map[string]interface{}{
			"status": "ended",
		}
		session := &db.Session{
			SessionID: auctionID,
			UserID:    action.CreatorID,
			GroupID:   action.GroupID,
			State:     "auction:ended",
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(p.db, session)
	}

	// 处理竞拍结果
	if action.CurrentWinner != "" {
		// 扣除中标者的积分
		_ = db.UnfreezePoints(p.db, action.CurrentWinner, action.CurrentPrice, fmt.Sprintf(common.T("", "auction_reason_unfreeze_win|竞拍 %s 中标，解冻积分进行扣除"), action.Name))
		_ = db.AddPoints(p.db, action.CreatorID, action.CurrentPrice, fmt.Sprintf(common.T("", "auction_reason_income|竞拍 %s 成交，获得积分收益"), action.Name), "auction_income")

		// 如果是群冠名竞拍，需要设置群名称
		if action.Type == "group_name" {
			// 这里可以添加设置群名称的逻辑
			// 需要调用机器人API来修改群名称
			// 注意：需要管理员权限
			// 使用竞拍时设置的SponsorDate作为冠名开始时间
			sponsorStartTime := action.SponsorDate
			// 群冠名只持续1天
			sponsorEndTime := sponsorStartTime.AddDate(0, 0, 1)

			message := fmt.Sprintf(common.T("", "auction_group_name_end_msg|🎉 竞拍结束！\n\n【群冠名竞拍】\n冠名内容：%s\n中标人：%s\n最终价格：%d积分\n生效时间：%s\n结束时间：%s\n\n管理员将尽快为您修改群名称。"),
				action.Description, action.CurrentWinner, action.CurrentPrice,
				sponsorStartTime.Format("2006-01-02 15:04:05"),
				sponsorEndTime.Format("2006-01-02 15:04:05"))
			return message, nil
		} else {
			message := fmt.Sprintf(common.T("", "auction_end_success_msg|🎉 竞拍结束！\n\n物品名称：%s\n中标人：%s\n最终价格：%d积分\n\n请联系发起者进行交付。"),
				action.Name, action.CurrentWinner, action.CurrentPrice)
			return message, nil
		}
	} else {
		message := fmt.Sprintf(common.T("", "auction_end_no_bid_msg|⌛ 竞拍结束，由于无人参与，该次竞拍已流拍。\n\n物品名称：%s"),
			action.Name)
		return message, nil
	}
}

// 定时检查竞拍状态
func (p *AuctionPlugin) checkAuctionStatus(robot plugin.Robot) {
	for {
		time.Sleep(60 * time.Second) // 每分钟检查一次

		for _, action := range p.actions {
			if action.Status == "active" && time.Now().After(action.EndTime) {
				// 结束超时的竞拍
				p.endAuction(robot, nil, action.ID, "system")
			}
		}
	}
}

// sendMessage 发送消息 (已在上方定义，此处删除以避免重复定义)
/*
func (p *AuctionPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil || message == "" {
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("Failed to send message: %v\n", err)
	}
}
*/

// 设置自动跟价
func (p *AuctionPlugin) setAutoBid(robot plugin.Robot, event *onebot.Event, auctionID string, maxPrice int, increment int, userID string) {
	msg, _ := p.doSetAutoBid(auctionID, maxPrice, increment, userID)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doSetAutoBid(auctionID string, maxPrice int, increment int, userID string) (string, error) {
	// 检查竞拍是否存在
	action, ok := p.actions[auctionID]
	if !ok {
		return common.T("", "auction_not_exists|竞拍不存在"), nil
	}

	// 检查竞拍是否已结束
	if action.Status == "ended" {
		return common.T("", "auction_auto_bid_ended|❌ 竞拍已结束，无法设置自动出价"), nil
	}

	// 检查最高出价是否高于当前价格
	if maxPrice <= action.CurrentPrice {
		return fmt.Sprintf(common.T("", "auction_auto_bid_higher|❌ 最高出价必须高于当前价格 %d 积分"), action.CurrentPrice), nil
	}

	// 检查加价幅度是否大于0
	if increment <= 0 {
		return common.T("", "auction_auto_bid_increment_positive|❌ 加价幅度必须大于0"), nil
	}

	// 检查用户积分是否足够
	userPoints := p.pointsPlugin.GetPoints(userID)
	if userPoints < maxPrice {
		return fmt.Sprintf(common.T("", "auction_auto_bid_points_insufficient|❌ 积分不足，当前积分：%d，需要：%d"), userPoints, maxPrice), nil
	}

	// 创建自动跟价设置
	key := fmt.Sprintf("%s:%s", userID, auctionID)
	autoBid := &AutoBidSetting{
		UserID:       userID,
		AuctionID:    auctionID,
		MaxPrice:     maxPrice,
		BidIncrement: increment,
		Status:       "active",
	}

	// 保存到自动跟价设置
	p.autoBids[key] = autoBid

	// 保存到数据库
	if p.db != nil {
		data := map[string]interface{}{
			"max_price":     maxPrice,
			"bid_increment": increment,
			"status":        "active",
		}
		session := &db.Session{
			SessionID: fmt.Sprintf("auto_bid:%s", key),
			UserID:    userID,
			GroupID:   action.GroupID,
			State:     "auto_bid:active",
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(p.db, session)
	}

	return fmt.Sprintf(common.T("", "auction_auto_bid_success|✅ 自动出价设置成功！\n竞拍ID：%s\n最高出价：%d积分\n加价幅度：%d积分"), auctionID, maxPrice, increment), nil
}

// cancelAutoBid 取消自动跟价
func (p *AuctionPlugin) cancelAutoBid(robot plugin.Robot, event *onebot.Event, auctionID string, userID string) {
	msg, _ := p.doCancelAutoBid(auctionID, userID)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doCancelAutoBid(auctionID string, userID string) (string, error) {
	key := fmt.Sprintf("%s:%s", userID, auctionID)

	// 检查自动跟价是否存在
	_, ok := p.autoBids[key]
	if !ok {
		return common.T("", "auction_auto_bid_not_set|❌ 您未对该竞拍设置自动出价"), nil
	}

	// 更新状态为禁用
	p.autoBids[key].Status = "disabled"

	// 更新数据库
	if p.db != nil {
		data := map[string]interface{}{
			"status": "disabled",
		}
		session := &db.Session{
			SessionID: fmt.Sprintf("auto_bid:%s", key),
			UserID:    userID,
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(p.db, session)
	}

	// 从内存中删除
	delete(p.autoBids, key)

	return fmt.Sprintf(common.T("", "auction_auto_bid_canceled|✅ 已成功取消竞拍 %s 的自动出价"), auctionID), nil
}

// showMyAutoBids 查看我的自动跟价
func (p *AuctionPlugin) showMyAutoBids(robot plugin.Robot, event *onebot.Event, userID string) {
	msg, _ := p.doShowMyAutoBids(userID)
	p.sendMessage(robot, event, msg)
}

func (p *AuctionPlugin) doShowMyAutoBids(userID string) (string, error) {
	var autoBids []*AutoBidSetting

	// 查找用户的所有自动跟价
	for _, autoBid := range p.autoBids {
		if autoBid.UserID == userID {
			autoBids = append(autoBids, autoBid)
		}
	}

	if len(autoBids) == 0 {
		return common.T("", "auction_my_auto_bid_empty|ℹ️ 您当前没有设置任何自动跟价"), nil
	}

	// 构建消息
	message := common.T("", "auction_my_auto_bid_title|📋 我的自动跟价列表\n")
	for _, autoBid := range autoBids {
		// 获取竞拍信息
		action, ok := p.actions[autoBid.AuctionID]
		if !ok {
			continue
		}

		message += fmt.Sprintf(common.T("", "auction_my_auto_bid_item|- [%s] %s (最高: %d, 幅度: %d, 状态: %s, 当前价: %d)\n"),
			autoBid.AuctionID, action.Name, autoBid.MaxPrice, autoBid.BidIncrement, autoBid.Status, action.CurrentPrice)
	}

	return message, nil
}

// executeAutoBids 执行自动跟价
func (p *AuctionPlugin) executeAutoBids(robot plugin.Robot, event *onebot.Event, auctionID string) {
	if robot == nil || event == nil {
		return
	}
	// 获取竞拍信息
	action, ok := p.actions[auctionID]
	if !ok || action.Status != "active" || time.Now().After(action.EndTime) {
		return
	}

	// 查找该竞拍的所有自动跟价设置
	for _, autoBid := range p.autoBids {
		if autoBid.AuctionID == auctionID && autoBid.Status == "active" {
			// 只有当当前赢家不是自己且当前价格低于最高出价时才执行跟价
			if autoBid.UserID != action.CurrentWinner {
				nextPrice := action.CurrentPrice + autoBid.BidIncrement
				// 如果下一个价格不超过最高出价
				if nextPrice <= autoBid.MaxPrice {
					// 检查用户积分是否足够
					userPoints := p.pointsPlugin.GetPoints(autoBid.UserID)
					if userPoints >= nextPrice {
						// 执行自动出价
						p.placeBid(robot, event, auctionID, nextPrice, autoBid.UserID)
						// 只自动跟价一次，避免无限循环
						return
					}
				}
			}
		}
	}
}

// 在placeBid函数之后调用executeAutoBids来执行自动跟价
func (p *AuctionPlugin) placeBidAfterHook(robot plugin.Robot, event *onebot.Event, auctionID string) {
	// 异步执行自动跟价，避免阻塞
	go func() {
		// 稍微延迟一下，让用户看到出价结果
		time.Sleep(1 * time.Second)
		p.executeAutoBids(robot, event, auctionID)
	}()
}
