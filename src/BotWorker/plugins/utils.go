package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var GlobalDB *sql.DB

func SetGlobalDB(database *sql.DB) {
	GlobalDB = database
}

var GlobalRedis *redis.Client

func SetGlobalRedis(client *redis.Client) {
	GlobalRedis = client
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type PendingConfirmation struct {
	GroupID     int64
	UserID      int64
	Action      string
	Params      map[string]string
	ConfirmCode string
	CancelCode  string
	ExpiresAt   time.Time
}

type PendingDialog struct {
	GroupID   int64
	UserID    int64
	Type      string
	Step      int
	Data      map[string]string
	ExpiresAt time.Time
}

var (
	pendingConfirmationsMu sync.Mutex
	pendingConfirmations   = make(map[string]*PendingConfirmation)

	pendingDialogsMu sync.Mutex
	pendingDialogs   = make(map[string]*PendingDialog)
)

var FeatureDefaults = map[string]bool{
	"weather":                 true,
	"points":                  true,
	"lottery":                 true,
	"translate":               true,
	"music":                   true,
	"games":                   true,
	"greetings":               true,
	"utils":                   true,
	"moderation":              true,
	"pets":                    true,
	"welcome":                 true,
	"kick_to_black":           true,
	"kick_notify":             true,
	"leave_to_black":          true,
	"leave_notify":            true,
	"join_mute":               false,
	"signin":                  true,
	"voice_reply":             false,
	"burn_after_reading":      false,
	"feature_disabled_notice": false,
	"tarot":                   true,
	"plugin_manager":          true,
	"gift":                    true,
	"fishing":                 true,
	"cultivation":             true,
	"farm":                    true,
	"robbery":                 true,
	"word_guess":              true,
	"idiom_guess":             true,
	"auction":                 true,
	"badge":                   true,
	"medal":                   true,
}

var FeatureDisplayNames = map[string]string{
	"weather":            "天气",
	"points":             "积分",
	"signin":             "签到",
	"lottery":            "抽签",
	"translate":          "翻译",
	"music":              "点歌",
	"games":              "游戏",
	"greetings":          "问候",
	"utils":              "工具",
	"moderation":         "群管",
	"pets":               "宠物",
	"welcome":            "欢迎语",
	"kick_to_black":      "踢出拉黑",
	"kick_notify":        "被踢提示",
	"leave_to_black":     "退群拉黑",
	"leave_notify":       "退群提示",
	"join_mute":          "进群禁言",
	"voice_reply":        "语音回复",
	"burn_after_reading": "阅后即焚",
	"tarot":              "塔罗牌",
	"plugin_manager":     "插件管理",
	"gift":               "礼物",
	"fishing":            "钓鱼",
	"cultivation":        "修炼",
	"farm":               "农场",
	"robbery":            "打劫",
	"word_guess":         "猜单词",
	"idiom_guess":        "猜成语",
	"auction":            "竞拍系统",
	"badge":              "徽章系统",
	"medal":              "勋章系统",
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsFeatureEnabledForGroup(database *sql.DB, groupID string, featureID string) bool {
	defaultEnabled, ok := FeatureDefaults[featureID]
	if !ok {
		return false
	}

	if database == nil || groupID == "" {
		return defaultEnabled
	}

	enabled, hasOverride, err := db.GetGroupFeatureOverride(database, groupID, featureID)
	if err != nil || !hasOverride {
		return defaultEnabled
	}

	return enabled
}

func HandleFeatureDisabled(robot plugin.Robot, event *onebot.Event, featureID string) {
	if event == nil || event.MessageType != "group" {
		return
	}

	if GlobalDB == nil {
		return
	}

	groupID := fmt.Sprintf("%d", event.GroupID)
	userID := fmt.Sprintf("%d", event.UserID)

	if IsFeatureEnabledForGroup(GlobalDB, groupID, "feature_disabled_notice") {
		displayName, ok := FeatureDisplayNames[featureID]
		if !ok || displayName == "" {
			displayName = featureID
		}
		message := fmt.Sprintf("%s功能已关闭", displayName)
		params := &onebot.SendMessageParams{
			GroupID: event.GroupID,
			UserID:  event.UserID,
			Message: message,
		}

		if _, err := robot.SendMessage(params); err != nil {
			log.Printf("发送功能关闭提示失败: %v\n", err)
		}
		return
	}

	messageID := fmt.Sprintf("no-reply-%s-%s-%s-%d", featureID, groupID, userID, time.Now().UnixNano())
	content := fmt.Sprintf("no_reply: feature %s disabled for group %s", featureID, groupID)

	record := &db.Message{
		MessageID: messageID,
		UserID:    userID,
		GroupID:   groupID,
		Type:      "no_reply",
		Content:   content,
	}

	if err := db.CreateMessage(GlobalDB, record); err != nil {
		log.Printf("记录无回复原因失败: %v\n", err)
	}
}

func SendAIVoiceMessage(robot plugin.Robot, event *onebot.Event, message string) (*onebot.Response, error) {
	if event == nil {
		params := &onebot.SendMessageParams{
			Message: message,
		}
		return robot.SendMessage(params)
	}

	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	return robot.SendMessage(params)
}

func SendTextReply(robot plugin.Robot, event *onebot.Event, message string) (*onebot.Response, error) {
	if event == nil {
		params := &onebot.SendMessageParams{
			Message: message,
		}
		return robot.SendMessage(params)
	}

	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if event.MessageType == "group" && GlobalDB != nil {
		groupID := fmt.Sprintf("%d", event.GroupID)

		if IsFeatureEnabledForGroup(GlobalDB, groupID, "voice_reply") {
			return SendAIVoiceMessage(robot, event, message)
		}

		resp, err := robot.SendMessage(params)
		if err != nil {
			return resp, err
		}

		if IsFeatureEnabledForGroup(GlobalDB, groupID, "burn_after_reading") {
			scheduleAutoDelete(robot, resp)
		}

		return resp, nil
	}

	return robot.SendMessage(params)
}

func scheduleAutoDelete(robot plugin.Robot, resp *onebot.Response) {
	if resp == nil || resp.Data == nil {
		return
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return
	}

	rawID, ok := data["message_id"]
	if !ok {
		return
	}

	var msgID int64

	switch v := rawID.(type) {
	case int64:
		msgID = v
	case int:
		msgID = int64(v)
	case float64:
		msgID = int64(v)
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			msgID = id
		}
	}

	if msgID == 0 {
		return
	}

	delay := 10 * time.Second

	go func() {
		time.Sleep(delay)
		_, err := robot.DeleteMessage(&onebot.DeleteMessageParams{
			MessageID: msgID,
		})
		if err != nil {
			log.Printf("自动撤回消息失败: %v\n", err)
		}
	}()
}

func makeSessionKey(groupID, userID int64) string {
	return fmt.Sprintf("%d:%d", groupID, userID)
}

func StartConfirmation(action string, event *onebot.Event, confirmCode, cancelCode string, params map[string]string, ttl time.Duration) *PendingConfirmation {
	if event == nil {
		return nil
	}

	if confirmCode == "" {
		confirmCode = fmt.Sprintf("%03d", rand.Intn(900)+100)
	}
	if cancelCode == "" {
		for {
			code := fmt.Sprintf("%03d", rand.Intn(900)+100)
			if code != confirmCode {
				cancelCode = code
				break
			}
		}
	}
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}

	pc := &PendingConfirmation{
		GroupID:     event.GroupID,
		UserID:      event.UserID,
		Action:      action,
		Params:      params,
		ConfirmCode: confirmCode,
		CancelCode:  cancelCode,
		ExpiresAt:   time.Now().Add(ttl),
	}

	key := makeSessionKey(pc.GroupID, pc.UserID)

	if GlobalRedis != nil {
		ctx := context.Background()
		data, err := json.Marshal(pc)
		if err == nil {
			_ = GlobalRedis.Set(ctx, "bot:confirm:"+key, data, ttl).Err()
		}
	} else if GlobalDB != nil {
		sessionID := "confirm:" + key
		userIDStr := fmt.Sprintf("%d", pc.UserID)
		groupIDStr := fmt.Sprintf("%d", pc.GroupID)
		data := map[string]interface{}{
			"action":       pc.Action,
			"params":       pc.Params,
			"confirm_code": pc.ConfirmCode,
			"cancel_code":  pc.CancelCode,
			"expires_at":   pc.ExpiresAt.Unix(),
		}
		session := &db.Session{
			SessionID: sessionID,
			UserID:    userIDStr,
			GroupID:   groupIDStr,
			State:     "confirm:" + action,
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(GlobalDB, session)
	} else {
		pendingConfirmationsMu.Lock()
		pendingConfirmations[key] = pc
		pendingConfirmationsMu.Unlock()
	}

	return pc
}

func GetPendingConfirmation(groupID, userID int64) *PendingConfirmation {
	key := makeSessionKey(groupID, userID)

	if GlobalRedis != nil {
		ctx := context.Background()
		val, err := GlobalRedis.Get(ctx, "bot:confirm:"+key).Bytes()
		if err != nil || len(val) == 0 {
			return nil
		}

		var pc PendingConfirmation
		if err := json.Unmarshal(val, &pc); err != nil {
			return nil
		}

		if time.Now().After(pc.ExpiresAt) {
			_ = GlobalRedis.Del(ctx, "bot:confirm:"+key).Err()
			return nil
		}

		return &pc
	}

	if GlobalDB != nil {
		sessionID := "confirm:" + key
		session, err := db.GetSessionBySessionID(GlobalDB, sessionID)
		if err != nil || session == nil {
			return nil
		}

		data := session.Data
		if data == nil {
			_ = db.DeleteSession(GlobalDB, sessionID)
			return nil
		}

		action, _ := data["action"].(string)
		confirmCode, _ := data["confirm_code"].(string)
		cancelCode, _ := data["cancel_code"].(string)

		var expiresAt time.Time
		if sec, ok := data["expires_at"].(float64); ok {
			expiresAt = time.Unix(int64(sec), 0)
		}

		if expiresAt.IsZero() || time.Now().After(expiresAt) {
			_ = db.DeleteSession(GlobalDB, sessionID)
			return nil
		}

		params := map[string]string{}
		if raw, ok := data["params"]; ok {
			if m, ok := raw.(map[string]interface{}); ok {
				for k, v := range m {
					if s, ok := v.(string); ok {
						params[k] = s
					} else {
						params[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		groupIDStr := session.GroupID
		userIDStr := session.UserID
		gid, _ := strconv.ParseInt(groupIDStr, 10, 64)
		uid, _ := strconv.ParseInt(userIDStr, 10, 64)

		pc := &PendingConfirmation{
			GroupID:     gid,
			UserID:      uid,
			Action:      action,
			Params:      params,
			ConfirmCode: confirmCode,
			CancelCode:  cancelCode,
			ExpiresAt:   expiresAt,
		}

		return pc
	}

	pendingConfirmationsMu.Lock()
	defer pendingConfirmationsMu.Unlock()

	pc, ok := pendingConfirmations[key]
	if !ok {
		return nil
	}

	if time.Now().After(pc.ExpiresAt) {
		delete(pendingConfirmations, key)
		return nil
	}

	return pc
}

func ClearPendingConfirmation(groupID, userID int64) {
	key := makeSessionKey(groupID, userID)

	if GlobalRedis != nil {
		ctx := context.Background()
		_ = GlobalRedis.Del(ctx, "bot:confirm:"+key).Err()
		return
	} else if GlobalDB != nil {
		sessionID := "confirm:" + key
		_ = db.DeleteSession(GlobalDB, sessionID)
		return
	}

	pendingConfirmationsMu.Lock()
	delete(pendingConfirmations, key)
	pendingConfirmationsMu.Unlock()
}

type ConfirmationResult struct {
	Matched   bool
	Confirmed bool
	Canceled  bool
	Action    string
	Params    map[string]string
}

func HandleConfirmationReply(event *onebot.Event) *ConfirmationResult {
	if event == nil {
		return nil
	}

	pc := GetPendingConfirmation(event.GroupID, event.UserID)
	if pc == nil {
		return nil
	}

	text := event.RawMessage
	if text == "" {
		if msg, ok := event.Message.(string); ok {
			text = msg
		}
	}

	if text == "" {
		return nil
	}

	if text == pc.ConfirmCode {
		ClearPendingConfirmation(event.GroupID, event.UserID)
		return &ConfirmationResult{
			Matched:   true,
			Confirmed: true,
			Action:    pc.Action,
			Params:    pc.Params,
		}
	}

	if text == pc.CancelCode {
		ClearPendingConfirmation(event.GroupID, event.UserID)
		return &ConfirmationResult{
			Matched:  true,
			Canceled: true,
			Action:   pc.Action,
			Params:   pc.Params,
		}
	}

	return nil
}

func StartDialog(dialogType string, event *onebot.Event, ttl time.Duration) *PendingDialog {
	if event == nil {
		return nil
	}

	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	pd := &PendingDialog{
		GroupID:   event.GroupID,
		UserID:    event.UserID,
		Type:      dialogType,
		Step:      1,
		Data:      make(map[string]string),
		ExpiresAt: time.Now().Add(ttl),
	}

	key := makeSessionKey(pd.GroupID, pd.UserID)
	if GlobalRedis != nil {
		ctx := context.Background()
		data, err := json.Marshal(pd)
		if err == nil {
			_ = GlobalRedis.Set(ctx, "bot:dialog:"+key, data, ttl).Err()
		}
	} else if GlobalDB != nil {
		sessionID := "dialog:" + key
		userIDStr := fmt.Sprintf("%d", pd.UserID)
		groupIDStr := fmt.Sprintf("%d", pd.GroupID)
		data := map[string]interface{}{
			"type":       pd.Type,
			"step":       pd.Step,
			"data":       pd.Data,
			"expires_at": pd.ExpiresAt.Unix(),
		}
		session := &db.Session{
			SessionID: sessionID,
			UserID:    userIDStr,
			GroupID:   groupIDStr,
			State:     "dialog:" + dialogType,
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(GlobalDB, session)
	} else {
		pendingDialogsMu.Lock()
		pendingDialogs[key] = pd
		pendingDialogsMu.Unlock()
	}

	return pd
}

func GetDialog(groupID, userID int64) *PendingDialog {
	key := makeSessionKey(groupID, userID)

	if GlobalRedis != nil {
		ctx := context.Background()
		val, err := GlobalRedis.Get(ctx, "bot:dialog:"+key).Bytes()
		if err != nil || len(val) == 0 {
			return nil
		}

		var pd PendingDialog
		if err := json.Unmarshal(val, &pd); err != nil {
			return nil
		}

		if time.Now().After(pd.ExpiresAt) {
			_ = GlobalRedis.Del(ctx, "bot:dialog:"+key).Err()
			return nil
		}

		return &pd
	}

	if GlobalDB != nil {
		sessionID := "dialog:" + key
		session, err := db.GetSessionBySessionID(GlobalDB, sessionID)
		if err != nil || session == nil {
			return nil
		}

		data := session.Data
		if data == nil {
			_ = db.DeleteSession(GlobalDB, sessionID)
			return nil
		}

		var expiresAt time.Time
		if sec, ok := data["expires_at"].(float64); ok {
			expiresAt = time.Unix(int64(sec), 0)
		}

		if expiresAt.IsZero() || time.Now().After(expiresAt) {
			_ = db.DeleteSession(GlobalDB, sessionID)
			return nil
		}

		dialogType, _ := data["type"].(string)
		step := 1
		if v, ok := data["step"].(float64); ok {
			step = int(v)
		}

		dialogData := map[string]string{}
		if raw, ok := data["data"]; ok {
			if m, ok := raw.(map[string]interface{}); ok {
				for k, v := range m {
					if s, ok := v.(string); ok {
						dialogData[k] = s
					} else {
						dialogData[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		groupIDStr := session.GroupID
		userIDStr := session.UserID
		gid, _ := strconv.ParseInt(groupIDStr, 10, 64)
		uid, _ := strconv.ParseInt(userIDStr, 10, 64)

		pd := &PendingDialog{
			GroupID:   gid,
			UserID:    uid,
			Type:      dialogType,
			Step:      step,
			Data:      dialogData,
			ExpiresAt: expiresAt,
		}

		return pd
	}

	pendingDialogsMu.Lock()
	defer pendingDialogsMu.Unlock()

	pd, ok := pendingDialogs[key]
	if !ok {
		return nil
	}

	if time.Now().After(pd.ExpiresAt) {
		delete(pendingDialogs, key)
		return nil
	}

	return pd
}

func UpdateDialog(pd *PendingDialog, nextStep int, ttl time.Duration) {
	if pd == nil {
		return
	}

	if nextStep > 0 {
		pd.Step = nextStep
	}
	if ttl > 0 {
		pd.ExpiresAt = time.Now().Add(ttl)
	}

	key := makeSessionKey(pd.GroupID, pd.UserID)

	if GlobalRedis != nil {
		ctx := context.Background()
		data, err := json.Marshal(pd)
		if err == nil {
			if ttl <= 0 {
				ttl = 5 * time.Minute
			}
			_ = GlobalRedis.Set(ctx, "bot:dialog:"+key, data, ttl).Err()
		}
	} else if GlobalDB != nil {
		if ttl <= 0 {
			ttl = 5 * time.Minute
		}
		sessionID := "dialog:" + key
		userIDStr := fmt.Sprintf("%d", pd.UserID)
		groupIDStr := fmt.Sprintf("%d", pd.GroupID)
		data := map[string]interface{}{
			"type":       pd.Type,
			"step":       pd.Step,
			"data":       pd.Data,
			"expires_at": pd.ExpiresAt.Unix(),
		}
		session := &db.Session{
			SessionID: sessionID,
			UserID:    userIDStr,
			GroupID:   groupIDStr,
			State:     "dialog:" + pd.Type,
			Data:      data,
		}
		_ = db.CreateOrUpdateSession(GlobalDB, session)
	} else {
		pendingDialogsMu.Lock()
		pendingDialogs[key] = pd
		pendingDialogsMu.Unlock()
	}
}

func EndDialog(groupID, userID int64) {
	key := makeSessionKey(groupID, userID)

	if GlobalRedis != nil {
		ctx := context.Background()
		_ = GlobalRedis.Del(ctx, "bot:dialog:"+key).Err()
		return
	} else if GlobalDB != nil {
		sessionID := "dialog:" + key
		_ = db.DeleteSession(GlobalDB, sessionID)
		return
	}

	pendingDialogsMu.Lock()
	delete(pendingDialogs, key)
	pendingDialogsMu.Unlock()
}
