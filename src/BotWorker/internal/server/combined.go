package server

import (
	"botworker/internal/config"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
)

type CombinedServer struct {
	wsServer *WebSocketServer
	httpServer *HTTPServer
	pluginManager *plugin.Manager
}

func NewCombinedServer(cfg *config.Config) *CombinedServer {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	
	server := &CombinedServer{
		wsServer:  NewWebSocketServer(&cfg.WebSocket),
		httpServer: NewHTTPServer(&cfg.HTTP),
	}
	server.pluginManager = plugin.NewManager(server)
	return server
}

// 实现plugin.Robot接口
func (s *CombinedServer) OnMessage(fn onebot.EventHandler) {
	s.wsServer.OnMessage(fn)
	s.httpServer.OnMessage(fn)
}

func (s *CombinedServer) OnNotice(fn onebot.EventHandler) {
	s.wsServer.OnNotice(fn)
	s.httpServer.OnNotice(fn)
}

func (s *CombinedServer) OnRequest(fn onebot.EventHandler) {
	s.wsServer.OnRequest(fn)
	s.httpServer.OnRequest(fn)
}

func (s *CombinedServer) OnEvent(eventName string, fn onebot.EventHandler) {
	s.wsServer.OnEvent(eventName, fn)
	s.httpServer.OnEvent(eventName, fn)
}

func (s *CombinedServer) HandleAPI(action string, fn onebot.RequestHandler) {
	s.wsServer.HandleAPI(action, fn)
	s.httpServer.HandleAPI(action, fn)
}

func (s *CombinedServer) SendMessage(params *onebot.SendMessageParams) (*onebot.Response, error) {
	// 优先使用WebSocket发送消息
	return s.wsServer.SendMessage(params)
}

func (s *CombinedServer) DeleteMessage(params *onebot.DeleteMessageParams) (*onebot.Response, error) {
	return s.wsServer.DeleteMessage(params)
}

func (s *CombinedServer) SendLike(params *onebot.SendLikeParams) (*onebot.Response, error) {
	return s.wsServer.SendLike(params)
}

func (s *CombinedServer) SetGroupKick(params *onebot.SetGroupKickParams) (*onebot.Response, error) {
	return s.wsServer.SetGroupKick(params)
}

func (s *CombinedServer) SetGroupBan(params *onebot.SetGroupBanParams) (*onebot.Response, error) {
	return s.wsServer.SetGroupBan(params)
}

// 插件管理
func (s *CombinedServer) GetPluginManager() *plugin.Manager {
	return s.pluginManager
}

// 启动服务器
func (s *CombinedServer) Run() error {
	// 启动HTTP服务器
	go func() {
		if err := s.httpServer.Run(); err != nil {
			panic(err)
		}
	}()

	// 启动WebSocket服务器
	return s.wsServer.Run()
}

func (s *CombinedServer) Stop() {
	s.wsServer.Stop()
	s.httpServer.Stop()
}
