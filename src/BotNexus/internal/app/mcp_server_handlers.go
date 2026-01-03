package app

import (
	"BotMatrix/common/ai"
	clog "BotMatrix/common/log"
	"BotMatrix/common/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HandleMCPSSE 处理 MCP SSE 连接
// GET /api/mcp/v1/sse
func HandleMCPSSE(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 SSE 响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 发送初始 endpoint 事件，告知客户端后续请求应该发往何处
		// 参考 MCP SSE 规范
		fmt.Fprintf(w, "event: endpoint\ndata: /api/mcp/v1/tools/call\n\n")
		flusher.Flush()

		clog.Info("[MCP] SSE client connected")

		// 保持连接
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 发送心跳包
				fmt.Fprintf(w, ": keep-alive\n\n")
				flusher.Flush()
			case <-r.Context().Done():
				clog.Info("[MCP] SSE client disconnected")
				return
			}
		}
	}
}

// HandleMCPListTools 处理 MCP tools/list 请求
// GET /api/mcp/v1/tools
func HandleMCPListTools(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := NewInternalSkillMCPHost(m)
		tools, err := host.ListTools(r.Context(), "internal")
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to list tools: "+err.Error(), nil)
			return
		}

		resp := ai.MCPListToolsResponse{
			Tools: tools,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleMCPCallTool 处理 MCP tools/call 请求
// POST /api/mcp/v1/tools/call
func HandleMCPCallTool(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ai.MCPCallToolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.MCPManager == nil {
			utils.SendJSONResponse(w, false, "MCP Manager not initialized", nil)
			return
		}

		// 使用 MCPManager 调用工具，它会自动处理隐私脱敏
		result, err := m.MCPManager.CallTool(r.Context(), req.Name, req.Arguments)

		if err != nil {
			resp := ai.MCPCallToolResponse{
				IsError: true,
				Content: []ai.MCPContent{
					{Type: "text", Text: err.Error()},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// 将结果转换为 MCPContent 格式
		var content []ai.MCPContent
		if mcpResp, ok := result.(ai.MCPCallToolResponse); ok {
			content = mcpResp.Content
		} else {
			resData, _ := json.Marshal(result)
			content = []ai.MCPContent{
				{Type: "text", Text: string(resData)},
			}
		}

		resp := ai.MCPCallToolResponse{
			Content: content,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
