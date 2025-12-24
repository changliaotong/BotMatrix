package core

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// WSServer WebSocket 服务器
type WSServer struct {
	config     *WebSocketConfig
	upgrader   websocket.Upgrader
	connections map[*websocket.Conn]bool
}

// NewWSServer 创建新的 WebSocket 服务器
func NewWSServer(config *WebSocketConfig) *WSServer {
	return &WSServer{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源，生产环境请根据实际情况配置
				return true
			},
		},
		connections: make(map[*websocket.Conn]bool),
	}
}

// Start 启动 WebSocket 服务器
func (s *WSServer) Start() error {
	if !s.config.Enabled {
		return nil
	}

	// 注册 WebSocket 处理函数
	http.HandleFunc(s.config.Path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("WebSocket upgrade error: %v\n", err)
			return
		}
		defer conn.Close()

		// 记录连接
		s.connections[conn] = true
		fmt.Printf("New WebSocket connection from %s\n", conn.RemoteAddr())

		// 处理消息
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("WebSocket read error: %v\n", err)
				delete(s.connections, conn)
				break
			}

			fmt.Printf("Received message: %s\n", message)

			// 简单回显消息
			err = conn.WriteMessage(websocket.TextMessage, []byte("Received: "+string(message)))
			if err != nil {
				fmt.Printf("WebSocket write error: %v\n", err)
				break
			}
		}
	})

	// 启动服务器
	addr := s.config.Host + ":" + s.config.Port
	fmt.Printf("Starting WebSocket server on %s%s\n", addr, s.config.Path)
	return http.ListenAndServe(addr, nil)
}

// Broadcast 广播消息到所有连接
func (s *WSServer) Broadcast(message []byte) {
	for conn := range s.connections {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("WebSocket broadcast error: %v\n", err)
			conn.Close()
			delete(s.connections, conn)
		}
	}
}