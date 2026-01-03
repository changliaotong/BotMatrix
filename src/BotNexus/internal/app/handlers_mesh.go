package app

import (
	"BotMatrix/common/utils"
	"encoding/json"
	"net/http"
)

// HandleMeshDiscover 处理服务发现请求
// @Summary 发现 Mesh 节点
// @Description 在已连接的 B2B 联邦网络中搜索可用的 MCP 节点
// @Tags Mesh
// @Accept json
// @Produce json
// @Param q query string true "搜索关键字"
// @Success 200 {object} utils.JSONResponse "搜索结果列表"
// @Router /api/mesh/discover [get]
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
// @Summary 注册 Mesh 节点
// @Description 将本地 MCP 节点注册到 Mesh 目录中，供其他企业发现
// @Tags Mesh
// @Accept json
// @Produce json
// @Param body body object true "注册信息"
// @Success 200 {object} utils.JSONResponse "注册成功"
// @Router /api/mesh/register [post]
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
// @Summary 建立 B2B 连接
// @Description 发起与远程企业的握手请求，建立信任关系
// @Tags Mesh
// @Accept json
// @Produce json
// @Param body body object true "连接请求"
// @Success 200 {object} utils.JSONResponse "连接成功"
// @Router /api/mesh/connect [post]
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

// HandleB2BHandshake 处理来自外部企业的握手请求
// @Summary 响应 B2B 握手
// @Description 接收并验证来自外部企业的握手请求 (内部接口)
// @Tags Mesh
// @Accept json
// @Produce json
// @Param body body HandshakeRequest true "握手请求"
// @Success 200 {object} HandshakeResponse "握手响应"
// @Router /api/b2b/handshake [post]
func HandleB2BHandshake(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HandshakeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		resp, err := m.B2BService.HandleHandshake(req)
		if err != nil {
			utils.SendJSONResponse(w, false, "Handshake failed: "+err.Error(), nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleMeshCall 处理跨域工具调用代理
// @Summary 跨企业工具调用
// @Description 通过 Mesh 网络调用远程企业的 MCP 工具
// @Tags Mesh
// @Accept json
// @Produce json
// @Param body body object true "调用请求"
// @Success 200 {object} utils.JSONResponse "调用结果"
// @Router /api/mesh/call [post]
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

// HandleB2BRequestSkill 处理技能共享申请
// @Summary 申请 B2B 技能共享
// @Description 向远程企业申请使用其特定的 MCP 技能
// @Tags B2B
// @Accept json
// @Produce json
// @Param body body object true "申请信息"
// @Success 200 {object} utils.JSONResponse "申请成功"
// @Router /api/b2b/skills/request [post]
func HandleB2BRequestSkill(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetEntID uint   `json:"target_ent_id"`
			SkillName   string `json:"skill_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID
		// 这里假设从 context 或 JWT 中获取，简化起见先默认 ID 1
		fromEntID := uint(1)

		err := m.B2BService.RequestSkillSharing(fromEntID, req.TargetEntID, req.SkillName)
		if err != nil {
			utils.SendJSONResponse(w, false, "Request failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Skill sharing request sent", nil)
	}
}

// HandleB2BApproveSkill 处理技能共享审批
// @Summary 审批 B2B 技能共享
// @Description 审批其他企业对本企业技能的使用申请
// @Tags B2B
// @Accept json
// @Produce json
// @Param body body object true "审批信息"
// @Success 200 {object} utils.JSONResponse "审批成功"
// @Router /api/b2b/skills/approve [post]
func HandleB2BApproveSkill(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SharingID uint   `json:"sharing_id"`
			Status    string `json:"status"` // approved, rejected, blocked
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := m.B2BService.ApproveSkillSharing(req.SharingID, req.Status)
		if err != nil {
			utils.SendJSONResponse(w, false, "Approval failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Skill sharing status updated", nil)
	}
}

// HandleB2BListSkills 获取技能共享列表
// @Summary 获取 B2B 技能共享列表
// @Description 获取本企业相关的所有技能共享记录 (作为提供方或使用方)
// @Tags B2B
// @Accept json
// @Produce json
// @Param role query string false "角色: provider (提供方) 或 consumer (使用方)"
// @Success 200 {object} utils.JSONResponse "记录列表"
// @Router /api/b2b/skills/list [get]
func HandleB2BListSkills(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")

		if m.B2BService == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID
		entID := uint(1)

		sharings, err := m.B2BService.ListSkillSharings(entID, role)
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to list skills: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Success", sharings)
	}
}
