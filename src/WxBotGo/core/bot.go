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

type WxBot struct {
	ManagerURL    string
	SelfID        string
	ReportSelfMsg bool // 是否上报自身消息

	wsConn  *websocket.Conn
	wsMutex sync.Mutex
	mySelf  *openwechat.Self
	bot     *openwechat.Bot

	callback BotCallback
}

func NewWxBot(managerUrl, selfId string, cb BotCallback) *WxBot {
	if managerUrl == "" {
		managerUrl = "ws://localhost:3001"
	}
	if selfId == "" {
		selfId = "1098299491"
	}
	return &WxBot{
		ManagerURL:    managerUrl,
		SelfID:        selfId,
		ReportSelfMsg: true, // 默认上报自身消息
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

func (b *WxBot) Start() {
	b.Log("[WxBotGo] Starting... Target Manager: %s, SelfID: %s", b.ManagerURL, b.SelfID)

	b.bot = openwechat.DefaultBot(openwechat.Desktop)

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
				event.UserID = groupSender.UserName
				event.Sender = &Sender{
					UserID:   groupSender.UserName,
					Nickname: groupSender.NickName,
				}
			} else {
				event.UserID = sender.UserName
				event.Sender = &Sender{UserID: sender.UserName, Nickname: sender.NickName}
			}

			group := sender
			event.GroupID = group.UserName

		} else {
			event.MessageType = "private"
			event.SubType = "friend"
			event.UserID = sender.UserName
			event.Sender = &Sender{
				UserID:   sender.UserName,
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

	event.MessageID = msg.MsgId
	return event
}

// Handle WeChat Message -> OneBot Event
func (b *WxBot) HandleWeChatMsg(msg *openwechat.Message) {
	// 如果是自身消息且未开启上报，则忽略
	if msg.IsSendBySelf() && !b.ReportSelfMsg {
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
		// Note: openwechat library does not support this operation
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: set_group_kick (openwechat library does not support this operation)"
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
		// Note: openwechat library does not support deleting messages by ID
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: delete_msg (openwechat library does not support deleting messages by ID)"
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
	err := b.sendText(targetID, message)
	if err != nil {
		return &OneBotResponse{
			Status:  "failed",
			RetCode: -1,
			Message: err.Error(),
		}, err
	}

	return &OneBotResponse{
		Status: "ok",
		Data: map[string]any{
			"message_id": "1", // WeChat doesn't return message ID, so we use a placeholder
		},
	}, nil
}

func (b *WxBot) sendText(targetID string, text string) error {
	friends, err := b.mySelf.Friends()
	if err == nil {
		for _, f := range friends {
			if f.UserName == targetID {
				_, err = f.SendText(text)
				return err
			}
		}
	}

	groups, err := b.mySelf.Groups()
	if err == nil {
		for _, g := range groups {
			if g.UserName == targetID {
				_, err = g.SendText(text)
				return err
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
					_, err = mp.SendText(text)
					return err
				}
				// 检查是否为好友
				if friend, ok := c.AsFriend(); ok {
					_, err = friend.SendText(text)
					return err
				}
				// 检查是否为群
				if group, ok := c.AsGroup(); ok {
					_, err = group.SendText(text)
					return err
				}
			}
		}
	}

	return fmt.Errorf("target not found: %s", targetID)
}
