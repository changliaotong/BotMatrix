package plugins

import (
	"BotMatrix/common"
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
	"regexp"
	"strconv"
	"strings"
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

func GetFeatureDisplayName(featureID string) string {
	return common.T("", "feature_"+featureID)
}

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
	"ai":                      true,
	"knowledge_base":          true,
}

var FeatureDisplayNames = map[string]string{
	"weather":            common.T("", "feature_weather"),
	"points":             common.T("", "feature_points"),
	"signin":             common.T("", "feature_signin"),
	"lottery":            common.T("", "feature_lottery"),
	"translate":          common.T("", "feature_translate"),
	"music":              common.T("", "feature_music"),
	"games":              common.T("", "feature_games"),
	"greetings":          common.T("", "feature_greetings"),
	"utils":              common.T("", "feature_utils"),
	"moderation":         common.T("", "feature_moderation"),
	"pets":               common.T("", "feature_pets"),
	"welcome":            common.T("", "feature_welcome"),
	"kick_to_black":      common.T("", "feature_kick_to_black"),
	"kick_notify":        common.T("", "feature_kick_notify"),
	"leave_to_black":     common.T("", "feature_leave_to_black"),
	"leave_notify":       common.T("", "feature_leave_notify"),
	"join_mute":          common.T("", "feature_join_mute"),
	"voice_reply":        common.T("", "feature_voice_reply"),
	"burn_after_reading": common.T("", "feature_burn_after_reading"),
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
	"ai":                 common.T("", "feature_ai"),
	"knowledge_base":     common.T("", "feature_knowledge_base"),
}

type VoiceItem struct {
	ID         string
	Name       string
	PreviewURL string
}

type VoiceCategory struct {
	Name  string
	Items []VoiceItem
}

var VoiceCategories = []VoiceCategory{
	{
		Name: "推荐",
		Items: []VoiceItem{
			{ID: "lucy-voice-laibixiaoxin", Name: "小新", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-laibixiaoxin.wav"},
			{ID: "lucy-voice-houge", Name: "猴哥", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-houge.wav"},
			{ID: "lucy-voice-silang", Name: "四郎", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-silang.wav"},
			{ID: "lucy-voice-guangdong-f1", Name: "东北老妹儿", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangdong-f1.wav"},
			{ID: "lucy-voice-guangxi-m1", Name: "广西大表哥", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangxi-m1.wav"},
			{ID: "lucy-voice-daji", Name: "妲己", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-daji.wav"},
			{ID: "lucy-voice-lizeyan", Name: "霸道总裁", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lizeyan-2.wav"},
			{ID: "lucy-voice-suxinjiejie", Name: "酥心御姐", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-suxinjiejie.wav"},
		},
	},
	{
		Name: "搞怪",
		Items: []VoiceItem{
			{ID: "lucy-voice-laibixiaoxin", Name: "小新", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-laibixiaoxin.wav"},
			{ID: "lucy-voice-houge", Name: "猴哥", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-houge.wav"},
			{ID: "lucy-voice-guangdong-f1", Name: "东北老妹儿", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangdong-f1.wav"},
			{ID: "lucy-voice-guangxi-m1", Name: "广西大表哥", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-guangxi-m1.wav"},
			{ID: "lucy-voice-m8", Name: "说书先生", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m8.wav"},
			{ID: "lucy-voice-male1", Name: "憨憨小弟", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male1.wav"},
			{ID: "lucy-voice-male3", Name: "憨厚老哥", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male3.wav"},
		},
	},
	{
		Name: "古风",
		Items: []VoiceItem{
			{ID: "lucy-voice-daji", Name: "妲己", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-daji.wav"},
			{ID: "lucy-voice-silang", Name: "四郎", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-silang.wav"},
			{ID: "lucy-voice-lvbu", Name: "吕布", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lvbu.wav"},
		},
	},
	{
		Name: "现代",
		Items: []VoiceItem{
			{ID: "lucy-voice-lizeyan", Name: "霸道总裁", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-lizeyan-2.wav"},
			{ID: "lucy-voice-suxinjiejie", Name: "酥心御姐", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-suxinjiejie.wav"},
			{ID: "lucy-voice-xueling", Name: "元气少女", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-xueling.wav"},
			{ID: "lucy-voice-f37", Name: "文艺少女", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f37.wav"},
			{ID: "lucy-voice-male2", Name: "磁性大叔", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-male2.wav"},
			{ID: "lucy-voice-female1", Name: "邻家小妹", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-female1.wav"},
			{ID: "lucy-voice-m14", Name: "低沉男声", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m14.wav"},
			{ID: "lucy-voice-f38", Name: "傲娇少女", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f38.wav"},
			{ID: "lucy-voice-m101", Name: "爹系男友", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-m101.wav"},
			{ID: "lucy-voice-female2", Name: "暖心姐姐", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-female2.wav"},
			{ID: "lucy-voice-f36", Name: "温柔妹妹", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f36.wav"},
			{ID: "lucy-voice-f34", Name: "书香少女", PreviewURL: "https://res.qpt.qq.com/qpilot/tts_sample/group/lucy-voice-f34.wav"},
		},
	},
}

var (
	voiceNameToID = map[string]string{}
	allVoices     []VoiceItem
)

func init() {
	rand.Seed(time.Now().UnixNano())
	initVoiceData()
}

func initVoiceData() {
	voiceNameToID = make(map[string]string)
	allVoices = allVoices[:0]

	for _, cat := range VoiceCategories {
		for _, item := range cat.Items {
			allVoices = append(allVoices, item)
			if item.Name != "" {
				if _, ok := voiceNameToID[item.Name]; !ok {
					voiceNameToID[item.Name] = item.ID
				}
			}
		}
	}
}

func BuildVoiceList(currentID string) string {
	var b strings.Builder

	if currentID != "" {
		current := findVoiceByID(currentID)
		if current != nil {
			b.WriteString(common.T("", "voice_current_prefix"))
			b.WriteString(current.Name)
		} else {
			b.WriteString(common.T("", "voice_current_prefix"))
			b.WriteString(currentID)
		}
		b.WriteString("\n\n")
	}

	b.WriteString(common.T("", "voice_list_title"))
	b.WriteString("\n")

	index := 1
	for _, cat := range VoiceCategories {
		b.WriteString("[")
		b.WriteString(cat.Name)
		b.WriteString("]\n")
		for _, item := range cat.Items {
			b.WriteString(fmt.Sprintf("%d. %s\n", index, item.Name))
			index++
		}
		b.WriteString("\n")
	}

	return strings.TrimSpace(b.String())
}

func FindVoiceByGlobalIndex(index int) *VoiceItem {
	if index <= 0 {
		return nil
	}

	cur := 1
	for _, cat := range VoiceCategories {
		for i := range cat.Items {
			if cur == index {
				return &cat.Items[i]
			}
			cur++
		}
	}

	return nil
}

func findVoiceByID(id string) *VoiceItem {
	if id == "" {
		return nil
	}

	for _, cat := range VoiceCategories {
		for i := range cat.Items {
			if cat.Items[i].ID == id {
				return &cat.Items[i]
			}
		}
	}

	return nil
}

func FindVoiceByName(name string) *VoiceItem {
	if name == "" {
		return nil
	}

	id, ok := voiceNameToID[name]
	if !ok {
		return nil
	}

	return findVoiceByID(id)
}

func FindVoiceFuzzy(keyword string) *VoiceItem {
	if keyword == "" {
		return nil
	}

	kw := strings.ToLower(keyword)
	for _, v := range allVoices {
		if strings.Contains(strings.ToLower(v.Name), kw) {
			return &v
		}
	}

	return nil
}

func GetRandomVoice() *VoiceItem {
	if len(allVoices) == 0 {
		return nil
	}
	return &allVoices[rand.Intn(len(allVoices))]
}

func GetVoiceCategoriesForID(voiceID string) []string {
	if voiceID == "" {
		return nil
	}

	var names []string
	for _, cat := range VoiceCategories {
		for _, item := range cat.Items {
			if item.ID == voiceID {
				names = append(names, cat.Name)
				break
			}
		}
	}
	return names
}

func GetVoicePreviewURL(voiceID string) string {
	item := findVoiceByID(voiceID)
	if item == nil {
		return ""
	}
	return item.PreviewURL
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
		displayName := GetFeatureDisplayName(featureID)
		message := fmt.Sprintf(common.T("", "feature_disabled_msg"), displayName)
		params := &onebot.SendMessageParams{
			GroupID: event.GroupID,
			UserID:  event.UserID,
			Message: message,
		}

		if _, err := robot.SendMessage(params); err != nil {
			log.Printf(common.T("", "feature_disabled_notice_failed"), err)
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
		log.Printf(common.T("", "feature_disabled_record_failed"), err)
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
			log.Printf(common.T("", "utils_auto_delete_failed"), err)
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

func NormalizeQuestion(text string) string {
	if text == "" {
		return ""
	}
	s := strings.TrimSpace(text)
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "")
	s = replacer.Replace(s)
	return s
}

func SubstituteSystemVariables(text string, event *onebot.Event) string {
	if text == "" || event == nil {
		return text
	}

	now := time.Now()

	userName := fmt.Sprintf("%d", event.UserID)
	if event.Sender.Card != "" {
		userName = event.Sender.Card
	} else if event.Sender.Nickname != "" {
		userName = event.Sender.Nickname
	}

	selfName := "机器人"
	groupName := ""

	groupIDStr := ""
	if event.GroupID != 0 {
		groupIDStr = fmt.Sprintf("%d", event.GroupID)
	}

	userIDStr := fmt.Sprintf("%d", event.UserID)

	pointsStr := ""
	if GlobalDB != nil && userIDStr != "0" {
		if pts, err := db.GetPoints(GlobalDB, userIDStr); err == nil {
			pointsStr = strconv.Itoa(pts)
		}
	}

	yearStr := fmt.Sprintf("%d", now.Year())
	monthStr := fmt.Sprintf("%d", int(now.Month()))
	dayStr := fmt.Sprintf("%d", now.Day())
	hourStr := fmt.Sprintf("%d", now.Hour())
	minuteStr := fmt.Sprintf("%d", now.Minute())
	secondStr := fmt.Sprintf("%d", now.Second())

	weekdayNames := []string{"日", "一", "二", "三", "四", "五", "六"}
	weekday := weekdayNames[int(now.Weekday())]

	replacements := map[string]string{
		"#你#":   userName,
		"{你}":   userName,
		"#我#":   selfName,
		"{我}":   selfName,
		"#群#":   groupName,
		"{群}":   groupName,
		"#群主#": "",
		"{群主}": "",
		"#主人#": "",
		"{主人}": "",
		"#积分#": pointsStr,
		"{积分}": pointsStr,
		"#年#":  yearStr,
		"{年}":  yearStr,
		"#月#":  monthStr,
		"{月}":  monthStr,
		"#日#":  dayStr,
		"{日}":  dayStr,
		"#时#":  hourStr,
		"{时}":  hourStr,
		"#分#":  minuteStr,
		"{分}":  minuteStr,
		"#秒#":  secondStr,
		"{秒}":  secondStr,
		"#群号#": groupIDStr,
		"{群号}": groupIDStr,
		"#星期#": "星期" + weekday,
		"{星期}": "星期" + weekday,
	}

	result := text
	for k, v := range replacements {
		if v == "" {
			continue
		}
		result = strings.ReplaceAll(result, k, v)
	}

	return result
}

func SubstituteCustomVariables(text string, event *onebot.Event) string {
	if text == "" || event == nil || GlobalDB == nil {
		return text
	}

	if event.GroupID == 0 {
		return text
	}

	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	re := regexp.MustCompile(`\{\{([^{}]+)\}\}`)

	result := re.ReplaceAllStringFunc(text, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 2 {
			return ""
		}
		name := strings.TrimSpace(matches[1])
		if name == "" {
			return ""
		}
		normalized := NormalizeQuestion(name)
		if normalized == "" {
			return ""
		}

		q, err := db.GetQuestionByGroupAndNormalized(GlobalDB, groupIDStr, normalized)
		if err != nil || q == nil {
			return ""
		}

		answer, err := db.GetRandomApprovedAnswer(GlobalDB, q.ID)
		if err != nil || answer == nil {
			return ""
		}

		return answer.Answer
	})

	return result
}

func SubstituteAllVariables(text string, event *onebot.Event) string {
	if text == "" || event == nil {
		return text
	}
	t := SubstituteSystemVariables(text, event)
	t = SubstituteCustomVariables(t, event)
	return t
}

func IsAtMe(event *onebot.Event) bool {
	if event == nil {
		return false
	}
	raw := event.RawMessage
	if raw == "" {
		if msg, ok := event.Message.(string); ok {
			raw = msg
		}
	}
	if raw == "" {
		return false
	}
	if event.SelfID == 0 {
		return false
	}
	target := fmt.Sprintf("[CQ:at,qq=%d]", event.SelfID)
	return strings.Contains(raw, target)
}
