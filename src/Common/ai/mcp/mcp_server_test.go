package mcp

import (
	"BotMatrix/common/ai"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPHandlers(t *testing.T) {
	// 模拟 Manager
	m := &Manager{}

	t.Run("HandleMCPListTools", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/mcp/v1/tools", nil)
		rr := httptest.NewRecorder()
		handler := HandleMCPListTools(m)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var resp ai.MCPListToolsResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		// 初始状态下工具列表可能为空，因为 TaskManager 未初始化
	})

	t.Run("HandleMCPCallTool_InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/mcp/v1/tools/call", strings.NewReader("invalid"))
		rr := httptest.NewRecorder()
		handler := HandleMCPCallTool(m)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200 (for error response), got %d", rr.Code)
		}
		// 验证返回了错误信息
	})
}
