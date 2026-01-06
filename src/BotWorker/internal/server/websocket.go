package server

import (
	"BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"botworker/internal/config"
	"botworker/internal/onebot"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	config          *config.WebSocketConfig
	upgrader        websocket.Upgrader
	clients         map[*websocket.Conn]bool
	clientsMutex    sync.Mutex
	messageHandlers []onebot.EventHandler
	noticeHandlers  []onebot.EventHandler
	requestHandlers []onebot.EventHandler
	eventHandlers   map[string][]onebot.EventHandler
	apiHandlers     map[string]onebot.RequestHandler
	middlewares     []MiddlewareFunc
	closeChan       chan struct{}
}

// NewWebSocketServer 创建WebSocket服务器实例
func NewWebSocketServer(cfg *config.WebSocketConfig) *WebSocketServer {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg.WebSocket
	}

	// 创建upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return cfg.CheckOrigin
		},
	}

	return &WebSocketServer{
		config:          cfg,
		upgrader:        upgrader,
		clients:         make(map[*websocket.Conn]bool),
		eventHandlers:   make(map[string][]onebot.EventHandler),
		apiHandlers:     make(map[string]onebot.RequestHandler),
		closeChan:       make(chan struct{}),
		messageHandlers: []onebot.EventHandler{},
		noticeHandlers:  []onebot.EventHandler{},
		requestHandlers: []onebot.EventHandler{},
	}
}

func (s *WebSocketServer) OnMessage(fn onebot.EventHandler) {
	s.messageHandlers = append(s.messageHandlers, fn)
}

func (s *WebSocketServer) OnNotice(fn onebot.EventHandler) {
	s.noticeHandlers = append(s.noticeHandlers, fn)
}

func (s *WebSocketServer) OnRequest(fn onebot.EventHandler) {
	s.requestHandlers = append(s.requestHandlers, fn)
}

func (s *WebSocketServer) OnEvent(eventName string, fn onebot.EventHandler) {
	s.eventHandlers[eventName] = append(s.eventHandlers[eventName], fn)
}

func (s *WebSocketServer) HandleAPI(action string, fn onebot.RequestHandler) {
	s.apiHandlers[action] = fn
}

func (s *WebSocketServer) BroadcastJSON(v any) error {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	var lastErr error
	for client := range s.clients {
		if err := client.WriteJSON(v); err != nil {
			log.Printf("[WS] Failed to send JSON to client: %v", err)
			lastErr = err
		}
	}
	return lastErr
}

func (s *WebSocketServer) handleConnection(conn *websocket.Conn) {
	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息错误:", err)
			break
		}

		var event onebot.Event
		if err := json.Unmarshal(message, &event); err != nil {
			var request onebot.Request
			if err := json.Unmarshal(message, &request); err != nil {
				log.Println("解析消息错误:", err)
				continue
			}
			// 处理API请求
			s.handleAPIRequest(conn, request)
			continue
		}

		// 处理事件
		processEventIDs(&event)
		s.handleEvent(event)
	}
}

func (s *WebSocketServer) handleEvent(event onebot.Event) {
	s.DispatchEvent(&event)
}

func (s *WebSocketServer) DispatchEvent(event *onebot.Event) {
	// 分发到对应的事件处理器
	switch event.PostType {
	case "message":
		for _, handler := range s.messageHandlers {
			if err := handler(event); err != nil {
				log.Println("消息处理错误:", err)
			}
		}
	case "notice":
		for _, handler := range s.noticeHandlers {
			if err := handler(event); err != nil {
				log.Println("通知处理错误:", err)
			}
		}
	case "request":
		for _, handler := range s.requestHandlers {
			if err := handler(event); err != nil {
				log.Println("请求处理错误:", err)
			}
		}
	}

	// 分发到命名事件处理器
	eventName := event.EventName
	if eventName == "" {
		eventName = event.PostType
	}
	if handlers, ok := s.eventHandlers[eventName]; ok {
		for _, handler := range handlers {
			if err := handler(event); err != nil {
				log.Println("事件处理错误:", err)
			}
		}
	}
}

func (s *WebSocketServer) handleAPIRequest(conn *websocket.Conn, request onebot.Request) {
	handler, ok := s.apiHandlers[request.Action]
	if !ok {
		response := onebot.Response{
			Status:  "failed",
			Message: fmt.Sprintf("未知的API动作: %s", request.Action),
			Echo:    request.Echo,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Println("发送API响应错误:", err)
		}
		return
	}

	response, err := handler(&request)
	if err != nil {
		response = &onebot.Response{
			Status:  "failed",
			Message: err.Error(),
			Echo:    request.Echo,
		}
	} else {
		response.Status = "ok"
		response.Echo = request.Echo
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Println("发送API响应错误:", err)
	}
}

func (s *WebSocketServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("升级WebSocket连接错误:", err)
			return
		}

		// 设置连接超时时间
		conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
		conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(s.config.PongTimeout))
			return nil
		})

		s.handleConnection(conn)
	})

	server := &http.Server{
		Addr:         s.config.Addr,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	go func() {
		<-s.closeChan
		server.Close()
	}()

	log.Printf("WebSocket服务器启动在 %s\n", s.config.Addr)
	return server.ListenAndServe()
}

func (s *WebSocketServer) Stop() {
	close(s.closeChan)
}

func (s *WebSocketServer) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	return s.CallAction("send_msg", params)
}

func (s *WebSocketServer) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	return s.CallAction("delete_msg", params)
}

func (s *WebSocketServer) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	return s.CallAction("send_like", params)
}

func (s *WebSocketServer) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	return s.CallAction("set_group_kick", params)
}

func (s *WebSocketServer) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	return s.CallAction("set_group_ban", params)
}

func (s *WebSocketServer) GetGroupMemberList(params *onebot.GetGroupMemberListParams) (*onebot.Response, error) {
	return s.CallAction("get_group_member_list", params)
}

func (s *WebSocketServer) GetGroupMemberInfo(params *onebot.GetGroupMemberInfoParams) (*onebot.Response, error) {
	return s.CallAction("get_group_member_info", params)
}

func (s *WebSocketServer) SetGroupSpecialTitle(params *onebot.SetGroupSpecialTitleParams) (*onebot.Response, error) {
	return s.CallAction("set_group_special_title", params)
}

func (s *WebSocketServer) GetSelfID() int64 {
	// 返回默认的机器人ID，实际实现中应该从配置或连接状态获取
	return 123456789
}

func (s *WebSocketServer) CallAction(action string, params any) (*onebot.Response, error) {
	// 获取第一个客户端连接
	s.clientsMutex.Lock()
	var conn *websocket.Conn
	for c := range s.clients {
		conn = c
		break
	}
	s.clientsMutex.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("没有连接的客户端")
	}

	request := onebot.Request{
		Action: action,
		Params: params,
	}

	if err := conn.WriteJSON(request); err != nil {
		return nil, err
	}

	// 简化实现，直接返回成功响应
	// 实际实现中应该处理echo来匹配请求和响应
	return &onebot.Response{
		Status: "ok",
		Data: map[string]any{
			"message_id": 123456,
		},
	}, nil
}
