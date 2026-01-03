package app

import (
	"BotMatrix/common/utils"
	"encoding/json"
	"net/http"
)

// HandleMeshDiscover 处理服务发现请求
// GET /api/mesh/discover?q=keyword
func HandleMeshDiscover(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 执行联邦搜索 (包含本地和远程已连接的企业)
		allServers, err := m.B2BService.DiscoverMeshEndpoints(query)
		if err != nil {
			utils.SendJSONResponse(w, false, "Mesh discovery failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Discovery successful", allServers)
	}
}

// HandleMeshRegister 处理服务注册请求
// POST /api/mesh/register
func HandleMeshRegister(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Endpoint string `json:"endpoint"`
			EntID    uint   `json:"ent_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := m.B2BService.RegisterEndpoint(req.EntID, req.Name, req.Type, req.Endpoint)
		if err != nil {
			utils.SendJSONResponse(w, false, "Registration failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Registration successful", nil)
	}
}

// HandleMeshConnect 处理企业间握手请求
// POST /api/mesh/connect
func HandleMeshConnect(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SourceCode string `json:"source_code"`
			TargetCode string `json:"target_code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := m.B2BService.Connect(req.SourceCode, req.TargetCode)
		if err != nil {
			utils.SendJSONResponse(w, false, "Connection failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "B2B Connection established", nil)
	}
}

// HandleMeshCall 处理跨域工具调用代理
// POST /api/mesh/call
func HandleMeshCall(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetEntID uint           `json:"target_ent_id"`
			ToolName    string         `json:"tool_name"`
			Arguments   map[string]any `json:"arguments"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID (从 JWT Claims 中获取，假设是企业管理员在调用)
		// 这里简化逻辑，实际应从 ctx 中提取
		fromEntID := uint(1) // 默认模拟为 ID 1

		result, err := m.B2BService.CallRemoteTool(fromEntID, req.TargetEntID, req.ToolName, req.Arguments)
		if err != nil {
			utils.SendJSONResponse(w, false, "Mesh call failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Mesh call successful", result)
	}
}
