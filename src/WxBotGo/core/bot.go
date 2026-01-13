package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/eatmoreapple/openwechat"
	"github.com/gorilla/websocket"
)

// Callback interface for mobile binding
type BotCallback interface {
	OnLog(msg string)
	OnQrCode(url string)
}

// IDMapping 用于存储微信临时ID到持久化ID的映射
type IDMapping struct {
	WeChatID     string `json:"wechat_id"`     // 微信临时ID
	PersistentID string `json:"persistent_id"` // 持久化ID
	Nickname     string `json:"nickname"`      // 用户/群组昵称
	Type         string `json:"type"`          // 类型：user/group
	CreatedAt    int64  `json:"created_at"`    // 创建时间
	UpdatedAt    int64  `json:"updated_at"`    // 更新时间
}

// IDMapStore 用于管理ID映射的持久化存储
type IDMapStore struct {
	Mappings map[string]IDMapping `json:"mappings"` // wechat_id -> mapping
}

// WxBot 微信机器人核心结构体
type WxBot struct {
	ManagerURL    string
	SelfID        string
	ReportSelfMsg bool // 是否上报自身消息

	wsConn  *websocket.Conn
	wsMutex sync.Mutex
	mySelf  *openwechat.Self
	bot     *openwechat.Bot

	// 消息映射表，用于保存消息ID到SentMessage的映射，支持撤回功能
	msgMap   map[string]*openwechat.SentMessage
	msgMutex sync.Mutex

	// ID映射表，用于实现ID固化
	idMap   map[string]IDMapping // wechat_id -> mapping
	idMutex sync.Mutex

	callback BotCallback
}

func NewWxBot(managerUrl, selfId string, cb BotCallback) *WxBot {
	if managerUrl == "" {
		managerUrl = "ws://localhost:3001"
	}
	// 不再硬编码默认selfid，完全由配置文件或调用者决定
	return &WxBot{
		ManagerURL:    managerUrl,
		SelfID:        selfId,
		ReportSelfMsg: true, // 默认上报自身消息
		msgMap:        make(map[string]*openwechat.SentMessage),
		idMap:         make(map[string]IDMapping),
		callback:      cb,
	}
}

func (b *WxBot) Log(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Println(msg)
	if b.callback != nil {
		b.callback.OnLog(msg)
	}
}

// loadIDMap 从文件加载ID映射表
func (b *WxBot) loadIDMap() {
	filePath := "id_mapping.json"
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		b.Log("[ID Mapping] No existing ID mapping file, will create new one")
		return
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		b.Log("[ID Mapping] Failed to read ID mapping file: %v", err)
		return
	}

	// 解析JSON
	var store IDMapStore
	if err := json.Unmarshal(data, &store); err != nil {
		b.Log("[ID Mapping] Failed to parse ID mapping file: %v", err)
		return
	}

	// 更新到内存
	b.idMutex.Lock()
	b.idMap = store.Mappings
	b.idMutex.Unlock()
	b.Log("[ID Mapping] Loaded %d ID mappings from file", len(store.Mappings))
}

// saveIDMap 将ID映射表保存到文件
func (b *WxBot) saveIDMap() {
	b.idMutex.Lock()
	defer b.idMutex.Unlock()

	store := IDMapStore{
		Mappings: b.idMap,
	}

	// 序列化JSON
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		b.Log("[ID Mapping] Failed to marshal ID mappings: %v", err)
		return
	}

	// 写入文件
	if err := os.WriteFile("id_mapping.json", data, 0644); err != nil {
		b.Log("[ID Mapping] Failed to write ID mapping file: %v", err)
		return
	}

	b.Log("[ID Mapping] Saved %d ID mappings to file", len(b.idMap))
}

// getPersistentID 获取持久化ID，如果不存在则创建新的
func (b *WxBot) getPersistentID(wechatID, nickname, idType string) string {
	b.idMutex.Lock()
	defer b.idMutex.Unlock()

	// 检查是否已有映射
	if mapping, exists := b.idMap[wechatID]; exists {
		// 更新昵称和时间
		mapping.Nickname = nickname
		mapping.UpdatedAt = time.Now().Unix()
		b.idMap[wechatID] = mapping
		return mapping.PersistentID
	}

	// 创建新的映射
	now := time.Now().Unix()
	// 使用昵称+时间戳生成相对唯一的持久化ID
	persistentID := fmt.Sprintf("%s_%s_%d", idType, nickname, now)
	// 如果昵称为空，使用wechatID的前8位
	if nickname == "" {
		persistentID = fmt.Sprintf("%s_%s", idType, wechatID[:8])
	}

	mapping := IDMapping{
		WeChatID:     wechatID,
		PersistentID: persistentID,
		Nickname:     nickname,
		Type:         idType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	b.idMap[wechatID] = mapping
	return persistentID
}

// batchUpdateIDMap 批量更新ID映射表
func (b *WxBot) batchUpdateIDMap() {
	b.Log("[ID Mapping] Starting batch update...")

	// 更新好友映射
	friends, err := b.mySelf.Friends()
	if err == nil {
		for _, friend := range friends {
			b.getPersistentID(friend.UserName, friend.NickName, "user")
		}
		b.Log("[ID Mapping] Updated friend mappings")
	}

	// 更新群组映射
	groups, err := b.mySelf.Groups()
	if err == nil {
		for _, group := range groups {
			b.getPersistentID(group.UserName, group.NickName, "group")
			// 更新群成员映射
			members, err := group.Members()
			if err == nil {
				for _, member := range members {
					b.getPersistentID(member.UserName, member.NickName, "user")
				}
			}
		}
		b.Log("[ID Mapping] Updated group and member mappings")
	}

	// 保存到文件
	b.saveIDMap()
	b.Log("[ID Mapping] Batch update completed")
}

func (b *WxBot) Start() {
	b.Log("[WxBotGo] Starting... Target Manager: %s, SelfID: %s", b.ManagerURL, b.SelfID)

	// 使用普通网页模式，不修改协议头，避免账号被封
	b.Log("[WxBotGo] 使用普通网页模式，不修改协议头")
	b.bot = openwechat.DefaultBot(openwechat.Normal) // 使用Normal模式，不修改协议头

	b.bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() {
			b.Log("[WeChat] Recv Text: %s", msg.Content)
		}
		go b.HandleWeChatMsg(msg)
	}

	// Add message handlers for different types
	b.bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() {
			b.Log("[WeChat] Recv Text: %s", msg.Content)
		}
		go b.HandleWeChatMsg(msg)
	}

	b.bot.UUIDCallback = func(uuid string) {
		qrcodeUrl := "https://login.weixin.qq.com/l/" + uuid
		b.Log("QRCODE:%s", qrcodeUrl) // Special prefix for easy parsing if callback fails
		if b.callback != nil {
			b.callback.OnQrCode(qrcodeUrl)
		}
	}

	// Login logic
	// Note: On Android, local storage path might need adjustment.
	// But standard file operations usually work in app sandbox.
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()

	if err := b.bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		b.Log("[WxBotGo] Hot login failed, scanning QR code required.")
		if err := b.bot.Login(); err != nil {
			b.Log("[WxBotGo] Login error: %v", err)
			return
		}
	}

	var err error
	b.mySelf, err = b.bot.GetCurrentUser()
	if err != nil {
		b.Log("[WxBotGo] GetCurrentUser error: %v", err)
		return
	}
	b.Log("[WxBotGo] Login Success! User: %s (%s)", b.mySelf.NickName, b.mySelf.UserName)

	b.Log("[WxBotGo] Loading contacts...")
	if err := b.mySelf.UpdateMembersDetail(); err != nil {
		b.Log("[WxBotGo] Error updating members: %v", err)
	}
	b.mySelf.Friends()
	b.mySelf.Groups()
	b.Log("[WxBotGo] Contacts loaded.")

	// ID映射表管理
	b.loadIDMap()
	b.batchUpdateIDMap()

	go b.connectToNexus()

	b.bot.Block()
}

func (b *WxBot) connectToNexus() {
	// Optional token
	token := os.Getenv("MANAGER_TOKEN")
	u, _ := url.Parse(b.ManagerURL)
	if token != "" {
		q := u.Query()
		q.Set("token", token)
		u.RawQuery = q.Encode()
	}
	b.Log("[WebSocket] Connecting to %s", u.String())

	for {
		var err error
		b.wsMutex.Lock()
		b.wsConn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		b.wsMutex.Unlock()

		if err != nil {
			b.Log("[WebSocket] Connect error: %v, retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Send Identify Packet
		identify := map[string]any{
			"type":            "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         b.SelfID,
			"platform":        "wechat-go",
		}

		// Send Identify and wait for configuration
		b.wsMutex.Lock()
		b.wsConn.WriteJSON(identify)
		b.wsMutex.Unlock()

		// Receive initial configuration from BotNexus
		_, configMsg, err := b.wsConn.ReadMessage()
		if err != nil {
			b.Log("[WebSocket] Failed to receive configuration: %v", err)
			break
		}

		// Parse configuration
		var config map[string]any
		if err := json.Unmarshal(configMsg, &config); err == nil {
			b.Log("[WebSocket] Received configuration from BotNexus")
			// Apply configuration
			if features, ok := config["features"].(map[string]any); ok {
				if reportSelfMsg, ok := features["report_self_msg"].(bool); ok {
					b.ReportSelfMsg = reportSelfMsg
				}
			}
		}
		b.wsMutex.Lock()
		b.wsConn.WriteJSON(identify)
		b.wsMutex.Unlock()

		b.Log("[WebSocket] Connected to BotNexus!")

		// Start Heartbeat Loop
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					b.wsMutex.Lock()
					if b.wsConn == nil {
						b.wsMutex.Unlock()
						return
					}
					heartbeat := map[string]any{
						"type":            "meta_event",
						"meta_event_type": "heartbeat",
						"time":            time.Now().Unix(),
						"self_id":         b.SelfID,
						"status": map[string]any{
							"online": true,
							"good":   true,
						},
					}
					b.wsConn.WriteJSON(heartbeat)
					b.wsMutex.Unlock()
				}
			}
		}()

		// Listen for commands and configuration updates
		for {
			_, message, err := b.wsConn.ReadMessage()
			if err != nil {
				b.Log("[WebSocket] Read error: %v", err)
				break
			}

			// Check if it's a configuration update
			var configUpdate map[string]any
			if err := json.Unmarshal(message, &configUpdate); err == nil {
				if _, ok := configUpdate["features"]; ok {
					// Apply configuration update
					if features, ok := configUpdate["features"].(map[string]any); ok {
						if reportSelfMsg, ok := features["report_self_msg"].(bool); ok {
							b.ReportSelfMsg = reportSelfMsg
							b.Log("[Configuration] Updated report_self_msg to %v", reportSelfMsg)
						}
					}
					continue
				}
			}

			// Handle Action
			var action OneBotAction
			if err := json.Unmarshal(message, &action); err == nil && action.Action != "" {
				b.Log("[WebSocket] Action Recv: %s", action.Action)
				go b.HandleAction(action)
			}
		}

		b.wsMutex.Lock()
		b.wsConn.Close()
		b.wsConn = nil
		b.wsMutex.Unlock()
		b.Log("[WebSocket] Disconnected, reconnecting...")
	}
}

// Send Event to BotNexus
func (b *WxBot) sendEvent(event OneBotEvent) {
	b.wsMutex.Lock()
	defer b.wsMutex.Unlock()
	if b.wsConn == nil {
		return
	}
	// Fill Common Fields
	event.Time = time.Now().Unix()
	event.SelfID = b.SelfID

	bytes, _ := json.Marshal(event)
	b.wsConn.WriteMessage(websocket.TextMessage, bytes)
}

// ConvertToOneBotEvent converts a WeChat message to OneBot event format
func (b *WxBot) ConvertToOneBotEvent(msg *openwechat.Message) OneBotEvent {
	event := OneBotEvent{
		PostType: "message",
	}

	if msg.IsSendBySelf() {
		// 自身消息处理
		event.MessageType = "private"
		event.SubType = "self"
		event.UserID = b.SelfID
		event.Sender = &Sender{
			UserID:   b.SelfID,
			Nickname: "Self",
		}
	} else {
		// 他人消息处理
		sender, err := msg.Sender()
		if err != nil {
			b.Log("Error getting sender: %v", err)
			sender = &openwechat.User{NickName: "Unknown"}
		}

		if msg.IsSendByGroup() {
			event.MessageType = "group"
			event.SubType = "normal"

			groupSender, err := msg.SenderInGroup()
			if err == nil {
				// 获取群成员的持久化ID
				event.UserID = b.getPersistentID(groupSender.UserName, groupSender.NickName, "user")
				event.Sender = &Sender{
					UserID:   event.UserID,
					Nickname: groupSender.NickName,
				}
			} else {
				// 获取发送者的持久化ID
				event.UserID = b.getPersistentID(sender.UserName, sender.NickName, "user")
				event.Sender = &Sender{UserID: event.UserID, Nickname: sender.NickName}
			}

			group := sender
			// 获取群组的持久化ID
			event.GroupID = b.getPersistentID(group.UserName, group.NickName, "group")

		} else {
			event.MessageType = "private"
			event.SubType = "friend"
			// 获取好友的持久化ID
			event.UserID = b.getPersistentID(sender.UserName, sender.NickName, "user")
			event.Sender = &Sender{
				UserID:   event.UserID,
				Nickname: sender.NickName,
			}
		}
	}

	// Handle different message types according to OneBot 11 specification
	if msg.IsText() {
		// For text messages, we can send as plain string
		event.Message = msg.Content
		event.RawMessage = msg.Content
	} else if msg.IsPicture() {
		// For picture messages, we need to convert to CQ code format
		// However, since we don't have access to the actual image URL, we use a placeholder
		event.Message = "[CQ:image,file=placeholder.jpg]"
		event.RawMessage = "[图片]"
	} else if msg.IsVoice() {
		// For voice messages, we need to convert to CQ code format
		event.Message = "[CQ:record,file=placeholder.amr]"
		event.RawMessage = "[语音]"
	} else if msg.IsCard() {
		// For card messages, we need to convert to CQ code format
		event.Message = "[CQ:contact,type=qq,id=123456]"
		event.RawMessage = "[名片]"
	} else if msg.IsEmoticon() {
		// For emoticon messages, we need to convert to CQ code format
		event.Message = "[CQ:face,id=1]"
		event.RawMessage = "[表情]"
	} else {
		// Unsupported message type
		event.Message = "[不支持的消息类型]"
		event.RawMessage = "[不支持的消息类型]"
	}

	// 消息ID使用微信原始ID，不进行固化
	event.MessageID = msg.MsgId
	return event
}

// Handle WeChat Message -> OneBot Event
func (b *WxBot) HandleWeChatMsg(msg *openwechat.Message) {
	// 如果是自身消息且未开启上报，则忽略
	if msg.IsSendBySelf() && !b.ReportSelfMsg {
		return
	}

	// 处理群系统消息
	if msg.IsSystem() && msg.IsSendByGroup() {
		// 群成员加入事件 - 基于openwechat库的IsJoinGroup()方法，这是可靠的
		if msg.IsJoinGroup() {
			// 构造群成员增加通知事件
			noticeEvent := OneBotEvent{
				PostType:    "notice",
				NoticeType:  "group_member_increase",
				MessageType: "group",
				GroupID:     b.getPersistentID(msg.FromUserName, "", "group"), // 获取群组的持久化ID
				UserID:      "",                                               // 系统消息无法直接获取用户ID，需要业务系统解析
				RawMessage:  msg.Content,
			}
			// 发送通知事件
			b.sendEvent(noticeEvent)
			return
		}
	}

	// 处理拍一拍事件
	if msg.IsTickled() {
		// 构造拍一拍通知事件
		noticeEvent := OneBotEvent{
			PostType:    "notice",
			NoticeType:  "group_poke",
			MessageType: "group",
			GroupID:     b.getPersistentID(msg.FromUserName, "", "group"), // 获取群组的持久化ID
			UserID:      "",                                               // 系统消息无法直接获取用户ID，需要解析
			RawMessage:  msg.Content,
		}
		// 发送通知事件
		b.sendEvent(noticeEvent)
		return
	}

	// Convert to OneBot event format
	event := b.ConvertToOneBotEvent(msg)

	// Send to BotNexus via WebSocket
	b.sendEvent(event)
}

// Handle OneBot Action -> WeChat API
func (b *WxBot) HandleAction(action OneBotAction) {
	resp := OneBotResponse{
		Status: "ok",
		Echo:   action.Echo,
		Data:   map[string]any{},
	}

	// Parse action parameters
	var params ActionParams
	bytes, _ := json.Marshal(action.Params)
	if err := json.Unmarshal(bytes, &params); err != nil {
		resp.Status = "failed"
		resp.RetCode = -1
		resp.Message = fmt.Sprintf("failed to parse action params: %v", err)
		b.wsMutex.Lock()
		if b.wsConn != nil {
			respBytes, _ := json.Marshal(resp)
			b.wsConn.WriteMessage(websocket.TextMessage, respBytes)
		}
		b.wsMutex.Unlock()
		return
	}

	var err error

	switch action.Action {
	case "send_private_msg":
		// Convert to SendMessageParams format
		msgParams := &SendMessageParams{
			MessageType: "private",
			UserID:      params.UserID,
			Message:     params.Message,
			AutoEscape:  false, // Default to false
		}
		_, err = b.SendMessage(msgParams)
	case "send_group_msg":
		// Convert to SendMessageParams format
		msgParams := &SendMessageParams{
			MessageType: "group",
			GroupID:     params.GroupID,
			Message:     params.Message,
			AutoEscape:  false, // Default to false
		}
		_, err = b.SendMessage(msgParams)
	case "send_msg":
		// Convert to SendMessageParams format
		msgParams := &SendMessageParams{
			MessageType: params.MessageType,
			UserID:      params.UserID,
			GroupID:     params.GroupID,
			Message:     params.Message,
			AutoEscape:  false, // Default to false
		}
		_, err = b.SendMessage(msgParams)
	case "get_login_info":
		user, _ := b.bot.GetCurrentUser()
		resp.Data = map[string]any{
			"user_id":  user.UserName,
			"nickname": user.NickName,
		}
	case "get_self_info":
		user, _ := b.bot.GetCurrentUser()
		resp.Data = map[string]any{
			"user_id":     user.UserName,
			"nickname":    user.NickName,
			"user_remark": user.RemarkName,
		}
	case "get_friend_list":
		friends, err := b.mySelf.Friends()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var friendList []map[string]any
		for _, f := range friends {
			friend := map[string]any{
				"user_id":  f.UserName,
				"nickname": f.NickName,
				"remark":   f.RemarkName,
			}
			friendList = append(friendList, friend)
		}
		resp.Data = map[string]any{
			"data": friendList,
		}
	case "get_group_list":
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var groupList []map[string]any
		for _, g := range groups {
			group := map[string]any{
				"group_id":   g.UserName,
				"group_name": g.NickName,
			}
			groupList = append(groupList, group)
		}
		resp.Data = map[string]any{
			"data": groupList,
		}
	case "get_group_member_list":
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var group *openwechat.Group
		for _, g := range groups {
			if g.UserName == params.GroupID {
				group = g
				break
			}
		}
		if group == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Group not found: " + params.GroupID
			break
		}
		members, err := group.Members()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var memberList []map[string]any
		for _, m := range members {
			member := map[string]any{
				"user_id":  m.UserName,
				"nickname": m.NickName,
				"card":     m.DisplayName,
			}
			memberList = append(memberList, member)
		}
		resp.Data = map[string]any{
			"data": memberList,
		}
	case "set_group_kick":
		// Check if group ID and user ID are valid
		if params.GroupID == "" || params.UserID == "" {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "group_id and user_id are required"
			break
		}
		// Try to kick user from group
		// Find the group
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Failed to get groups: " + err.Error()
			break
		}
		var targetGroup *openwechat.Group
		for _, g := range groups {
			if g.UserName == params.GroupID {
				targetGroup = g
				break
			}
		}
		if targetGroup == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Group not found: " + params.GroupID
			break
		}
		// Get group members
		members, err := targetGroup.Members()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Failed to get group members: " + err.Error()
			break
		}
		// Find the member to kick
		var memberToKick *openwechat.User
		for _, m := range members {
			if m.UserName == params.UserID {
				memberToKick = m
				break
			}
		}
		if memberToKick == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Member not found in group: " + params.UserID
			break
		}
		// Kick the member
		if err := b.mySelf.RemoveMemberFromGroup(targetGroup, openwechat.Members{memberToKick}); err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Failed to kick member: " + err.Error()
			break
		}
		resp.Data = map[string]interface{}{
			"result": true,
		}
	case "set_group_ban":
		// Check if group ID and user ID are valid
		if params.GroupID == "" || params.UserID == "" {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "group_id and user_id are required"
			break
		}
		// Try to ban user from group
		// Note: openwechat library does not support this operation
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: set_group_ban (openwechat library does not support this operation)"
	case "set_group_name":
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var group *openwechat.Group
		for _, g := range groups {
			if g.UserName == params.GroupID {
				group = g
				break
			}
		}
		if group == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Group not found: " + params.GroupID
			break
		}
		// Change group name
		err = group.Rename(params.Message)
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		resp.Data = map[string]any{
			"result": true,
		}
	case "get_version_info":
		resp.Data = map[string]any{
			"app_name":         "WxBotGo",
			"app_version":      "1.0.2",
			"protocol_version": "11",
			"onebot_version":   "11",
		}
	case "get_status":
		resp.Data = map[string]any{
			"online": true,
			"good":   true,
		}
	case "set_friend_add_request":
		// 处理好友请求
		// Check if flag is valid
		if params.Flag == "" {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "flag is required"
			break
		}
		// Try to handle friend add request
		// Note: openwechat library does not support this operation
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: set_friend_add_request (openwechat library does not support this operation)"
	case "set_group_add_request":
		// 处理群请求
		// Check if flag is valid
		if params.Flag == "" {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "flag is required"
			break
		}
		// Try to handle group add request
		// Note: openwechat library does not support this operation
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: set_group_add_request (openwechat library does not support this operation)"
	case "get_group_info":
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var group *openwechat.Group
		for _, g := range groups {
			if g.UserName == params.GroupID {
				group = g
				break
			}
		}
		if group == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Group not found: " + params.GroupID
			break
		}
		members, _ := group.Members()
		resp.Data = map[string]any{
			"group_id":     group.UserName,
			"group_name":   group.NickName,
			"member_count": len(members),
		}
	case "get_group_member_info":
		groups, err := b.mySelf.Groups()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var group *openwechat.Group
		for _, g := range groups {
			if g.UserName == params.GroupID {
				group = g
				break
			}
		}
		if group == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Group not found: " + params.GroupID
			break
		}
		members, err := group.Members()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = err.Error()
			break
		}
		var member *openwechat.User
		for _, m := range members {
			if m.UserName == params.UserID {
				member = m
				break
			}
		}
		if member == nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Member not found: " + params.UserID
			break
		}
		resp.Data = map[string]any{
			"user_id":  member.UserName,
			"nickname": member.NickName,
			"card":     member.DisplayName,
		}
	case "delete_msg":
		// 删除消息（撤回）
		// Check if message ID is valid
		if params.MessageID == "" {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "message_id is required"
			break
		}
		// Try to delete message
		b.msgMutex.Lock()
		sentMsg, ok := b.msgMap[params.MessageID]
		b.msgMutex.Unlock()
		if !ok {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Message not found: " + params.MessageID
			break
		}
		// 检查消息是否可以撤回
		if !sentMsg.CanRevoke() {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Message cannot be revoked (only messages within 2 minutes can be revoked)"
			break
		}
		// 调用openwechat库的Revoke方法撤回消息
		err = sentMsg.Revoke()
		if err != nil {
			resp.Status = "failed"
			resp.RetCode = -1
			resp.Message = "Failed to revoke message: " + err.Error()
			break
		}
		// 撤回成功后从映射表中删除
		b.msgMutex.Lock()
		delete(b.msgMap, params.MessageID)
		b.msgMutex.Unlock()
		resp.Data = map[string]interface{}{
			"result": true,
		}
	default:
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: " + action.Action
	}

	if err != nil {
		resp.Status = "failed"
		resp.RetCode = -1
		resp.Message = err.Error()
	}

	b.wsMutex.Lock()
	if b.wsConn != nil {
		respBytes, _ := json.Marshal(resp)
		b.wsConn.WriteMessage(websocket.TextMessage, respBytes)
	}
	b.wsMutex.Unlock()
}

// SendMessageParams represents the parameters for sending a message
// according to OneBot 11 protocol

type SendMessageParams struct {
	MessageType string `json:"message_type"`
	UserID      string `json:"user_id"`
	GroupID     string `json:"group_id"`
	Message     string `json:"message"`
	AutoEscape  bool   `json:"auto_escape"`
}

// SendMessage sends a message according to OneBot 11 protocol
func (b *WxBot) SendMessage(params *SendMessageParams) (*OneBotResponse, error) {
	// Check message type
	if params.MessageType != "private" && params.MessageType != "group" {
		return &OneBotResponse{
			Status:  "failed",
			RetCode: 100,
			Message: fmt.Sprintf("unsupported message type: %s", params.MessageType),
		}, nil
	}

	// Determine target ID
	targetID := params.UserID
	if params.MessageType == "group" {
		targetID = params.GroupID
	}

	// Handle auto_escape parameter
	message := params.Message
	if params.AutoEscape {
		// If auto_escape is true, we need to escape special characters
		// However, since we're sending plain text, we can just send it as-is
		// because the openwechat library handles escaping internally
	}

	// Send message to target
	msgID, sentMsg, err := b.sendText(targetID, message)
	if err != nil {
		return &OneBotResponse{
			Status:  "failed",
			RetCode: -1,
			Message: err.Error(),
		}, err
	}

	// 保存消息ID和对应的SentMessage到映射表
	b.msgMutex.Lock()
	b.msgMap[msgID] = sentMsg
	b.msgMutex.Unlock()

	return &OneBotResponse{
		Status: "ok",
<<<<<<< HEAD
		Data: map[string]interface{}{
			"message_id": msgID, // 返回真实的消息ID
<<<<<<< HEAD
=======
		Data: map[string]any{
			"message_id": "1", // WeChat doesn't return message ID, so we use a placeholder
>>>>>>> e5150c2482302cbf3c41db97a64afb3fe5c878df
=======
>>>>>>> 455bcfd (Finalize all changes: WeComBot to WeWorkBot renaming, WxBotGo fixes, DEPLOY_CN.md)
		},
	}, nil
}

func (b *WxBot) sendText(targetID string, text string) (string, *openwechat.SentMessage, error) {
	friends, err := b.mySelf.Friends()
	if err == nil {
		for _, f := range friends {
			if f.UserName == targetID {
				sentMsg, err := f.SendText(text)
				if err != nil {
					return "", nil, err
				}
				return sentMsg.MsgId, sentMsg, nil
			}
		}
	}

	groups, err := b.mySelf.Groups()
	if err == nil {
		for _, g := range groups {
			if g.UserName == targetID {
				sentMsg, err := g.SendText(text)
				if err != nil {
					return "", nil, err
				}
				return sentMsg.MsgId, sentMsg, nil
			}
		}
	}

	// 尝试发送到公众号或其他类型
	contacts, err := b.mySelf.Members()
	if err == nil {
		for _, c := range contacts {
			if c.UserName == targetID {
				// 检查是否为公众号
				if mp, ok := c.AsMP(); ok {
					sentMsg, err := mp.SendText(text)
					if err != nil {
						return "", nil, err
					}
					return sentMsg.MsgId, sentMsg, nil
				}
				// 检查是否为好友
				if friend, ok := c.AsFriend(); ok {
					sentMsg, err := friend.SendText(text)
					if err != nil {
						return "", nil, err
					}
					return sentMsg.MsgId, sentMsg, nil
				}
				// 检查是否为群
				if group, ok := c.AsGroup(); ok {
					sentMsg, err := group.SendText(text)
					if err != nil {
						return "", nil, err
					}
					return sentMsg.MsgId, sentMsg, nil
				}
			}
		}
	}

	return "", nil, fmt.Errorf("target not found: %s", targetID)
}
