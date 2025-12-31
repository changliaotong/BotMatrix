package app

import (
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
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
	msgType := p.identifyMessageType(msg)

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

	return true, "", nil
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

func (p *CorePlugin) identifyMessageType(msg types.InternalMessage) string {
	// For InternalMessage, we can rely on Protocol and MessageType
	if msg.MessageType != "" {
		if p.isInternalAdminCommand(msg) {
			return "admin_command"
		}
		return "user_message"
	}
	return "system_event"
}

func (p *CorePlugin) isInternalAdminCommand(msg types.InternalMessage) bool {
	// Implementation for admin command detection
	// Usually starts with a specific prefix like /system or /admin
	message := msg.RawMessage
	if message == "" {
		// If raw message is empty (v12), check text segments
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
