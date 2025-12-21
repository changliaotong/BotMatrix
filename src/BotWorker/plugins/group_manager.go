package plugins

import (
	"botworker/internal/db"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/redis"
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GroupManagerPlugin struct {
	// 数据库连接
	db *sql.DB
	// Redis客户端
	redisClient *redis.Client
}

func NewGroupManagerPlugin(database *sql.DB, redisClient *redis.Client) *GroupManagerPlugin {
	return &GroupManagerPlugin{
		db:          database,
		redisClient: redisClient,
	}
}

func (p *GroupManagerPlugin) Name() string {
	return "group_manager"
}

func (p *GroupManagerPlugin) Description() string {
	return "群管插件，提供踢人、禁言、关键词过滤等功能"
}

func (p *GroupManagerPlugin) Version() string {
	return "1.0.0"
}

func (p *GroupManagerPlugin) Init(robot plugin.Robot) {
	log.Println("加载群管插件")

	// 如果数据库连接可用，添加默认敏感词
	if p.db != nil {
		// 添加默认敏感词（如果不存在）
		defaultSensitiveWords := []string{"敏感词1", "敏感词2", "敏感词3"}
		for _, word := range defaultSensitiveWords {
			if err := db.AddSensitiveWord(p.db, word); err != nil {
				log.Printf("添加默认敏感词失败: %v", err)
			}
		}

		// 设置默认群规（如果不存在）
		defaultRules := `1. 遵守国家法律法规
2. 禁止发布违法信息
3. 禁止发布广告
4. 禁止人身攻击
5. 禁止刷屏
6. 保持文明交流`
		if err := db.SetGroupRules(p.db, "0", defaultRules); err != nil {
			log.Printf("设置默认群规失败: %v", err)
		}
	}

	// 处理群消息事件
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" {
			return nil
		}

		// 检查是否为管理员命令
		if p.isAdminCommand(event) {
			return p.handleAdminCommand(robot, event)
		}

		// 关键词过滤
		if p.containsSensitiveWords(event.RawMessage) {
			// 警告用户
			warningMsg := fmt.Sprintf("@%d 请注意你的发言，包含敏感词汇！", event.UserID)
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: warningMsg,
			})

			// 记录日志
			log.Printf("用户 %d 在群 %d 发送了敏感消息: %s", event.UserID, event.GroupID, event.RawMessage)
		}

		// 检查是否是命令
		switch strings.ToLower(strings.TrimSpace(event.RawMessage)) {
		case "群规":
			p.sendGroupRules(robot, event)
		case "help":
			p.sendHelp(robot, event)
		}

		return nil
	})

	// 处理群成员增加事件
	robot.OnNotice(func(event *onebot.Event) error {
		if event.NoticeType == "group_member_increase" {
			// 发送欢迎消息和群规
			p.sendWelcomeAndRules(robot, event)
		}
		return nil
	})

	// 定期检查禁言时间
	go p.checkBanExpiration(robot)
}

// 检查是否是管理员命令
func (p *GroupManagerPlugin) isAdminCommand(event *onebot.Event) bool {
	if event.MessageType != "group" {
		return false
	}

	// 检查消息是否以 "!" 或 "/" 开头
	return strings.HasPrefix(event.RawMessage, "!") || strings.HasPrefix(event.RawMessage, "/")
}

// 处理管理员命令
func (p *GroupManagerPlugin) handleAdminCommand(robot plugin.Robot, event *onebot.Event) error {
	// 检查是否为管理员
	if !p.isAdmin(event.GroupID, event.UserID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "权限不足，只有管理员可以执行此命令！",
		})
		return nil
	}

	// 解析命令
	parts := strings.Fields(event.RawMessage)
	if len(parts) < 1 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// 处理不同的命令
	switch command {
	case "!kick", "/kick":
		p.handleKickCommand(robot, event, args)
	case "!ban", "/ban":
		p.handleBanCommand(robot, event, args)
	case "!unban", "/unban":
		p.handleUnbanCommand(robot, event, args)
	case "!addadmin", "/addadmin":
		p.handleAddAdminCommand(robot, event, args)
	case "!deladmin", "/deladmin":
		p.handleDelAdminCommand(robot, event, args)
	case "!setrules", "/setrules":
		p.handleSetRulesCommand(robot, event, args)
	case "!addword", "/addword":
		p.handleAddWordCommand(robot, event, args)
	case "!delword", "/delword":
		p.handleDelWordCommand(robot, event, args)
	case "!members", "/members":
		p.handleGetMembersCommand(robot, event, args)
	case "!memberinfo", "/memberinfo":
		p.handleGetMemberInfoCommand(robot, event, args)
	}

	return nil
}

// 处理踢人命令
func (p *GroupManagerPlugin) handleKickCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !kick <用户ID> [是否拒绝入群]",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 检查是否是管理员
	if p.isAdmin(event.GroupID, userID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "不能踢管理员！",
		})
		log.Printf("[GroupManager] 尝试踢群 %d 中的管理员 %d，操作被拒绝", event.GroupID, userID)
		return
	}

	// 执行踢人操作
	refuse := false
	if len(args) > 1 && (args[1] == "true" || args[1] == "1") {
		refuse = true
	}

	// 记录踢人操作
	log.Printf("[GroupManager] 尝试踢群 %d 中的用户 %d，拒绝再次加入: %v", event.GroupID, userID, refuse)

	_, err = robot.SetGroupKick(&onebot.SetGroupKickParams{
		GroupID:   event.GroupID,
		UserID:    userID,
		RejectAdd: refuse,
	})

	if err != nil {
		log.Printf("[GroupManager] 踢群 %d 中的用户 %d 失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: fmt.Sprintf("踢人失败: %v", err),
		})
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 已成功将用户 %d 踢出群 %d", userID, event.GroupID)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已将用户 %d 踢出群聊", userID),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "kick",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf("将用户 %d 踢出群聊，拒绝再次加入: %v", userID, refuse),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理禁言命令
func (p *GroupManagerPlugin) handleBanCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !ban <用户ID> [时长(分钟)]",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 检查是否是管理员
	if p.isAdmin(event.GroupID, userID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "不能禁言管理员！",
		})
		log.Printf("[GroupManager] 尝试禁言群 %d 中的管理员 %d，操作被拒绝", event.GroupID, userID)
		return
	}

	// 解析禁言时长
	duration := 30 * time.Minute // 默认30分钟
	if len(args) > 1 {
		minutes, err := parseDuration(args[1])
		if err == nil && minutes > 0 {
			duration = time.Duration(minutes) * time.Minute
		} else {
			log.Printf("[GroupManager] 解析禁言时长 '%s' 失败，使用默认时长30分钟", args[1])
		}
	}

	// 执行禁言操作
	log.Printf("[GroupManager] 尝试禁言群 %d 中的用户 %d，时长 %d 分钟", event.GroupID, userID, int(duration.Minutes()))

	_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
		GroupID:  event.GroupID,
		UserID:   userID,
		Duration: int64(duration.Seconds()),
	})

	if err != nil {
		log.Printf("[GroupManager] 禁言群 %d 中的用户 %d 失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: fmt.Sprintf("禁言失败: %v", err),
		})
		return
	}

	// 存储禁言信息到Redis
	if p.redisClient != nil {
		ctx := context.Background()
		groupIDStr := fmt.Sprintf("%d", event.GroupID)
		userIDStr := fmt.Sprintf("%d", userID)
		banKey := fmt.Sprintf("group:%s:ban:%s", groupIDStr, userIDStr)

		// 设置禁言记录，带过期时间
		if err := p.redisClient.Set(ctx, banKey, time.Now().Add(duration).Unix(), duration).Err(); err != nil {
			log.Printf("[GroupManager] 存储禁言信息到Redis失败: %v", err)
			// 回退到数据库存储
			if p.db != nil {
				banEndTime := time.Now().Add(duration)
				if err := db.BanUser(p.db, groupIDStr, userIDStr, banEndTime); err != nil {
					log.Printf("[GroupManager] 存储禁言信息到数据库也失败: %v", err)
				} else {
					log.Printf("[GroupManager] 已将禁言信息回退到数据库存储")
				}
			}
		} else {
			log.Printf("[GroupManager] 已将禁言信息存储到Redis")
		}
	} else if p.db != nil {
		// Redis不可用时，使用数据库存储
		banEndTime := time.Now().Add(duration)
		if err := db.BanUser(p.db, fmt.Sprintf("%d", event.GroupID), fmt.Sprintf("%d", userID), banEndTime); err != nil {
			log.Printf("[GroupManager] 存储禁言信息到数据库失败: %v", err)
		} else {
			log.Printf("[GroupManager] 已将禁言信息存储到数据库")
		}
	} else {
		log.Printf("[GroupManager] Redis和数据库都不可用，无法持久化禁言信息")
	}

	// 记录成功操作
	log.Printf("[GroupManager] 已成功将用户 %d 禁言 %d 分钟", userID, int(duration.Minutes()))
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已将用户 %d 禁言 %d 分钟", userID, int(duration.Minutes())),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "ban",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf("将用户 %d 禁言 %d 分钟", userID, int(duration.Minutes())),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理解除禁言命令
func (p *GroupManagerPlugin) handleUnbanCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !unban <用户ID>",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 执行解除禁言操作
	log.Printf("[GroupManager] 尝试解除群 %d 中用户 %d 的禁言", event.GroupID, userID)

	_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
		GroupID:  event.GroupID,
		UserID:   userID,
		Duration: 0,
	})

	if err != nil {
		log.Printf("[GroupManager] 解除群 %d 中用户 %d 的禁言失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: fmt.Sprintf("解除禁言失败: %v", err),
		})
		return
	}

	// 从Redis移除禁言记录
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	if p.redisClient != nil {
		ctx := context.Background()
		banKey := fmt.Sprintf("group:%s:ban:%s", groupIDStr, userIDStr)

		if err := p.redisClient.Del(ctx, banKey).Err(); err != nil {
			log.Printf("[GroupManager] 从Redis移除禁言记录失败: %v", err)
		} else {
			log.Printf("[GroupManager] 已从Redis移除禁言记录")
		}
	}

	// 同时从数据库移除禁言记录，确保数据一致性
	if p.db != nil {
		if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
			log.Printf("[GroupManager] 从数据库移除禁言记录失败: %v", err)
		} else {
			log.Printf("[GroupManager] 已从数据库移除禁言记录")
		}
	}

	// 记录成功操作
	log.Printf("[GroupManager] 已成功解除用户 %d 的禁言", userID)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已解除用户 %d 的禁言", userID),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "unban",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf("解除用户 %d 的禁言", userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理添加管理员命令
func (p *GroupManagerPlugin) handleAddAdminCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	// 只有超级管理员可以添加管理员
	if !p.isSuperAdmin(event.GroupID, event.UserID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "权限不足，只有超级管理员可以添加管理员！",
		})
		log.Printf("[GroupManager] 用户 %d 尝试在群 %d 中添加管理员，但不是超级管理员，操作被拒绝", event.UserID, event.GroupID)
		return
	}

	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !addadmin <用户ID>",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "数据库连接不可用，操作失败！",
		})
		log.Printf("[GroupManager] 数据库连接不可用，无法添加管理员")
		return
	}

	// 添加到管理员列表（数据库）
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	// 检查是否已经是管理员
	isAdmin, err := db.IsGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] 检查群 %d 中用户 %d 的管理员状态失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "操作失败，请稍后重试！",
		})
		return
	}

	if isAdmin {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "该用户已经是管理员！",
		})
		return
	}

	// 添加管理员，默认权限级别为1（普通管理员）	if err := db.AddGroupAdmin(p.db, groupIDStr, userIDStr, 1); err != nil {
		log.Printf("[GroupManager] 向群 %d 添加管理员 %d 失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "操作失败，请稍后重试！",
		})
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 群 %d 中用户 %d 已成功添加为管理员", event.GroupID, userID)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已将用户 %d 添加为管理员", userID),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "add_admin",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf("将用户 %d 添加为群管理员", userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理删除管理员命令
func (p *GroupManagerPlugin) handleDelAdminCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	// 只有超级管理员可以删除管理员
	if !p.isSuperAdmin(event.GroupID, event.UserID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "权限不足，只有超级管理员可以删除管理员！",
		})
		log.Printf("[GroupManager] 用户 %d 尝试在群 %d 中删除管理员，但不是超级管理员，操作被拒绝", event.UserID, event.GroupID)
		return
	}

	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !deladmin <用户ID>",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "数据库连接不可用，操作失败！",
		})
		log.Printf("[GroupManager] 数据库连接不可用，无法删除管理员")
		return
	}

	// 从管理员列表中删除（数据库）
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	userIDStr := fmt.Sprintf("%d", userID)

	// 移除管理员
	err = db.RemoveGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] 从群 %d 中删除管理员 %d 失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "该用户不是管理员！",
		})
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 群 %d 中用户 %d 已成功移除管理员身份", event.GroupID, userID)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已将用户 %d 移除管理员身份", userID),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "del_admin",
			TargetUserID: fmt.Sprintf("%d", userID),
			Description:  fmt.Sprintf("将用户 %d 移除管理员身份", userID),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理设置群规命令
func (p *GroupManagerPlugin) handleSetRulesCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !setrules <群规内容>",
		})
		return
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "数据库连接不可用，操作失败！",
		})
		log.Printf("[GroupManager] 数据库连接不可用，无法设置群规")
		return
	}

	// 设置群规
	rules := strings.Join(args, " ")
	groupIDStr := fmt.Sprintf("%d", event.GroupID)

	if err := db.SetGroupRules(p.db, groupIDStr, rules); err != nil {
		log.Printf("[GroupManager] 设置群 %d 的群规失败: %v", event.GroupID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "设置群规失败，请稍后重试！",
		})
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 群 %d 的群规已成功更新", event.GroupID)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: "群规已更新！",
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:      fmt.Sprintf("%d", event.GroupID),
			AdminID:      fmt.Sprintf("%d", event.UserID),
			Action:       "set_rules",
			Description:  fmt.Sprintf("更新群规为: %s", rules),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理添加敏感词命令
func (p *GroupManagerPlugin) handleAddWordCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !addword <敏感词>",
		})
		return
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "数据库连接不可用，操作失败！",
		})
		log.Printf("[GroupManager] 数据库连接不可用，无法添加敏感词")
		return
	}

	// 添加敏感词
	word := strings.Join(args, " ")

	// 添加到数据库
	if err := db.AddSensitiveWord(p.db, word); err != nil {
		log.Printf("[GroupManager] 添加敏感词 '%s' 失败: %v", word, err)
		// 检查是否为重复添加
		if strings.Contains(err.Error(), "duplicate key") {
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: "该敏感词已经存在！",
			})
		} else {
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: "添加敏感词失败，请稍后重试！",
			})
		}
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 敏感词 '%s' 已成功添加", word)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已添加敏感词: %s", word),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:     fmt.Sprintf("%d", event.GroupID),
			AdminID:     fmt.Sprintf("%d", event.UserID),
			Action:      "add_sensitive_word",
			Description: fmt.Sprintf("添加敏感词: %s", word),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 处理删除敏感词命令
func (p *GroupManagerPlugin) handleDelWordCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !delword <敏感词>",
		})
		return
	}

	// 检查数据库连接是否可用
	if p.db == nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "数据库连接不可用，操作失败！",
		})
		log.Printf("[GroupManager] 数据库连接不可用，无法删除敏感词")
		return
	}

	// 删除敏感词
	word := strings.Join(args, " ")

	// 从数据库删除
	if err := db.RemoveSensitiveWord(p.db, word); err != nil {
		log.Printf("[GroupManager] 删除敏感词 '%s' 失败: %v", word, err)
		// 检查是否为不存在的敏感词
		if strings.Contains(err.Error(), "no rows in result set") || strings.Contains(err.Error(), "not found") {
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: "该敏感词不存在！",
			})
		} else {
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: "删除敏感词失败，请稍后重试！",
			})
		}
		return
	}

	// 记录成功操作
	log.Printf("[GroupManager] 敏感词 '%s' 已成功删除", word)
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fmt.Sprintf("已删除敏感词: %s", word),
	})

	// 记录审核日志
	if p.db != nil {
		auditLog := &db.AuditLog{
			GroupID:     fmt.Sprintf("%d", event.GroupID),
			AdminID:     fmt.Sprintf("%d", event.UserID),
			Action:      "del_sensitive_word",
			Description: fmt.Sprintf("删除敏感词: %s", word),
		}
		if err := db.AddAuditLog(p.db, auditLog); err != nil {
			log.Printf("[GroupManager] 添加审核日志失败: %v", err)
		}
	}
}

// 检查是否为管理员
func (p *GroupManagerPlugin) isAdmin(groupID, userID int64) bool {
	// 从数据库检查是否为群管理员
	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	isAdmin, err := db.IsGroupAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] 检查群 %d 中用户 %d 的管理员状态失败: %v", groupID, userID, err)
		return false
	}

	return isAdmin
	// 可以添加更多管理员检查逻辑，如群主检查
}

// 检查是否为超级管理员
func (p *GroupManagerPlugin) isSuperAdmin(groupID, userID int64) bool {
	// 从数据库检查是否为超级管理员
	groupIDStr := fmt.Sprintf("%d", groupID)
	userIDStr := fmt.Sprintf("%d", userID)

	isSuperAdmin, err := db.IsSuperAdmin(p.db, groupIDStr, userIDStr)
	if err != nil {
		log.Printf("[GroupManager] 检查群 %d 中用户 %d 的超级管理员状态失败: %v", groupID, userID, err)
		return false
	}

	return isSuperAdmin
}

// 检查消息是否包含敏感词
func (p *GroupManagerPlugin) containsSensitiveWords(message string) bool {
	// 检查数据库连接是否可用
	if p.db == nil {
		log.Printf("[GroupManager] 数据库连接不可用，无法检查敏感词")
		return false
	}

	// 从数据库获取所有敏感词
	words, err := db.GetAllSensitiveWords(p.db)
	if err != nil {
		log.Printf("[GroupManager] 从数据库获取敏感词失败: %v", err)
		return false
	}

	for _, word := range words {
		if strings.Contains(message, word) {
			log.Printf("[GroupManager] 检测到敏感词 '%s'", word)
			return true
		}
	}
	return false
}

// 发送欢迎消息和群规
func (p *GroupManagerPlugin) sendWelcomeAndRules(robot plugin.Robot, event *onebot.Event) {
	// 发送欢迎消息
	welcomeMsg := fmt.Sprintf("欢迎新成员 @%d 加入本群！\n\n请遵守群规：", event.UserID)

	// 从数据库获取群规
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	rules, err := db.GetGroupRules(p.db, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] 获取群 %d 的群规失败: %v", event.GroupID, err)
		// 使用默认群规
		if err == sql.ErrNoRows {
			defaultRules, err := db.GetGroupRules(p.db, "0")
			if err != nil {
				log.Printf("[GroupManager] 获取默认群规失败: %v", err)
				rules = ""
			} else {
				rules = defaultRules
			}
		}
	}

	if rules == "" {
		// 如果数据库中没有群规，使用默认群规
		rules = `1. 遵守国家法律法规
2. 禁止发布违法信息
3. 禁止发布广告
4. 禁止人身攻击
5. 禁止刷屏
6. 保持文明交流`
		log.Printf("[GroupManager] 使用内置默认群规")
	}

	// 合并消息
	fullMsg := welcomeMsg + "\n" + rules

	// 发送消息
	_, err = robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: fullMsg,
	})
	if err != nil {
		log.Printf("[GroupManager] 向群 %d 发送欢迎消息失败: %v", event.GroupID, err)
	}
}

// 发送群规
func (p *GroupManagerPlugin) sendGroupRules(robot plugin.Robot, event *onebot.Event) {
	// 从数据库获取群规
	groupIDStr := fmt.Sprintf("%d", event.GroupID)
	rules, err := db.GetGroupRules(p.db, groupIDStr)
	if err != nil {
		log.Printf("[GroupManager] 获取群 %d 的群规失败: %v", event.GroupID, err)
		// 使用默认群规
		if err == sql.ErrNoRows {
			defaultRules, err := db.GetGroupRules(p.db, "0")
			if err != nil {
				log.Printf("[GroupManager] 获取默认群规失败: %v", err)
				rules = ""
			} else {
				rules = defaultRules
			}
		}
	}

	if rules == "" {
		// 如果数据库中没有群规，使用默认群规
		rules = `1. 遵守国家法律法规
2. 禁止发布违法信息
3. 禁止发布广告
4. 禁止人身攻击
5. 禁止刷屏
6. 保持文明交流`
		log.Printf("[GroupManager] 使用内置默认群规")
	}

	// 发送群规
	_, err = robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: "群规：\n" + rules,
	})
	if err != nil {
		log.Printf("[GroupManager] 向群 %d 发送群规失败: %v", event.GroupID, err)
	}
}

// 发送帮助信息
func (p *GroupManagerPlugin) sendHelp(robot plugin.Robot, event *onebot.Event) {
	helpMsg := `群管机器人帮助信息：

普通成员命令：
- 群规：查看群规
- help：查看帮助信息

管理员命令：
- !kick <用户ID> [是否拒绝入群]：踢人
- !ban <用户ID> [时长(分钟)]：禁言
- !unban <用户ID>：解除禁言
- !addadmin <用户ID>：添加管理员
- !deladmin <用户ID>：删除管理员
- !setrules <群规内容>：设置群规
- !addword <敏感词>：添加敏感词
- !delword <敏感词>：删除敏感词
- !members：查看群成员列表
- !memberinfo <用户ID>：查看特定成员信息`

	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: helpMsg,
	})
}

// 定期检查禁言时间
func (p *GroupManagerPlugin) checkBanExpiration(robot plugin.Robot) {
	for {
		// 每隔1分钟检查一次
		time.Sleep(1 * time.Minute)

		// 检查Redis中的禁言记录
		if p.redisClient != nil {
			ctx := context.Background()
			var cursor uint64 = 0

			for {
				// 使用SCAN命令遍历所有禁言记录
				keys, nextCursor, err := p.redisClient.Scan(ctx, cursor, "group:*:ban:*", 10).Result()
				if err != nil {
					log.Printf("从Redis获取禁言记录失败: %v", err)
					break
				}

				// 处理每个禁言记录
				for _, key := range keys {
					// 获取禁言过期时间
					banEndTimeStr, err := p.redisClient.Get(ctx, key).Result()
					if err != nil {
						log.Printf("获取禁言记录失败: %v", err)
						continue
					}

					banEndTime, err := strconv.ParseInt(banEndTimeStr, 10, 64)
					if err != nil {
						log.Printf("解析禁言时间失败: %v", err)
						continue
					}

					// 检查是否过期
					if time.Now().Unix() >= banEndTime {
						// 解析groupID和userID
						parts := strings.Split(key, ":")
						if len(parts) != 4 {
							log.Printf("无效的禁言记录键: %s", key)
							continue
						}

						groupIDStr := parts[1]
						userIDStr := parts[3]
						groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
						if err != nil {
							log.Printf("转换群ID失败: %v", err)
							continue
						}
						userID, err := strconv.ParseInt(userIDStr, 10, 64)
						if err != nil {
							log.Printf("转换用户ID失败: %v", err)
							continue
						}

						// 解除禁言
						_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
							GroupID:  groupID,
							UserID:   userID,
							Duration: 0,
						})
						if err != nil {
							log.Printf("解除禁言失败: %v", err)
							continue
						}

						// 从Redis移除禁言记录
						if err := p.redisClient.Del(ctx, key).Err(); err != nil {
							log.Printf("从Redis移除禁言记录失败: %v", err)
						}

						// 同时从数据库移除禁言记录（如果存在）
						if p.db != nil {
							if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
								log.Printf("从数据库移除禁言记录失败: %v", err)
							}
						}

						// 发送通知
						robot.SendMessage(&onebot.SendMessageParams{
							GroupID: groupID,
							Message: fmt.Sprintf("用户 %d 的禁言时间已到", userID),
						})
					}
				}

				// 检查是否遍历完毕
				if nextCursor == 0 {
					break
				}
				cursor = nextCursor
			}
		}

		// 同时检查数据库中的禁言记录（作为后备）
		if p.db != nil {
			// 从数据库获取所有过期的禁言记录
			expiredBans, err := db.GetExpiredBans(p.db)
			if err != nil {
				log.Printf("获取过期禁言记录失败: %v", err)
				continue
			}

			// 遍历所有过期的禁言记录
			for _, ban := range expiredBans {
				// 转换groupID和userID为int64
				groupIDStr := ban["group_id"].(string)
				userIDStr := ban["user_id"].(string)
				groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
				if err != nil {
					log.Printf("转换群ID失败: %v", err)
					continue
				}
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err != nil {
					log.Printf("转换用户ID失败: %v", err)
					continue
				}

				// 解除禁言
				_, err = robot.SetGroupBan(&onebot.SetGroupBanParams{
					GroupID:  groupID,
					UserID:   userID,
					Duration: 0,
				})
				if err != nil {
					log.Printf("解除禁言失败: %v", err)
					continue
				}

				// 从数据库移除禁言记录
				if err := db.UnbanUser(p.db, groupIDStr, userIDStr); err != nil {
					log.Printf("移除禁言记录失败: %v", err)
					continue
				}

				// 发送通知
				robot.SendMessage(&onebot.SendMessageParams{
					GroupID: groupID,
					Message: fmt.Sprintf("用户 %d 的禁言时间已到", userID),
				})
			}
		}
	}
}

// 解析用户ID
func parseUserID(str string) (int64, error) {
	// 处理 @ 开头的用户ID
	if strings.HasPrefix(str, "@") {
		str = str[1:]
	}

	// 提取数字
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(str)
	if numStr == "" {
		return 0, fmt.Errorf("无效的用户ID")
	}

	// 转换为int64
	userID := int64(0)
	for _, c := range numStr {
		userID = userID*10 + int64(c-'0')
	}

	return userID, nil
}

// 解析时长
func parseDuration(str string) (int, error) {
	// 提取数字
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(str)
	if numStr == "" {
		return 0, fmt.Errorf("无效的时长")
	}

	// 转换为int
	duration := 0
	for _, c := range numStr {
		duration = duration*10 + int(c-'0')
	}

	return duration, nil
}

// 处理获取群成员列表命令
func (p *GroupManagerPlugin) handleGetMembersCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	// 只有管理员可以查看群成员列表
	if !p.isAdmin(event.GroupID, event.UserID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "权限不足，只有管理员可以查看群成员列表！",
		})
		return
	}

	// 调用OneBot API获取群成员列表
	resp, err := robot.GetGroupMemberList(&onebot.GetGroupMemberListParams{
		GroupID: event.GroupID,
		NoCache: true,
	})

	if err != nil {
		log.Printf("[GroupManager] 获取群 %d 成员列表失败: %v", event.GroupID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: fmt.Sprintf("获取群成员列表失败: %v", err),
		})
		return
	}

	// 解析返回数据
	memberList, ok := resp.Data.([]interface{})
	if !ok {
		log.Printf("[GroupManager] 解析群成员列表数据失败: %T", resp.Data)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "解析群成员列表数据失败",
		})
		return
	}

	// 格式化群成员信息
	var membersInfo strings.Builder
	membersInfo.WriteString(fmt.Sprintf("群 %d 成员列表 (共%d人):\n\n", event.GroupID, len(memberList)))

	for i, member := range memberList {
		memberMap, ok := member.(map[string]interface{})
		if !ok {
			continue
		}

		userID, _ := memberMap["user_id"].(float64)
		nickname, _ := memberMap["nickname"].(string)
		card, _ := memberMap["card"].(string)
		sex, _ := memberMap["sex"].(string)
		joinTime, _ := memberMap["join_time"].(float64)

		// 显示群名片或昵称
		name := nickname
		if card != "" {
			name = card
		}

		// 格式化加入时间
		joinDate := time.Unix(int64(joinTime), 0).Format("2006-01-02")

		// 添加到信息字符串
		membersInfo.WriteString(fmt.Sprintf("%d. ID: %d | 昵称: %s | 性别: %s | 入群时间: %s\n", 
			i+1, int64(userID), name, sex, joinDate))

		// 每50个成员发送一次消息，避免消息过长
		if (i+1)%50 == 0 || i == len(memberList)-1 {
			robot.SendMessage(&onebot.SendMessageParams{
				GroupID: event.GroupID,
				Message: membersInfo.String(),
			})
			membersInfo.Reset()
			membersInfo.WriteString(fmt.Sprintf("群 %d 成员列表 (续):\n\n", event.GroupID))
		}
	}
}

// 处理获取群成员信息命令
func (p *GroupManagerPlugin) handleGetMemberInfoCommand(robot plugin.Robot, event *onebot.Event, args []string) {
	// 只有管理员可以查看群成员信息
	if !p.isAdmin(event.GroupID, event.UserID) {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "权限不足，只有管理员可以查看群成员信息！",
		})
		return
	}

	if len(args) < 1 {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "用法: !memberinfo <用户ID>",
		})
		return
	}

	// 解析用户ID
	userID, err := parseUserID(args[0])
	if err != nil {
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "无效的用户ID！",
		})
		log.Printf("[GroupManager] 解析用户ID '%s' 失败: %v", args[0], err)
		return
	}

	// 调用OneBot API获取群成员信息
	resp, err := robot.GetGroupMemberInfo(&onebot.GetGroupMemberInfoParams{
		GroupID: event.GroupID,
		UserID:  userID,
		NoCache: true,
	})

	if err != nil {
		log.Printf("[GroupManager] 获取群 %d 成员 %d 信息失败: %v", event.GroupID, userID, err)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: fmt.Sprintf("获取群成员信息失败: %v", err),
		})
		return
	}

	// 解析返回数据
	memberInfo, ok := resp.Data.(map[string]interface{})
	if !ok {
		log.Printf("[GroupManager] 解析群成员信息数据失败: %T", resp.Data)
		robot.SendMessage(&onebot.SendMessageParams{
			GroupID: event.GroupID,
			Message: "解析群成员信息数据失败",
		})
		return
	}

	// 提取成员信息
	userIDFloat, _ := memberInfo["user_id"].(float64)
	nickname, _ := memberInfo["nickname"].(string)
	card, _ := memberInfo["card"].(string)
	sex, _ := memberInfo["sex"].(string)
	age, _ := memberInfo["age"].(float64)
	joinTime, _ := memberInfo["join_time"].(float64)
	lastSentTime, _ := memberInfo["last_sent_time"].(float64)
	level, _ := memberInfo["level"].(float64)
	role, _ := memberInfo["role"].(string)

	// 显示群名片或昵称
	name := nickname
	if card != "" {
		name = card
	}

	// 格式化时间
	joinDate := time.Unix(int64(joinTime), 0).Format("2006-01-02 15:04:05")
	lastSentDate := time.Unix(int64(lastSentTime), 0).Format("2006-01-02 15:04:05")

	// 格式化成员信息
	memberDetail := fmt.Sprintf(
		"成员信息:\n"
		"ID: %d\n"
		"昵称: %s\n"
		"群名片: %s\n"
		"性别: %s\n"
		"年龄: %d\n"
		"入群时间: %s\n"
		"最后发言: %s\n"
		"群等级: %d\n"
		"角色: %s",
		int64(userIDFloat), name, card, sex, int(age), joinDate, lastSentDate, int(level), role)

	// 发送成员信息
	robot.SendMessage(&onebot.SendMessageParams{
		GroupID: event.GroupID,
		Message: memberDetail,
	})
}
