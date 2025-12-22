// BotNexus - 统一构建入口文件
package main

import (
	"BotMatrix/common"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// 版本号定义
const VERSION = "86"

// Manager 是 BotNexus 本地的包装结构，允许在其上定义方法
type Manager struct {
	*common.Manager
	Core *CorePlugin
}

// 主函数 - 整合所有功能
func main() {
	log.Printf("启动 BotNexus 服务... 版本号: %s", VERSION)

	// 创建管理器 (内部会初始化数据库和管理员)
	manager := NewManager()

	// 启动超时检测
	go manager.StartWorkerTimeoutDetection()
	go manager.StartBotTimeoutDetection()

	// 启动统计信息收集
	go manager.StartTrendCollection()

	// 启动统计信息重置和定期保存
	go manager.StartPeriodicStatsSave()

	// 启动 Core Gateway (WebSocket 转发引擎 - 仅处理机器人和工作节点连接)
	log.Printf("[Core] WebSocket 转发引擎启动在端口 %s", common.WS_PORT)
	coreMux := manager.createCoreHandler()
	if err := http.ListenAndServe(common.WS_PORT, coreMux); err != nil {
		log.Fatalf("[Core] 启动失败: %v", err)
	}
}

func (m *Manager) createCoreHandler() http.Handler {
	mux := http.NewServeMux()

	// 仅处理转发核心的 WebSocket 连接
	mux.HandleFunc("/ws/bots", m.handleBotWebSocket)
	mux.HandleFunc("/ws/workers", m.handleWorkerWebSocket)

	return mux
}

// 简化的管理器创建函数
func NewManager() *Manager {
	m := &Manager{
		Manager: common.NewManager(),
	}

	// 初始化数据库
	if err := m.InitDB(); err != nil {
		log.Printf("[ERROR] 数据库初始化失败: %v", err)
	} else {
		// 从数据库加载路由规则
		if err := m.LoadRoutingRulesFromDB(); err != nil {
			log.Printf("[WARN] 加载路由规则失败: %v", err)
		}
		// 从数据库加载联系人缓存
		if err := m.LoadCachesFromDB(); err != nil {
			log.Printf("[WARN] 加载联系人缓存失败: %v", err)
		}
		// 从数据库加载系统统计
		if err := m.LoadStatsFromDB(); err != nil {
			log.Printf("[WARN] 加载系统统计失败: %v", err)
		}
	}

	// 初始化Redis (用于统计信息等非持久化数据)
	m.Rdb = redis.NewClient(&redis.Options{
		Addr:     common.REDIS_ADDR,
		Password: common.REDIS_PWD,
		DB:       0,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.Rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARN] 无法连接到Redis: %v", err)
		m.Rdb = nil
	} else {
		log.Printf("[INFO] 已连接到Redis")
	}

	// 初始化核心插件
	m.Core = NewCorePlugin(m)

	return m
}
