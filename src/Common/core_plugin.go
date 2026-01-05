package common

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/rag"
	clog "BotMatrix/common/log"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CorePluginConfig represents the configuration for the system-level core plugin
type CorePluginConfig struct {
	Enabled            bool               `json:"enabled"`
	SystemControl      SystemControl      `json:"system_control"`
	Permissions        Permissions        `json:"permissions"`
	MessageFlowControl MessageFlowControl `json:"message_flow_control"`
	SensitiveWords     SensitiveWords     `json:"sensitive_words"`
	URLFilter          URLFilter          `json:"url_filter"`
	Statistics         Statistics         `json:"statistics"`
	AdminCommands      AdminCommands      `json:"admin_commands"`
	KBCommands         KBCommands         `json:"kb_commands"`
	FlowPriority       FlowPriority       `json:"flow_priority"`
	Scalability        Scalability        `json:"scalability"`
	Monitoring         Monitoring         `json:"monitoring"`
}

type SystemControl struct {
	EnableOpenClose bool `json:"enable_open_close"`
	MaintenanceMode bool `json:"maintenance_mode"`
	DowngradeMode   bool `json:"downgrade_mode"`
}

type Permissions struct {
	Levels    []string  `json:"levels"`
	Whitelist FilterSet `json:"whitelist"`
	Blacklist FilterSet `json:"blacklist"`
}

type FilterSet struct {
	System        []string `json:"system"`
	Robot         []string `json:"robot"`
	Group         []string `json:"group"`
	CloudOfficial []string `json:"cloud_official,omitempty"`
}

type MessageFlowControl struct {
	Types map[string]string `json:"types"` // user_message, system_event, admin_command
}

type SensitiveWords struct {
	Levels     FilterSet `json:"levels"`
	MatchModes []string  `json:"match_modes"` // exact, prefix, regex
}

type URLFilter struct {
	Whitelist  []string `json:"whitelist"`
	Blacklist  []string `json:"blacklist"`
	MatchModes []string `json:"match_modes"` // exact, domain_suffix, regex
}

type Statistics struct {
	Enable                     bool `json:"enable"`
	Async                      bool `json:"async"`
	RecordShortCircuitMessages bool `json:"record_short_circuit_messages"`
}

type AdminCommands struct {
	Support            []string `json:"support"`
	PermissionRequired bool     `json:"permission_required"`
}

type KBCommands struct {
	Enabled              bool `json:"enabled"`
	EnableChatManagement bool `json:"enable_chat_management"`
}

type FlowPriority struct {
	WhitelistHighest    bool `json:"whitelist_highest"`
	BlacklistEnforce    bool `json:"blacklist_enforce"`
	SensitiveWordsCheck bool `json:"sensitive_words_check"`
	URLFilterCheck      bool `json:"url_filter_check"`
}

type Scalability struct {
	MultiInstance bool      `json:"multi_instance"`
	RedisSync     RedisSync `json:"redis_sync"`
}

type RedisSync struct {
	SystemState        bool `json:"system_state"`
	Permissions        bool `json:"permissions"`
	BlacklistWhitelist bool `json:"blacklist_whitelist"`
	StatisticsQueue    bool `json:"statistics_queue"`
}

type Monitoring struct {
	Enable           bool `json:"enable"`
	AlertOnException bool `json:"alert_on_exception"`
	ReportStatistics bool `json:"report_statistics"`
}

// CorePlugin is the system-level core plugin implementation
type CorePlugin struct {
	Manager *Manager
	Config  CorePluginConfig
	Mutex   sync.RWMutex

	// Compiled regex patterns for optimization
	sensitiveRegex []*regexp.Regexp
	urlRegex       []*regexp.Regexp

	// Internal state
	isOpen bool
}

func NewCorePlugin(m *Manager) *CorePlugin {
	p := &CorePlugin{
		Manager: m,
		Config: CorePluginConfig{
			Enabled: true,
			SystemControl: SystemControl{
				EnableOpenClose: true,
				MaintenanceMode: false,
				DowngradeMode:   false,
			},
			Permissions: Permissions{
				Levels: []string{"system_admin", "robot_admin", "group_admin"},
				Whitelist: FilterSet{
					System: []string{},
					Robot:  []string{},
					Group:  []string{},
				},
				Blacklist: FilterSet{
					System:        []string{},
					Robot:         []string{},
					Group:         []string{},
					CloudOfficial: []string{},
				},
			},
			MessageFlowControl: MessageFlowControl{
				Types: map[string]string{
					"user_message":  "can_block",
					"system_event":  "always_forward",
					"admin_command": "always_forward",
					"token_login":   "always_forward",
				},
			},
			SensitiveWords: SensitiveWords{
				MatchModes: []string{"exact", "prefix", "regex"},
			},
			URLFilter: URLFilter{
				MatchModes: []string{"exact", "domain_suffix", "regex"},
			},
			Statistics: Statistics{
				Enable:                     true,
				Async:                      true,
				RecordShortCircuitMessages: true,
			},
			AdminCommands: AdminCommands{
				Support: []string{
					"system_open_close",
					"strategy_update",
					"robot_control",
					"query_status",
					"broadcast_event",
				},
				PermissionRequired: true,
			},
			FlowPriority: FlowPriority{
				WhitelistHighest:    true,
				BlacklistEnforce:    true,
				SensitiveWordsCheck: true,
				URLFilterCheck:      true,
			},
			Scalability: Scalability{
				MultiInstance: true,
				RedisSync: RedisSync{
					SystemState:        true,
					Permissions:        true,
					BlacklistWhitelist: true,
					StatisticsQueue:    true,
				},
			},
			Monitoring: Monitoring{
				Enable:           true,
				AlertOnException: true,
				ReportStatistics: true,
			},
		},
		isOpen: true,
	}

	// Initial sync from Redis if available
	p.SyncFromRedis()

	return p
}

// SyncFromRedis loads state and configs from Redis
func (p *CorePlugin) SyncFromRedis() {
	if p.Manager.Rdb == nil {
		return
	}

	ctx := context.Background()
	// Sync system state
	if p.Config.Scalability.RedisSync.SystemState {
		val, err := p.Manager.Rdb.Get(ctx, "core:system_open").Result()
		if err == nil {
			p.isOpen = val == "true"
		}
	}

	// Sync config (simplified - can be more complex)
	val, err := p.Manager.Rdb.Get(ctx, "core:config").Result()
	if err == nil {
		var newConfig CorePluginConfig
		if err := json.Unmarshal([]byte(val), &newConfig); err == nil {
			p.Mutex.Lock()
			p.Config = newConfig
			p.Mutex.Unlock()
			p.RecompileRegex()
		}
	}
}

// SaveToRedis saves state and configs to Redis
func (p *CorePlugin) SaveToRedis() {
	if p.Manager.Rdb == nil {
		return
	}

	ctx := context.Background()
	p.Mutex.RLock()
	defer p.Mutex.RUnlock()

	// Save system state
	p.Manager.Rdb.Set(ctx, "core:system_open", fmt.Sprintf("%v", p.isOpen), 0)

	// Save config
	data, _ := json.Marshal(p.Config)
	p.Manager.Rdb.Set(ctx, "core:config", string(data), 0)
}

// RecompileRegex compiles sensitive words and URL patterns
func (p *CorePlugin) RecompileRegex() {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.sensitiveRegex = nil
	for _, pattern := range p.Config.SensitiveWords.Levels.System {
		if reg, err := regexp.Compile(pattern); err == nil {
			p.sensitiveRegex = append(p.sensitiveRegex, reg)
		}
	}

	p.urlRegex = nil
	for _, pattern := range p.Config.URLFilter.Blacklist {
		if reg, err := regexp.Compile(pattern); err == nil {
			p.urlRegex = append(p.urlRegex, reg)
		}
	}
}

// ProcessMessage checks if a message should be allowed through the system
// Returns: allowed (bool), reason (string), error
func (p *CorePlugin) ProcessMessage(msg types.InternalMessage) (bool, string, error) {
	p.Mutex.RLock()
	defer p.Mutex.RUnlock()

	if !p.Config.Enabled {
		return true, "", nil
	}

	// 1. Identify message type
	msgType := p.IdentifyMessageType(msg)

	// 2. System state check
	if !p.isOpen && msgType != "admin_command" {
		return false, "system_closed", nil
	}

	// 3. Handle always_forward types
	flowControl := p.Config.MessageFlowControl.Types[msgType]
	if flowControl == "always_forward" {
		return true, "", nil
	}

	// 4. Permission & Blacklist/Whitelist checks
	if allowed, reason := p.checkPermissions(msg); !allowed {
		if p.Config.Statistics.Enable && p.Config.Statistics.RecordShortCircuitMessages {
			go p.RecordBlockedMessage(msg, reason)
		}
		return false, reason, nil
	}

	// 5. Content filtering (Sensitive words & URLs)
	if allowed, reason := p.checkContent(msg); !allowed {
		if p.Config.Statistics.Enable && p.Config.Statistics.RecordShortCircuitMessages {
			go p.RecordBlockedMessage(msg, reason)
		}
		return false, reason, nil
	}

	// 6. Record success statistics
	if p.Config.Statistics.Enable {
		go p.RecordStatistics(msgType, true)
		go p.RecordUserActivity(msg)
	}

	// 7. Handle KB commands if identified
	if msgType == "kb_command" {
		response, err := p.HandleKBCommand(msg)
		if err != nil {
			return false, "kb_command_error", err
		}
		return false, response, nil // Intercepted with response
	}

	// 8. Handle Token Login commands if identified
	if msgType == "token_login" {
		response, err := p.HandleTokenLoginCommand(msg)
		if err != nil {
			return false, "token_login_error", err
		}
		return false, response, nil // Intercepted with response
	}

	// 9. Handle file uploads for RAG indexing
	if p.isFileUpload(msg) {
		response, err := p.HandleFileUpload(msg)
		if err != nil {
			return true, "", nil // Let it pass if handling fails, or block? For now let it pass
		}
		if response != "" {
			return false, response, nil
		}
	}

	return true, "", nil
}

func (p *CorePlugin) isFileUpload(msg types.InternalMessage) bool {
	for _, seg := range msg.Message {
		if seg.Type == "file" || seg.Type == "image" { // Some platforms send docs as images or specific file types
			return true
		}
	}
	return false
}

func (p *CorePlugin) HandleFileUpload(msg types.InternalMessage) (string, error) {
	// RAG check
	tm := p.Manager.GetTaskManager()
	if tm == nil || tm.GetAI() == nil || tm.GetAI().GetManifest() == nil || tm.GetAI().GetManifest().KnowledgeBase == nil {
		return "", nil
	}
	kb, ok := tm.GetAI().GetManifest().KnowledgeBase.(*rag.PostgresKnowledgeBase)
	if !ok {
		return "", nil
	}

	var lastResponse string

	for _, seg := range msg.Message {
		if seg.Type == "file" || seg.Type == "image" {
			data, ok := seg.Data.(map[string]any)
			if !ok {
				continue
			}

			fileURL, _ := data["url"].(string)
			fileName, _ := data["name"].(string)
			if fileName == "" && seg.Type == "image" {
				fileName = fmt.Sprintf("image_%d.jpg", time.Now().Unix())
			}
			if fileURL == "" {
				continue
			}

			// Determine scope and planning
			targetType := "user"
			targetID := msg.UserID
			scopeName := "个人私有"

			if msg.MessageType == "group" {
				targetType = "group"
				targetID = msg.GroupID
				scopeName = fmt.Sprintf("群组 [%s]", msg.GroupName)
			}

			// Asynchronous processing to avoid blocking message flow
			go func(name, url, tType, tID, sName, uploaderID string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()

				clog.Info("[RAG] 正在处理上传文件", zap.String("name", name), zap.String("scope", sName))

				// 1. Download file content
				resp, err := http.Get(url)
				if err != nil {
					clog.Error("[RAG] 下载文件失败", zap.String("url", url), zap.Error(err))
					return
				}
				defer resp.Body.Close()

				content, err := io.ReadAll(resp.Body)
				if err != nil {
					clog.Error("[RAG] 读取文件内容失败", zap.Error(err))
					return
				}

				// 2. Index content using rag.Indexer
				// We use AI service from Manager if available
				var aiSvc ai.AIService
				if p.Manager.AIIntegrationService != nil {
					aiSvc = p.Manager.AIIntegrationService
				}

				indexer := rag.NewIndexer(kb, aiSvc, 0) // 0 表示使用默认模型

				if err := indexer.IndexContent(ctx, name, url, content, "upload", uploaderID, tType, tID); err != nil {
					clog.Error("[RAG] 索引文件内容失败", zap.String("file", name), zap.Error(err))
				} else {
					clog.Info("[RAG] 文件索引成功", zap.String("file", name))
				}
			}(fileName, fileURL, targetType, targetID, scopeName, msg.UserID)

			lastResponse = fmt.Sprintf("📥 已收到文件 [%s]，正在为您存入%s知识库...", fileName, scopeName)
		}
	}

	return lastResponse, nil
}

// RecordStatistics records message statistics to Redis with multiple dimensions
func (p *CorePlugin) RecordStatistics(msgType string, allowed bool) {
	if p.Manager.Rdb == nil {
		return
	}
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("core:stats:%s", today)

	field := msgType
	if !allowed {
		field = "blocked_" + msgType
	}

	// 1. 基础类型统计
	p.Manager.Rdb.HIncrBy(ctx, key, field, 1)
	p.Manager.Rdb.HIncrBy(ctx, key, "total_messages", 1)

	p.Manager.Rdb.Expire(ctx, key, 7*24*time.Hour)
}

// RecordUserActivity records activity for specific users, groups, and bots
func (p *CorePlugin) RecordUserActivity(msg types.InternalMessage) {
	if p.Manager.Rdb == nil {
		return
	}
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	userID := msg.UserID
	groupID := msg.GroupID
	botID := msg.SelfID

	// 使用 Redis Hash 记录今日活跃度
	if userID != "" {
		p.Manager.Rdb.HIncrBy(ctx, fmt.Sprintf("core:stats:users:%s", today), userID, 1)
		p.Manager.Rdb.Expire(ctx, fmt.Sprintf("core:stats:users:%s", today), 2*24*time.Hour)
	}
	if groupID != "" {
		p.Manager.Rdb.HIncrBy(ctx, fmt.Sprintf("core:stats:groups:%s", today), groupID, 1)
		p.Manager.Rdb.Expire(ctx, fmt.Sprintf("core:stats:groups:%s", today), 2*24*time.Hour)
	}
	if botID != "" {
		p.Manager.Rdb.HIncrBy(ctx, fmt.Sprintf("core:stats:bots:%s", today), botID, 1)
		p.Manager.Rdb.Expire(ctx, fmt.Sprintf("core:stats:bots:%s", today), 2*24*time.Hour)
	}
}

// RecordBlockedMessage records details of a blocked message
func (p *CorePlugin) RecordBlockedMessage(msg types.InternalMessage, reason string) {
	if p.Manager.Rdb == nil {
		return
	}
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("core:blocked:%s", today)

	message := msg.RawMessage
	if message == "" {
		var sb strings.Builder
		for _, seg := range msg.Message {
			if seg.Type == "text" {
				if data, ok := seg.Data.(map[string]any); ok {
					if t, ok := data["text"].(string); ok {
						sb.WriteString(t)
					}
				} else if t, ok := seg.Data.(string); ok {
					sb.WriteString(t)
				}
			}
		}
		message = sb.String()
	}

	record := map[string]any{
		"time":     time.Now().Format(time.RFC3339),
		"bot_id":   msg.SelfID,
		"user_id":  msg.UserID,
		"group_id": msg.GroupID,
		"content":  message,
		"reason":   reason,
		"protocol": msg.Protocol,
	}

	data, _ := json.Marshal(record)
	p.Manager.Rdb.LPush(ctx, key, data)
	p.Manager.Rdb.LTrim(ctx, key, 0, 99) // Keep last 100
	p.Manager.Rdb.Expire(ctx, key, 7*24*time.Hour)
}

// IdentifyMessageType 识别消息类型并根据优先级分发
func (p *CorePlugin) IdentifyMessageType(msg types.InternalMessage) string {
	// For InternalMessage, we can rely on Protocol and MessageType
	if msg.MessageType != "" {
		if p.isInternalAdminCommand(msg) {
			return "admin_command"
		}
		if p.isKBCommand(msg) {
			return "kb_command"
		}
		if p.isTokenLoginCommand(msg) {
			return "token_login"
		}
		return "user_message"
	}
	return "system_event"
}

func (p *CorePlugin) isKBCommand(msg types.InternalMessage) bool {
	message := p.extractTextMessage(msg)
	return strings.HasPrefix(message, "/kb")
}

func (p *CorePlugin) extractTextMessage(msg types.InternalMessage) string {
	message := msg.RawMessage
	if message == "" {
		for _, seg := range msg.Message {
			if seg.Type == "text" {
				if data, ok := seg.Data.(map[string]any); ok {
					if t, ok := data["text"].(string); ok {
						message = t
						break
					}
				} else if t, ok := seg.Data.(string); ok {
					message = t
					break
				}
			}
		}
	}
	return message
}

func (p *CorePlugin) isInternalAdminCommand(msg types.InternalMessage) bool {
	message := p.extractTextMessage(msg)
	return strings.HasPrefix(message, "/system") || strings.HasPrefix(message, "/nexus")
}

func (p *CorePlugin) checkPermissions(msg types.InternalMessage) (bool, string) {
	userID := msg.UserID
	groupID := msg.GroupID
	botID := msg.SelfID

	// 1. Blacklist checks (Highest priority)
	if p.isInList(userID, p.Config.Permissions.Blacklist.System) {
		return false, "user_blacklisted"
	}
	if p.isInList(groupID, p.Config.Permissions.Blacklist.Group) {
		return false, "group_blacklisted"
	}
	if p.isInList(botID, p.Config.Permissions.Blacklist.Robot) {
		return false, "robot_blacklisted"
	}

	// 2. Whitelist checks
	if p.Config.FlowPriority.WhitelistHighest {
		if p.isInList(userID, p.Config.Permissions.Whitelist.System) ||
			p.isInList(groupID, p.Config.Permissions.Whitelist.Group) ||
			p.isInList(botID, p.Config.Permissions.Whitelist.Robot) {
			return true, ""
		}
	}

	return true, ""
}

func (p *CorePlugin) isTokenLoginCommand(msg types.InternalMessage) bool {
	message := strings.TrimSpace(p.extractTextMessage(msg))
	// 支持 "后台"、"登录"、"控制台" 等关键词触发
	return message == "后台" || message == "登录" || message == "控制台" || message == "admin"
}

func (p *CorePlugin) HandleTokenLoginCommand(msg types.InternalMessage) (string, error) {
	if p.Manager.GORMDB == nil {
		return "❌ 系统配置错误：数据库未连接", nil
	}

	// 生成随机 8 位 Token
	token := utils.GenerateRandomToken(4) // 4 bytes = 8 hex chars

	// 存入数据库，有效期 10 分钟
	tokenRecord := models.UserLoginTokenGORM{
		Platform:   msg.Platform,
		PlatformID: msg.UserID,
		Token:      token,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		CreatedAt:  time.Now(),
	}

	err := p.Manager.GORMDB.Create(&tokenRecord).Error
	if err != nil {
		return "❌ 登录令牌生成失败，请稍后再试", err
	}

	// 获取前端地址
	// TODO: 生产环境应从配置获取公网域名
	webUIAddr := "http://localhost:5173"

	loginURL := fmt.Sprintf("%s/auth/token-login?platform=%s&platform_id=%s&token=%s",
		webUIAddr, msg.Platform, msg.UserID, token)

	return fmt.Sprintf("🔑 临时登录令牌已生成（10分钟有效）：\n\n%s\n\n请点击链接直接登录控制台。请勿将此链接泄露给他人。", loginURL), nil
}

func (p *CorePlugin) isInList(target string, list []string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func (p *CorePlugin) checkContent(msg types.InternalMessage) (bool, string) {
	message := msg.RawMessage
	if message == "" {
		// If raw message is empty (v12), reconstruct for content check
		var sb strings.Builder
		for _, seg := range msg.Message {
			if seg.Type == "text" {
				if data, ok := seg.Data.(map[string]any); ok {
					if t, ok := data["text"].(string); ok {
						sb.WriteString(t)
					}
				} else if t, ok := seg.Data.(string); ok {
					sb.WriteString(t)
				}
			}
		}
		message = sb.String()
	}

	if message == "" {
		return true, ""
	}

	// 1. Sensitive words check
	if p.Config.FlowPriority.SensitiveWordsCheck {
		for _, word := range p.Config.SensitiveWords.Levels.System {
			if strings.Contains(message, word) {
				return false, "sensitive_word_detected"
			}
		}
		// Regex check
		for _, reg := range p.sensitiveRegex {
			if reg.MatchString(message) {
				return false, "sensitive_word_regex_detected"
			}
		}
	}

	// 2. URL filter check
	if p.Config.FlowPriority.URLFilterCheck {
		// Simple URL detection (can be improved)
		if strings.Contains(message, "http://") || strings.Contains(message, "https://") {
			for _, blacklistedURL := range p.Config.URLFilter.Blacklist {
				if strings.Contains(message, blacklistedURL) {
					return false, "blacklisted_url_detected"
				}
			}
			// Regex check
			for _, reg := range p.urlRegex {
				if reg.MatchString(message) {
					return false, "blacklisted_url_regex_detected"
				}
			}
		}
	}

	return true, ""
}

// HandleAdminCommand processes system admin commands
func (p *CorePlugin) HandleAdminCommand(msg types.InternalMessage) (string, error) {
	message := msg.RawMessage
	if message == "" {
		for _, seg := range msg.Message {
			if seg.Type == "text" {
				if data, ok := seg.Data.(map[string]any); ok {
					if t, ok := data["text"].(string); ok {
						message = t
						break
					}
				} else if t, ok := seg.Data.(string); ok {
					message = t
					break
				}
			}
		}
	}

	parts := strings.Fields(message)
	if len(parts) < 2 {
		return "Usage: /system <command> [args]", nil
	}

	cmd := parts[1]
	args := parts[2:]

	switch cmd {
	case "open":
		p.isOpen = true
		p.SaveToRedis()
		return "✅ 系统已开启 (Nexus Core System Opened)", nil
	case "close":
		p.isOpen = false
		p.SaveToRedis()
		return "🔒 系统已关闭 (Nexus Core System Closed)", nil
	case "status":
		status := "🟢 运行中 (Open)"
		if !p.isOpen {
			status = "🔴 已停止 (Closed)"
		}
		stats := ""
		if p.Manager.Rdb != nil {
			today := time.Now().Format("2006-01-02")
			res, _ := p.Manager.Rdb.HGetAll(context.Background(), fmt.Sprintf("core:stats:%s", today)).Result()
			for k, v := range res {
				stats += fmt.Sprintf("\n- %s: %s", k, v)
			}
		}
		return fmt.Sprintf("📊 Nexus 核心状态:\n状态: %s\n在线机器人: %d\n活跃工作节点: %d\n今日流水: %s",
			status, len(p.Manager.Bots), len(p.Manager.Workers), stats), nil
	case "top":
		if p.Manager.Rdb == nil {
			return "错误: Redis 未连接", nil
		}
		today := time.Now().Format("2006-01-02")
		ctx := context.Background()

		// 获取活跃用户前 5
		users, _ := p.Manager.Rdb.HGetAll(ctx, fmt.Sprintf("core:stats:users:%s", today)).Result()
		// 获取活跃群组前 5
		groups, _ := p.Manager.Rdb.HGetAll(ctx, fmt.Sprintf("core:stats:groups:%s", today)).Result()

		res := "🏆 今日发言排行榜:"
		res += "\n\n👤 活跃用户:"
		if len(users) == 0 {
			res += "\n暂无数据"
		} else {
			// 简单排序输出
			for id, count := range users {
				res += fmt.Sprintf("\n- %s: %s 次", id, count)
			}
		}

		res += "\n\n👥 活跃群组:"
		if len(groups) == 0 {
			res += "\n暂无数据"
		} else {
			for id, count := range groups {
				res += fmt.Sprintf("\n- %s: %s 次", id, count)
			}
		}
		return res, nil
	case "whitelist":
		if len(args) < 3 {
			return "用法: /system whitelist <add|remove> <system|robot|group> <id>", nil
		}
		action := args[0]
		target := args[1]
		id := args[2]

		p.Mutex.Lock()
		defer p.Mutex.Unlock()

		list := &p.Config.Permissions.Whitelist.System
		switch target {
		case "robot":
			list = &p.Config.Permissions.Whitelist.Robot
		case "group":
			list = &p.Config.Permissions.Whitelist.Group
		}

		if action == "add" {
			*list = append(*list, id)
			p.SaveToRedis()
			return fmt.Sprintf("✅ 已添加 %s 到 %s 白名单", id, target), nil
		} else {
			// Remove logic...
			return "暂不支持移除操作", nil
		}
	case "blacklist":
		if len(args) < 3 {
			return "用法: /system blacklist <add|remove> <system|robot|group> <id>", nil
		}
		action := args[0]
		target := args[1]
		id := args[2]

		p.Mutex.Lock()
		defer p.Mutex.Unlock()

		list := &p.Config.Permissions.Blacklist.System
		switch target {
		case "robot":
			list = &p.Config.Permissions.Blacklist.Robot
		case "group":
			list = &p.Config.Permissions.Blacklist.Group
		}

		if action == "add" {
			*list = append(*list, id)
			p.SaveToRedis()
			return fmt.Sprintf("🚫 已添加 %s 到 %s 黑名单", id, target), nil
		} else {
			return "暂不支持移除操作", nil
		}
	case "reload":
		p.SyncFromRedis()
		return "🔄 配置已从 Redis 重新加载", nil
	default:
		return "❓ 未知指令。可用指令: open, close, status, whitelist, blacklist, reload", nil
	}
}

// HandleKBCommand processes knowledge base management commands
func (p *CorePlugin) HandleKBCommand(msg types.InternalMessage) (string, error) {
	if !p.Config.KBCommands.Enabled {
		return "❌ 知识库管理功能未开启", nil
	}

	message := p.extractTextMessage(msg)
	parts := strings.Fields(message)
	if len(parts) < 2 {
		return "📚 知识库管理指令帮助:\n- /kb list : 查看我的文档\n- /kb del <ID> : 删除指定文档\n- /kb status : 知识库运行状态", nil
	}

	cmd := parts[1]
	args := parts[2:]

	// 权限检查
	isAdmin := p.isInList(msg.UserID, p.Config.Permissions.Whitelist.System)

	// 获取 RAG 组件
	tm := p.Manager.GetTaskManager()
	if tm == nil || tm.GetAI() == nil || tm.GetAI().GetManifest() == nil || tm.GetAI().GetManifest().KnowledgeBase == nil {
		return "❌ RAG 系统未初始化", nil
	}
	kb, ok := tm.GetAI().GetManifest().KnowledgeBase.(*rag.PostgresKnowledgeBase)
	if !ok {
		return "❌ 知识库引擎不支持当前操作", nil
	}

	switch cmd {
	case "list":
		// 根据场景判断展示范围
		var docs []rag.KnowledgeDoc
		var err error
		var scope string

		if msg.MessageType == "group" {
			docs, err = kb.GetUserDocs(context.Background(), "group", msg.GroupID)
			scope = fmt.Sprintf("群组 [%s]", msg.GroupName)
		} else {
			docs, err = kb.GetUserDocs(context.Background(), "user", msg.UserID)
			scope = "个人"
		}

		if err != nil {
			return fmt.Sprintf("❌ 获取文档列表失败: %v", err), nil
		}

		if len(docs) == 0 {
			return fmt.Sprintf("📭 %s知识库暂无文档", scope), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📂 %s知识库文档列表:\n", scope))
		for _, doc := range docs {
			status := "🟢"
			if doc.Status != "active" {
				status = "🟡"
			}
			sb.WriteString(fmt.Sprintf("%s ID:%d | %s\n", status, doc.ID, doc.Title))
		}
		return sb.String(), nil

	case "del":
		if len(args) < 1 {
			return "用法: /kb del <ID>", nil
		}
		docID, _ := strconv.ParseUint(args[0], 10, 32)
		if docID == 0 {
			return "❌ 无效的文档 ID", nil
		}

		// 检查所有权
		if !isAdmin && !kb.IsDocOwner(context.Background(), uint(docID), msg.UserID) {
			return "❌ 您没有权限删除此文档 (仅所有者或管理员可操作)", nil
		}

		if err := kb.DeleteDoc(context.Background(), uint(docID)); err != nil {
			return fmt.Sprintf("❌ 删除失败: %v", err), nil
		}
		return fmt.Sprintf("✅ 已成功删除文档 ID:%d", docID), nil

	case "status":
		stats := "🤖 RAG 知识库状态:\n"
		stats += fmt.Sprintf("- 存储引擎: PostgreSQL + pgvector\n")
		stats += fmt.Sprintf("- 当前机器人: %s\n", msg.SelfID)

		// 统计总数 (需要管理员权限查看更多)
		if isAdmin {
			// 这里可以添加更详细的统计
		}
		return stats, nil

	case "add":
		// /kb add 命令通常配合文件上传。
		// 在这里我们给出引导提示。
		return "📝 添加文档说明:\n请直接在聊天中发送文件 (PDF/Docx/TXT/MD/Code)，Nexus 会根据您的身份自动归类并索引。", nil

	case "sync":
		if !isAdmin {
			return "❌ 只有管理员可以手动触发系统文档同步", nil
		}
		go p.Manager.SyncSystemKnowledge()
		return "🔄 已在后台启动系统文档同步任务...", nil

	default:
		return "❓ 未知知识库指令。可用指令: list, del, status, add, sync", nil
	}
}
