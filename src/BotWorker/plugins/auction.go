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
	return "竞拍系统插件，支持竞拍物品和群冠名功能"
}

func (p *AuctionPlugin) Version() string {
	return "1.0.0"
}

func (p *AuctionPlugin) Init(robot plugin.Robot) {
	if p.db == nil {
		log.Println("竞拍系统插件未配置数据库，功能将不可用")
		return
	}
	log.Println("加载竞拍系统插件")

	// 启动定时检查竞拍状态的协程
	go p.checkAuctionStatus(robot)

	// 处理竞拍相关命令
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
		if match, _, params := p.cmdParser.MatchCommandWithParams("创建竞拍", `(\S+)\s+(\d+)\s+(\d+)\s+(.*)`, event.RawMessage); match && len(params) == 4 {
			itemName := params[0]
			basePrice, _ := strconv.Atoi(params[1])
			duration, _ := strconv.Atoi(params[2])
			description := params[3]
			p.createAuction(robot, event, itemName, basePrice, duration, description, "virtual", groupIDStr, userIDStr)
			return nil
		}

		// 检查是否为竞拍群冠名命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("竞拍群冠名", `(\d+)\s+(\d+)\s+(.*)`, event.RawMessage); match && len(params) == 3 {
			basePrice, _ := strconv.Atoi(params[0])
			duration, _ := strconv.Atoi(params[1])
			description := params[2]
			p.createAuction(robot, event, "群冠名", basePrice, duration, description, "group_name", groupIDStr, userIDStr)
			return nil
		}

		// 检查是否为出价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("出价", `(\S+)\s+(\d+)`, event.RawMessage); match && len(params) == 2 {
			auctionID := params[0]
			price, _ := strconv.Atoi(params[1])
			p.placeBid(robot, event, auctionID, price, userIDStr)
			return nil
		}

		// 检查是否为查看竞拍命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("查看竞拍", `(\S+)`, event.RawMessage); match && len(params) == 1 {
			auctionID := params[0]
			p.showAuctionStatus(robot, event, auctionID)
			return nil
		}

		// 检查是否为查看所有竞拍命令
		if match, _ := p.cmdParser.MatchCommand("查看所有竞拍", event.RawMessage); match {
			p.showAllAuctions(robot, event, groupIDStr)
			return nil
		}

		// 检查是否为设置自动跟价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("设置自动跟价", `(\S+)\s+(\d+)\s+(\d+)`, event.RawMessage); match && len(params) == 3 {
			actionID := params[0]
			maxPrice, _ := strconv.Atoi(params[1])
			increment, _ := strconv.Atoi(params[2])
			p.setAutoBid(robot, event, actionID, maxPrice, increment, userIDStr)
			return nil
		}

		// 检查是否为取消自动跟价命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("取消自动跟价", `(\S+)`, event.RawMessage); match && len(params) == 1 {
			actionID := params[0]
			p.cancelAutoBid(robot, event, actionID, userIDStr)
			return nil
		}

		// 检查是否为查看我的自动跟价命令
		if match, _ := p.cmdParser.MatchCommand("查看我的自动跟价", event.RawMessage); match {
			p.showMyAutoBids(robot, event, userIDStr)
			return nil
		}

		// 检查是否为结束竞拍命令
		if match, _, params := p.cmdParser.MatchCommandWithParams("结束竞拍", `(\S+)`, event.RawMessage); match && len(params) == 1 {
			actionID := params[0]
			p.endAuction(robot, event, actionID, userIDStr)
			return nil
		}

		return nil
	})
}

// 创建竞拍
func (p *AuctionPlugin) createAuction(robot plugin.Robot, event *onebot.Event, name string, basePrice int, durationMinutes int, description string, itemType string, groupID string, creatorID string) {
	if basePrice <= 0 {
		p.sendMessage(robot, event, "起拍价必须大于0")
		return
	}

	if durationMinutes <= 0 {
		p.sendMessage(robot, event, "竞拍时长必须大于0分钟")
		return
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
		itemTypeStr = "实物"
	case "virtual":
		itemTypeStr = "虚拟物品"
	case "group_name":
		itemTypeStr = "群冠名"
	default:
		itemTypeStr = "物品"
	}

	message := fmt.Sprintf(
		"📢 竞拍开始！\n"+
			"竞拍ID：%s\n"+
			"物品类型：%s\n"+
			"物品名称：%s\n"+
			"物品描述：%s\n"+
			"起拍价：%d积分\n"+
			"当前价格：%d积分\n"+
			"开始时间：%s\n"+
			"结束时间：%s\n"+
			"\n"+
			"使用 '出价 %s <积分>' 参与竞拍\n"+
			"使用 '查看竞拍 %s' 查看竞拍详情",
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

	p.sendMessage(robot, event, message)

	// 触发自动跟价
	p.placeBidAfterHook(robot, event, auctionID)
}

// 出价
func (p *AuctionPlugin) placeBid(robot plugin.Robot, event *onebot.Event, auctionID string, price int, userID string) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		p.sendMessage(robot, event, "竞拍不存在")
		return
	}

	// 检查竞拍状态
	if action.Status != "active" {
		p.sendMessage(robot, event, "竞拍已结束")
		return
	}

	// 检查是否超过结束时间
	if time.Now().After(action.EndTime) {
		p.endAuction(robot, event, auctionID, "system")
		p.sendMessage(robot, event, "竞拍已超时结束")
		return
	}

	// 检查出价是否高于当前价格
	if price <= action.CurrentPrice {
		p.sendMessage(robot, event, fmt.Sprintf("出价必须高于当前价格 %d 积分", action.CurrentPrice))
		return
	}

	// 检查用户积分是否足够
	userPoints := p.pointsPlugin.GetPoints(userID)
	if userPoints < price {
		p.sendMessage(robot, event, fmt.Sprintf("积分不足，当前积分：%d，需要：%d", userPoints, price))
		return
	}

	// 冻结上一位竞拍者的积分
	if action.CurrentWinner != "" {
		_ = db.UnfreezePoints(p.db, action.CurrentWinner, action.CurrentPrice, fmt.Sprintf("竞拍 %s 出价被超过，解冻积分", action.Name))
	}

	// 冻结当前出价者的积分
	err := db.FreezePoints(p.db, userID, price, fmt.Sprintf("参与竞拍 %s 的出价", action.Name))
	if err != nil {
		p.sendMessage(robot, event, fmt.Sprintf("出价失败：%v", err))
		return
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
		_ = db.UpdateSession(p.db, session)
	}

	// 发送出价成功消息
	var winnerMsg string
	if previousWinner == "" {
		winnerMsg = "首次出价"
	} else {
		winnerMsg = fmt.Sprintf("超过用户 %s", previousWinner)
	}

	message := fmt.Sprintf(
		"💰 出价成功！\n"+
			"竞拍ID：%s\n"+
			"物品名称：%s\n"+
			"出价者：%s\n"+
			"当前价格：%d积分\n"+
			"状态：%s\n"+
			"结束时间：%s\n"+
			"\n"+
			"使用 '出价 %s <积分>' 继续竞拍",
		action.ID,
		action.Name,
		userID,
		action.CurrentPrice,
		winnerMsg,
		action.EndTime.Format("2006-01-02 15:04:05"),
		action.ID,
	)

	p.sendMessage(robot, event, message)
}

// 查看竞拍状态
func (p *AuctionPlugin) showAuctionStatus(robot plugin.Robot, event *onebot.Event, auctionID string) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		p.sendMessage(robot, event, "竞拍不存在")
		return
	}

	// 检查竞拍是否已结束
	if action.Status == "active" && time.Now().After(action.EndTime) {
		p.endAuction(robot, event, auctionID, "system")
	}

	// 构建状态消息
	var statusStr string
	switch action.Status {
	case "pending":
		statusStr = "待开始"
	case "active":
		statusStr = "进行中"
	case "ended":
		statusStr = "已结束"
	}

	var winnerStr string
	if action.CurrentWinner != "" {
		winnerStr = action.CurrentWinner
	} else {
		winnerStr = "暂无"
	}

	var remainingTimeStr string
	if action.Status == "active" {
		remainingTime := action.EndTime.Sub(time.Now())
		if remainingTime > 0 {
			remainingTimeStr = fmt.Sprintf("剩余时间：%d分钟%d秒", int(remainingTime.Minutes()), int(remainingTime.Seconds())%60)
		} else {
			remainingTimeStr = "已超时"
		}
	}

	message := fmt.Sprintf(
		"📋 竞拍详情\n"+
			"竞拍ID：%s\n"+
			"物品名称：%s\n"+
			"物品描述：%s\n"+
			"起拍价：%d积分\n"+
			"当前价格：%d积分\n"+
			"当前竞拍者：%s\n"+
			"状态：%s\n"+
			"开始时间：%s\n"+
			"结束时间：%s\n"+
			"%s\n"+
			"\n"+
			"使用 '出价 %s <积分>' 参与竞拍",
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

	p.sendMessage(robot, event, message)
}

// 设置自动跟价

// 查看所有竞拍
func (p *AuctionPlugin) showAllAuctions(robot plugin.Robot, event *onebot.Event, groupID string) {
	// 筛选当前群的竞拍
	var activeAuctions []*AuctionItem
	var endedAuctions []*AuctionItem

	for _, action := range p.actions {
		if action.GroupID != groupID {
			continue
		}

		// 检查是否需要结束竞拍
		if action.Status == "active" && time.Now().After(action.EndTime) {
			p.endAuction(robot, event, action.ID, "system")
		}

		if action.Status == "active" {
			activeAuctions = append(activeAuctions, action)
		} else {
			endedAuctions = append(endedAuctions, action)
		}
	}

	// 构建消息
	message := "🏆 竞拍列表\n\n"

	if len(activeAuctions) > 0 {
		message += "📢 进行中的竞拍：\n"
		for _, action := range activeAuctions {
			remainingTime := action.EndTime.Sub(time.Now())
			var remainingStr string
			if remainingTime > 0 {
				remainingStr = fmt.Sprintf("剩余%d分钟", int(remainingTime.Minutes()))
			} else {
				remainingStr = "已超时"
			}
			message += fmt.Sprintf("ID: %s | %s | 当前价格: %d积分 | %s\n", action.ID, action.Name, action.CurrentPrice, remainingStr)
		}
		message += "\n"
	}

	if len(endedAuctions) > 0 {
		message += "🔚 已结束的竞拍：\n"
		for i, action := range endedAuctions {
			if i >= 5 { // 最多显示5个已结束的竞拍
				message += fmt.Sprintf("... 还有 %d 个已结束的竞拍\n", len(endedAuctions)-5)
				break
			}
			winner := "流拍"
			if action.CurrentWinner != "" {
				winner = action.CurrentWinner
			}
			message += fmt.Sprintf("ID: %s | %s | 最终价格: %d积分 | 赢家: %s\n", action.ID, action.Name, action.CurrentPrice, winner)
		}
	}

	if len(activeAuctions) == 0 && len(endedAuctions) == 0 {
		message += "暂无竞拍活动"
	}

	message += "\n\n使用 '创建竞拍 <名称> <起拍价> <时长(分钟)> <描述>' 创建新竞拍"

	p.sendMessage(robot, event, message)
}

// 结束竞拍
func (p *AuctionPlugin) endAuction(robot plugin.Robot, event *onebot.Event, auctionID string, operator string) {
	// 查找竞拍
	action, ok := p.actions[auctionID]
	if !ok {
		p.sendMessage(robot, event, "竞拍不存在")
		return
	}

	// 检查是否有权限结束竞拍（创建者或系统）
	if operator != "system" && operator != action.CreatorID {
		p.sendMessage(robot, event, "只有竞拍创建者可以结束竞拍")
		return
	}

	// 检查竞拍是否已经结束
	if action.Status == "ended" {
		p.sendMessage(robot, event, "竞拍已经结束")
		return
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
		_ = db.UpdateSession(p.db, session)
	}

	// 处理竞拍结果
	if action.CurrentWinner != "" {
		// 扣除中标者的积分
		_ = db.UnfreezePoints(p.db, action.CurrentWinner, action.CurrentPrice, fmt.Sprintf("竞拍 %s 中标，扣除积分", action.Name))
		_ = db.AddPoints(p.db, action.CreatorID, action.CurrentPrice, fmt.Sprintf("竞拍 %s 获得收入", action.Name), "auction_income")

		// 如果是群冠名竞拍，需要设置群名称
		if action.Type == "group_name" {
			// 这里可以添加设置群名称的逻辑
			// 需要调用机器人API来修改群名称
			// 注意：需要管理员权限
			// 使用竞拍时设置的SponsorDate作为冠名开始时间
			sponsorStartTime := action.SponsorDate
			// 群冠名只持续1天
			sponsorEndTime := sponsorStartTime.AddDate(0, 0, 1)

			message := fmt.Sprintf("🎉 群冠名竞拍结束！\n"+
				"群冠名：%s\n"+
				"中标者：%s\n"+
				"中标价格：%d积分\n"+
				"冠名开始时间：%s\n"+
				"冠名结束时间：%s\n",
				action.Description, action.CurrentWinner, action.CurrentPrice,
				sponsorStartTime.Format("2006-01-02 15:04:05"),
				sponsorEndTime.Format("2006-01-02 15:04:05"))
			p.sendMessage(robot, event, message)
		} else {
			message := fmt.Sprintf("🎉 竞拍结束！\n"+
				"竞拍物品：%s\n"+
				"中标者：%s\n"+
				"中标价格：%d积分\n"+
				"恭喜中标！",
				action.Name, action.CurrentWinner, action.CurrentPrice)
			p.sendMessage(robot, event, message)
		}
	} else {
		message := fmt.Sprintf("🔚 竞拍结束！\n"+
			"竞拍物品：%s\n"+
			"无人出价，流拍",
			action.Name)
		p.sendMessage(robot, event, message)
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

// sendMessage 发送消息
func (p *AuctionPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// setAutoBid 设置自动跟价
func (p *AuctionPlugin) setAutoBid(robot plugin.Robot, event *onebot.Event, auctionID string, maxPrice int, increment int, userID string) {
	// 检查竞拍是否存在
	action, ok := p.actions[auctionID]
	if !ok {
		p.sendMessage(robot, event, "竞拍不存在")
		return
	}

	// 检查竞拍是否已结束
	if action.Status == "ended" {
		p.sendMessage(robot, event, "竞拍已结束，无法设置自动跟价")
		return
	}

	// 检查最高出价是否高于当前价格
	if maxPrice <= action.CurrentPrice {
		p.sendMessage(robot, event, fmt.Sprintf("最高出价必须高于当前价格 %d 积分", action.CurrentPrice))
		return
	}

	// 检查加价幅度是否大于0
	if increment <= 0 {
		p.sendMessage(robot, event, "加价幅度必须大于0")
		return
	}

	// 检查用户积分是否足够
	userPoints := p.pointsPlugin.GetPoints(userID)
	if userPoints < maxPrice {
		p.sendMessage(robot, event, fmt.Sprintf("积分不足，当前积分：%d，最高出价需要：%d", userPoints, maxPrice))
		return
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

	p.sendMessage(robot, event, fmt.Sprintf("自动跟价设置成功！\n竞拍ID：%s\n最高出价：%d积分\n加价幅度：%d积分", auctionID, maxPrice, increment))
}

// cancelAutoBid 取消自动跟价
func (p *AuctionPlugin) cancelAutoBid(robot plugin.Robot, event *onebot.Event, auctionID string, userID string) {
	key := fmt.Sprintf("%s:%s", userID, auctionID)

	// 检查自动跟价是否存在
	_, ok := p.autoBids[key]
	if !ok {
		p.sendMessage(robot, event, "您未设置该竞拍的自动跟价")
		return
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
		_ = db.UpdateSession(p.db, session)
	}

	// 从内存中删除
	delete(p.autoBids, key)

	p.sendMessage(robot, event, fmt.Sprintf("自动跟价已取消！\n竞拍ID：%s", auctionID))
}

// showMyAutoBids 查看我的自动跟价
func (p *AuctionPlugin) showMyAutoBids(robot plugin.Robot, event *onebot.Event, userID string) {
	var autoBids []*AutoBidSetting

	// 查找用户的所有自动跟价
	for _, autoBid := range p.autoBids {
		if autoBid.UserID == userID {
			autoBids = append(autoBids, autoBid)
		}
	}

	if len(autoBids) == 0 {
		p.sendMessage(robot, event, "您没有设置任何自动跟价")
		return
	}

	// 构建消息
	message := "📋 我的自动跟价设置\n\n"
	for _, autoBid := range autoBids {
		// 获取竞拍信息
		action, ok := p.actions[autoBid.AuctionID]
		if !ok {
			continue
		}

		message += fmt.Sprintf("竞拍ID：%s\n竞拍物品：%s\n最高出价：%d积分\n加价幅度：%d积分\n状态：%s\n当前价格：%d积分\n\n",
			autoBid.AuctionID, action.Name, autoBid.MaxPrice, autoBid.BidIncrement, autoBid.Status, action.CurrentPrice)
	}

	p.sendMessage(robot, event, message)
}

// executeAutoBids 执行自动跟价
func (p *AuctionPlugin) executeAutoBids(robot plugin.Robot, event *onebot.Event, auctionID string) {
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
