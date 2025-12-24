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
	ManagerURL string
	SelfID     string

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
		ManagerURL: managerUrl,
		SelfID:     selfId,
		callback:   cb,
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
		identify := map[string]interface{}{
			"type":            "meta_event",
			"meta_event_type": "lifecycle",
			"sub_type":        "connect",
			"self_id":         b.SelfID,
			"platform":        "wechat-go",
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
					heartbeat := map[string]interface{}{
						"type":            "meta_event",
						"meta_event_type": "heartbeat",
						"time":            time.Now().Unix(),
						"self_id":         b.SelfID,
						"status": map[string]interface{}{
							"online": true,
							"good":   true,
						},
					}
					b.wsConn.WriteJSON(heartbeat)
					b.wsMutex.Unlock()
				}
			}
		}()

		// Listen for commands
		for {
			_, message, err := b.wsConn.ReadMessage()
			if err != nil {
				b.Log("[WebSocket] Read error: %v", err)
				break
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

// Handle WeChat Message -> OneBot Event
func (b *WxBot) HandleWeChatMsg(msg *openwechat.Message) {
	if msg.IsSendBySelf() {
		return
	}

	event := OneBotEvent{
		PostType: "message",
	}

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

	if msg.IsText() {
		event.Message = msg.Content
		event.RawMessage = msg.Content
	} else if msg.IsPicture() {
		event.Message = "[图片]"
		event.RawMessage = "[图片]"
	} else {
		return
	}

	event.MessageID = msg.MsgId
	b.sendEvent(event)
}

// Handle OneBot Action -> WeChat API
func (b *WxBot) HandleAction(action OneBotAction) {
	resp := OneBotResponse{
		Status: "ok",
		Echo:   action.Echo,
		Data:   map[string]interface{}{},
	}

	bytes, _ := json.Marshal(action.Params)
	var params ActionParams
	json.Unmarshal(bytes, &params)

	var err error

	switch action.Action {
	case "send_private_msg":
		err = b.sendText(params.UserID, params.Message)
	case "send_group_msg":
		err = b.sendText(params.GroupID, params.Message)
	case "send_msg":
		if params.MessageType == "group" {
			err = b.sendText(params.GroupID, params.Message)
		} else {
			err = b.sendText(params.UserID, params.Message)
		}
	case "get_login_info":
		user, _ := b.bot.GetCurrentUser()
		resp.Data = map[string]interface{}{
			"user_id":  user.UserName,
			"nickname": user.NickName,
		}
	case "get_self_info":
		user, _ := b.bot.GetCurrentUser()
		resp.Data = map[string]interface{}{
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
		var friendList []map[string]interface{}
		for _, f := range friends {
			friend := map[string]interface{}{
				"user_id":  f.UserName,
				"nickname": f.NickName,
				"remark":   f.RemarkName,
			}
			friendList = append(friendList, friend)
		}
		resp.Data = map[string]interface{}{
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
		var groupList []map[string]interface{}
		for _, g := range groups {
			group := map[string]interface{}{
				"group_id":   g.UserName,
				"group_name": g.NickName,
			}
			groupList = append(groupList, group)
		}
		resp.Data = map[string]interface{}{
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
		var memberList []map[string]interface{}
		for _, m := range members {
			member := map[string]interface{}{
				"user_id":  m.UserName,
				"nickname": m.NickName,
				"card":     m.DisplayName,
			}
			memberList = append(memberList, member)
		}
		resp.Data = map[string]interface{}{
			"data": memberList,
		}
	case "set_group_kick":
		resp.Status = "failed"
		resp.RetCode = 100
		resp.Message = "Unsupported action: set_group_kick (openwechat library does not support this operation)"
	case "set_group_ban":
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

	return fmt.Errorf("target not found: %s", targetID)
}
