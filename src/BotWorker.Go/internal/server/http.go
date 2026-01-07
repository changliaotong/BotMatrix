package server

import (
	"BotMatrix/common/log"
	"encoding/json"
	"fmt"
	"net/http"

	"botworker/internal/config"
	"botworker/internal/onebot"
)

type HTTPServer struct {
	config          *config.HTTPConfig
	messageHandlers []onebot.EventHandler
	noticeHandlers  []onebot.EventHandler
	requestHandlers []onebot.EventHandler
	eventHandlers   map[string][]onebot.EventHandler
	apiHandlers     map[string]onebot.RequestHandler
	middlewares     []MiddlewareFunc
	closeChan       chan struct{}
}

func NewHTTPServer(cfg *config.HTTPConfig) *HTTPServer {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		defaultCfg := config.DefaultConfig()
		cfg = &defaultCfg.HTTP
	}

	return &HTTPServer{
		config:          cfg,
		eventHandlers:   make(map[string][]onebot.EventHandler),
		apiHandlers:     make(map[string]onebot.RequestHandler),
		closeChan:       make(chan struct{}),
		messageHandlers: []onebot.EventHandler{},
		noticeHandlers:  []onebot.EventHandler{},
		requestHandlers: []onebot.EventHandler{},
	}
}

func (s *HTTPServer) Use(m ...MiddlewareFunc) {
	s.middlewares = append(s.middlewares, m...)
}

func (s *HTTPServer) OnMessage(fn onebot.EventHandler) {
	s.messageHandlers = append(s.messageHandlers, fn)
}

func (s *HTTPServer) OnNotice(fn onebot.EventHandler) {
	s.noticeHandlers = append(s.noticeHandlers, fn)
}

func (s *HTTPServer) OnRequest(fn onebot.EventHandler) {
	s.requestHandlers = append(s.requestHandlers, fn)
}

func (s *HTTPServer) OnEvent(eventName string, fn onebot.EventHandler) {
	s.eventHandlers[eventName] = append(s.eventHandlers[eventName], fn)
}

func (s *HTTPServer) HandleAPI(action string, fn onebot.RequestHandler) {
	s.apiHandlers[action] = fn
}

func (s *HTTPServer) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var event onebot.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Println("解析事件错误:", err)
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	// 处理 QQGuild ID 生成
	processEventIDs(&event)

	// 构建并执行中间件链
	finalHandler := func(e *onebot.Event) error {
		s.dispatchEvent(*e)
		return nil
	}

	wrappedHandler := ChainMiddleware(s.middlewares, finalHandler)
	wrappedHandler(&event)

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}

func (s *HTTPServer) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var request onebot.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println("解析API请求错误:", err)
		http.Error(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	handler, ok := s.apiHandlers[request.Action]
	if !ok {
		response := onebot.Response{
			Status:  "failed",
			Message: fmt.Sprintf("未知的API动作: %s", request.Action),
			Echo:    request.Echo,
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response, err := handler(&request)
	if err != nil {
		response = &onebot.Response{
			Status:  "failed",
			Message: err.Error(),
			Echo:    request.Echo,
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Status = "ok"
		response.Echo = request.Echo
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) dispatchEvent(event onebot.Event) {
	// 分发到对应的事件处理器
	switch event.PostType {
	case "message":
		for _, handler := range s.messageHandlers {
			if err := handler(&event); err != nil {
				log.Println("消息处理错误:", err)
			}
		}
	case "notice":
		for _, handler := range s.noticeHandlers {
			if err := handler(&event); err != nil {
				log.Println("通知处理错误:", err)
			}
		}
	case "request":
		for _, handler := range s.requestHandlers {
			if err := handler(&event); err != nil {
				log.Println("请求处理错误:", err)
			}
		}
	}

	// 分发到命名事件处理器
	if handlers, ok := s.eventHandlers[event.EventName]; ok {
		for _, handler := range handlers {
			if err := handler(&event); err != nil {
				log.Println("事件处理错误:", err)
			}
		}
	}
}

func (s *HTTPServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/event", s.handleEvent)
	mux.HandleFunc("/api", s.handleAPIRequest)

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

	log.Printf("HTTP服务器启动在 %s\n", s.config.Addr)
	return server.ListenAndServe()
}

func (s *HTTPServer) Stop() {
	close(s.closeChan)
}

// HTTP服务器的API方法实现
func (s *HTTPServer) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	return &onebot.Response{
		Status: "ok",
		Data: map[string]any{
			"message_id": 123456,
		},
	}, nil
}

func (s *HTTPServer) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	return &onebot.Response{
		Status: "ok",
		Data: map[string]any{
			"message_id": params.MessageID,
		},
	}, nil
}

func (s *HTTPServer) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	return &onebot.Response{
		Status: "ok",
		Data:   nil,
	}, nil
}

func (s *HTTPServer) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	return &onebot.Response{
		Status: "ok",
		Data:   nil,
	}, nil
}

func (s *HTTPServer) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	return &onebot.Response{
		Status: "ok",
		Data:   nil,
	}, nil
}
