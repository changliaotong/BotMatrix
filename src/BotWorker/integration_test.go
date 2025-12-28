package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestNexusWorkerHandshake 模拟 Nexus 和 Worker 之间的 WebSocket 握手
func TestNexusWorkerHandshake(t *testing.T) {
	upgrader := websocket.Upgrader{}

	// 1. 模拟 Nexus (Server)
	nexusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Failed to upgrade to websocket: %v", err)
			return
		}
		defer conn.Close()

		// 检查握手信息或发送初始化指令
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if string(message) == "PING" {
				err = conn.WriteMessage(mt, []byte("PONG"))
				if err != nil {
					break
				}
			}
		}
	}))
	defer nexusServer.Close()

	// 2. 模拟 Worker (Client)
	wsURL := "ws" + strings.TrimPrefix(nexusServer.URL, "http")
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to mock Nexus: %v", err)
	}
	defer conn.Close()

	// 3. 验证通信
	err = conn.WriteMessage(websocket.TextMessage, []byte("PING"))
	if err != nil {
		t.Fatalf("Failed to send PING: %v", err)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read PONG: %v", err)
	}

	if string(message) != "PONG" {
		t.Errorf("Expected PONG, got %s", string(message))
	}
}

// TestBotWorkerPluginLoading 模拟插件加载逻辑
func TestBotWorkerPluginLoading(t *testing.T) {
	// 这里可以添加模拟插件加载和消息分发的逻辑
	// 详见 src/BotWorker/internal/plugin/plugin.go
	t.Log("Integration test for plugin loading would go here")
}

// 提示：集成测试建议使用专门的测试数据库，或者使用内存数据库如 SQLite
func TestDatabaseIntegration(t *testing.T) {
	// skip if no real DB
	t.Skip("Skipping real database integration test")
}
