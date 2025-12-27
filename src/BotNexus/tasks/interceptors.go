package tasks

import (
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
	Event    map[string]interface{}
	DB       *gorm.DB // 允许拦截器访问数据库
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
	interceptors []Interceptor
}

func NewInterceptorManager(db *gorm.DB) *InterceptorManager {
	im := &InterceptorManager{
		db:           db,
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

func (im *InterceptorManager) ProcessBeforeDispatch(ctx *InterceptorContext) bool {
	ctx.DB = im.db
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
		ctx.Event["nexus_uid"] = identity.NexusUID
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
	msg, ok := ctx.Event["message"].(string)
	if !ok || msg == "" {
		return true, nil
	}

	// 这里应该调用 AI 接口进行意图识别
	// 模拟识别：如果是问题，打上 question 标签
	if strings.Contains(msg, "?") || strings.Contains(msg, "？") || strings.HasPrefix(msg, "为什么") {
		ctx.Event["intent_hint"] = "knowledge_base"
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
				ctx.Event["shadow_worker_id"] = rule.TargetWorkerID
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
