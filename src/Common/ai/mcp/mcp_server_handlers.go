package mcp

import (
	"BotMatrix/common/ai"
	clog "BotMatrix/common/log"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HandleMCPSSE 处理 MCP SSE 连接
// @Summary MCP SSE 传输
// @Description 建立符合 MCP 规范的 SSE 连接，用于双向异步通信
// @Tags MCP
// @Produce text/event-stream
// @Success 200 {string} string "SSE Event Stream"
// @Router /api/mcp/v1/sse [get]
func HandleMCPSSE(m types.Manager) http.HandlerFunc {
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
// @Summary 列出 MCP 工具
// @Description 获取当前节点暴露的所有 MCP 工具列表
// @Tags MCP
// @Produce json
// @Success 200 {object} ai.MCPListToolsResponse "工具列表"
// @Router /api/mcp/v1/tools [get]
func HandleMCPListTools(m types.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := NewInternalSkillMCPHost(NewInternalSkillProviderImpl(m))
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
// @Summary 调用 MCP 工具
// @Description 调用指定的 MCP 工具并返回执行结果，支持隐私数据脱敏
// @Tags MCP
// @Accept json
// @Produce json
// @Param body body ai.MCPCallToolRequest true "调用请求"
// @Success 200 {object} ai.MCPCallToolResponse "调用结果"
// @Router /api/mcp/v1/tools/call [post]
func HandleMCPCallTool(m types.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ai.MCPCallToolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		mcpManager := m.GetMCPManager()
		if mcpManager == nil {
			utils.SendJSONResponse(w, false, "MCP Manager not initialized", nil)
			return
		}

		result, err := mcpManager.CallTool(r.Context(), req.Name, req.Arguments)
		if err != nil {
			utils.SendJSONResponse(w, false, "Tool execution failed: "+err.Error(), nil)
			return
		}

		resp := ai.MCPCallToolResponse{
			Content: []ai.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("%v", result),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
