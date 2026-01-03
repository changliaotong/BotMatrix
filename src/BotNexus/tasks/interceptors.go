package tasks

import (
	"BotMatrix/common/types"
	"log"
	"math/rand"
	"strings"

	"gorm.io/gorm"
)

// InterceptorContext 拦截器上下文
type InterceptorContext struct {
	Platform string
	SelfID   string
	UserID   string
	GroupID  string
	Message  *types.InternalMessage
	DB       *gorm.DB // 允许拦截器访问数据库
	AI       *AIParser
}

// Interceptor 拦截器接口
type Interceptor interface {
	Name() string
	// BeforeDispatch 在分发给 Worker 之前执行
	// 返回 true 表示继续，false 表示拦截并中断后续流程
	BeforeDispatch(ctx *InterceptorContext) (bool, error)
}

// InterceptorManager 拦截器管理器
type InterceptorManager struct {
	db           *gorm.DB
	ai           *AIParser
	interceptors []Interceptor
}

func NewInterceptorManager(db *gorm.DB, ai *AIParser) *InterceptorManager {
	im := &InterceptorManager{
		db:           db,
		ai:           ai,
		interceptors: make([]Interceptor, 0),
	}
	// 注册默认拦截器
	im.Add(&StrategyInterceptor{})
	im.Add(&IdentityInterceptor{})
	im.Add(&SemanticRoutingInterceptor{})
	im.Add(&ShadowInterceptor{})
	return im
}

func (im *InterceptorManager) Add(i Interceptor) {
	im.interceptors = append(im.interceptors, i)
}

func (im *InterceptorManager) GetInterceptors() []string {
	names := make([]string, len(im.interceptors))
	for i, interceptor := range im.interceptors {
		names[i] = interceptor.Name()
	}
	return names
}

func (im *InterceptorManager) GetStrategies() []Strategy {
	var strategies []Strategy
	im.db.Find(&strategies)
	return strategies
}

func (im *InterceptorManager) GetStrategy(name string) (*Strategy, error) {
	var strategy Strategy
	if err := im.db.Where("name = ?", name).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (im *InterceptorManager) DeleteStrategy(name string) {
	im.db.Where("name = ?", name).Delete(&Strategy{})
}

func (im *InterceptorManager) ProcessBeforeDispatch(ctx *InterceptorContext) bool {
	ctx.DB = im.db
	ctx.AI = im.ai
	for _, i := range im.interceptors {
		continueFlow, err := i.BeforeDispatch(ctx)
		if err != nil {
			log.Printf("[Interceptor] %s error: %v", i.Name(), err)
		}
		if !continueFlow {
			log.Printf("[Interceptor] Event blocked by %s", i.Name())
			return false
		}
	}
	return true
}

// --- 核心拦截器实现 ---

// StrategyInterceptor 全局策略拦截器 (维护模式、全局限流)
type StrategyInterceptor struct{}

func (s *StrategyInterceptor) Name() string { return "Strategy" }
func (s *StrategyInterceptor) BeforeDispatch(ctx *InterceptorContext) (bool, error) {
	var strategies []Strategy
	if err := ctx.DB.Where("is_enabled = ?", true).Find(&strategies).Error; err != nil {
		return true, err
	}

	for _, strat := range strategies {
		switch strat.Type {
		case "maintenance":
			// 维护模式：拦截所有非管理员消息
			// 这里简单判断，实际可根据 UserID 判断是否为管理员
			log.Printf("[Interceptor] Global maintenance mode active. Blocking event.")
			return false, nil
		case "rate_limit":
			// 全局限流逻辑...
		}
	}
	return true, nil
}

// IdentityInterceptor 统一身份拦截器 (NexusUID 映射)
type IdentityInterceptor struct{}

func (i *IdentityInterceptor) Name() string { return "Identity" }
func (i *IdentityInterceptor) BeforeDispatch(ctx *InterceptorContext) (bool, error) {
	if ctx.UserID == "" {
		return true, nil
	}

	var identity UserIdentity
	err := ctx.DB.Where("platform = ? AND platform_uid = ?", ctx.Platform, ctx.UserID).First(&identity).Error
	if err == nil {
		// 注入统一身份 ID
		if ctx.Message.Extras == nil {
			ctx.Message.Extras = make(map[string]any)
		}
		ctx.Message.Extras["nexus_uid"] = identity.NexusUID
		log.Printf("[Interceptor] Identity mapped: %s:%s -> %s", ctx.Platform, ctx.UserID, identity.NexusUID)
	} else if err == gorm.ErrRecordNotFound {
		// 自动创建新身份 (可选)
		// newUID := uuid.New().String()
		// ctx.DB.Create(&UserIdentity{...})
	}

	return true, nil
}

// SemanticRoutingInterceptor 语义路由拦截器 (意图识别)
type SemanticRoutingInterceptor struct{}

func (s *SemanticRoutingInterceptor) Name() string { return "SemanticRouting" }
func (s *SemanticRoutingInterceptor) BeforeDispatch(ctx *InterceptorContext) (bool, error) {
	// 仅对文本消息进行语义识别
	msg := ctx.Message.RawMessage
	if msg == "" {
		return true, nil
	}

	// 1. 检查是否为数字员工 (Digital Employee)
	if ctx.Message.PostType == "message" && ctx.SelfID != "" {
		// 直接通过 DB 检查是否是数字员工，避免依赖注入
		var count int64
		ctx.DB.Table("digital_employees").Where("bot_id = ?", ctx.SelfID).Count(&count)
		if count > 0 {
			if ctx.Message.Extras == nil {
				ctx.Message.Extras = make(map[string]any)
			}
			// 标记为数字员工，以便后续逻辑识别
			ctx.Message.Extras["is_digital_employee"] = true
			
			// 简单的意图识别：如果是问身份，设置提示
			if strings.Contains(msg, "你是谁") || strings.Contains(msg, "介绍") || strings.Contains(msg, "是谁") {
				ctx.Message.Extras["intent_hint"] = "agent_info"
				log.Printf("[Interceptor] Digital Employee %s detected, intent: agent_info", ctx.SelfID)
			} else {
				// 通用聊天意图
				ctx.Message.Extras["intent_hint"] = "chat"
			}
		}
	}

	// 2. 优先尝试正则匹配 (Fast-Track)
	if ctx.AI != nil {
		if skill, matched := ctx.AI.MatchSkillByRegex(msg); matched {
			if ctx.Message.Extras == nil {
				ctx.Message.Extras = make(map[string]any)
			}
			// 命中正则，设置意图提示，后续路由将优先使用此提示
			ctx.Message.Extras["intent_hint"] = "skill:" + skill.Name
			log.Printf("[Interceptor] Regex match: skill=%s", skill.Name)
			return true, nil // 继续流转，由路由阶段处理
		}
	}

	// 2. 调用 AI 接口进行意图识别 (模拟)
	if strings.Contains(msg, "?") || strings.Contains(msg, "？") || strings.HasPrefix(msg, "为什么") {
		if ctx.Message.Extras == nil {
			ctx.Message.Extras = make(map[string]any)
		}
		ctx.Message.Extras["intent_hint"] = "knowledge_base"
		log.Printf("[Interceptor] Intent detected: knowledge_base")
	}

	return true, nil
}

// ShadowInterceptor 影子执行拦截器 (A/B 测试)
type ShadowInterceptor struct{}

func (s *ShadowInterceptor) Name() string { return "Shadow" }
func (s *ShadowInterceptor) BeforeDispatch(ctx *InterceptorContext) (bool, error) {
	var rules []ShadowRule
	if err := ctx.DB.Where("is_enabled = ?", true).Find(&rules).Error; err != nil {
		return true, err
	}

	for _, rule := range rules {
		// 检查匹配模式 (简化实现)
		if strings.Contains(ctx.SelfID, rule.MatchPattern) || rule.MatchPattern == "*" {
			// 流量随机采样
			if rand.Intn(100) < rule.TrafficPercent {
				if ctx.Message.Extras == nil {
					ctx.Message.Extras = make(map[string]any)
				}
				ctx.Message.Extras["shadow_worker_id"] = rule.TargetWorkerID
				log.Printf("[Interceptor] Shadow mode active: forwarding shadow copy to %s", rule.TargetWorkerID)
			}
		}
	}
	return true, nil
}

// MaintenanceInterceptor 维护模式拦截器 (保留作为向后兼容或单独开关)
type MaintenanceInterceptor struct {
	Enabled bool
}

func (m *MaintenanceInterceptor) Name() string { return "Maintenance" }
func (m *MaintenanceInterceptor) BeforeDispatch(ctx *InterceptorContext) (bool, error) {
	if m.Enabled {
		// 如果是维护模式，只允许管理员或特定指令通过
		return false, nil
	}
	return true, nil
}
