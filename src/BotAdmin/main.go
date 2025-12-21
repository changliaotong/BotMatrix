package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"BotMatrix/common"
)

func main() {
	log.Println("启动 BotAdmin 管理后台服务...")

	// 创建管理器
	manager := common.NewManager()

	// 启动统计信息收集 (如果 BotAdmin 也负责这个)
	go manager.StartTrendCollection()
	go manager.StartPeriodicStatsSave()

	// 获取端口配置
	webuiPort := manager.Config.WebUIPort
	if webuiPort == "" {
		webuiPort = ":5000"
	}

	// 创建路由处理器
	mux := http.NewServeMux()

	// --- 公开接口 ---
	mux.HandleFunc("/login", HandleLogin(manager))
	mux.HandleFunc("/api/login", HandleLogin(manager))

	// --- 需要登录的接口 ---
	mux.HandleFunc("/api/me", manager.JWTMiddleware(HandleGetUserInfo(manager)))
	mux.HandleFunc("/api/user/info", manager.JWTMiddleware(HandleGetUserInfo(manager)))
	mux.HandleFunc("/api/user/password", manager.JWTMiddleware(HandleChangePassword(manager)))

	mux.HandleFunc("/api/bots", manager.JWTMiddleware(HandleGetBots(manager)))
	mux.HandleFunc("/api/workers", manager.JWTMiddleware(HandleGetWorkers(manager)))
	mux.HandleFunc("/api/proxy/avatar", HandleProxyAvatar(manager))
	mux.HandleFunc("/api/stats", manager.JWTMiddleware(HandleGetStats(manager)))
	mux.HandleFunc("/api/stats/chat", manager.JWTMiddleware(HandleGetChatStats(manager)))
	mux.HandleFunc("/api/system/stats", manager.JWTMiddleware(HandleGetSystemStats(manager)))
	mux.HandleFunc("/api/logs", manager.JWTMiddleware(HandleGetLogs(manager)))
	mux.HandleFunc("/api/contacts", manager.JWTMiddleware(HandleGetContacts(manager)))
	mux.HandleFunc("/api/action", manager.JWTMiddleware(HandleSendAction(manager)))
	mux.HandleFunc("/api/smart_action", manager.JWTMiddleware(HandleSendAction(manager)))

	// --- 需要管理员权限的接口 ---
	mux.HandleFunc("/api/admin/config", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetConfig(manager)(w, r)
		case http.MethodPost:
			HandleUpdateConfig(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/users", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleAdminListUsers(manager)(w, r)
		case http.MethodPost:
			HandleAdminManageUsers(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/admin/routing", manager.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGetRoutingRules(manager)(w, r)
		case http.MethodPost:
			HandleSetRoutingRule(manager)(w, r)
		case http.MethodDelete:
			HandleDeleteRoutingRule(manager)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// --- Docker 接口 ---
	mux.HandleFunc("/api/docker/list", manager.AdminMiddleware(HandleDockerList(manager)))
	mux.HandleFunc("/api/docker/action", manager.AdminMiddleware(HandleDockerAction(manager)))
	mux.HandleFunc("/api/docker/add-bot", manager.AdminMiddleware(HandleDockerAddBot(manager)))
	mux.HandleFunc("/api/docker/add-worker", manager.AdminMiddleware(HandleDockerAddWorker(manager)))

	// --- WebSocket 接口 (仅供管理后台 UI 使用) ---
	mux.HandleFunc("/ws/subscriber", manager.JWTMiddleware(HandleSubscriberWebSocket(manager)))

	// --- 静态文件服务 ---
	webDir := "../WebUI/web"
	if _, err := os.Stat("./web"); err == nil {
		webDir = "./web"
	} else if _, err := os.Stat("/app/web"); err == nil {
		webDir = "/app/web"
	}
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	// Overmind (Flutter Web) 服务
	overmindDir := "../WebUI/overmind"
	if _, err := os.Stat("./overmind"); err == nil {
		overmindDir = "./overmind"
	} else if _, err := os.Stat("/app/overmind"); err == nil {
		overmindDir = "/app/overmind"
	}
	mux.Handle("/overmind/", http.StripPrefix("/overmind/", http.FileServer(http.Dir(overmindDir))))

	// 启动服务器
	server := &http.Server{
		Addr:    webuiPort,
		Handler: mux,
	}

	go func() {
		log.Printf("[Admin] 管理后台启动在端口 %s", webuiPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Admin] 启动失败: %v", err)
		}
	}()

	// 监听信号以优雅退出
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("正在关闭 BotAdmin 服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}

	// 关闭数据库连接
	if manager.DB != nil {
		manager.DB.Close()
		log.Printf("[INFO] 数据库已关闭")
	}

	log.Println("服务已关闭")
}
