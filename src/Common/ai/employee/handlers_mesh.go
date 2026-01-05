package employee

import (
	"BotMatrix/common/ai/b2b"
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
func HandleMeshDiscover(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 执行联邦搜索 (包含本地和远程已连接的企业)
		allServers, err := b2bSvc.DiscoverMeshEndpoints(query)
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
func HandleMeshRegister(b2bSvc b2b.B2BService) http.HandlerFunc {
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

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := b2bSvc.RegisterEndpoint(req.EntID, req.Name, req.Type, req.Endpoint)
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
func HandleMeshConnect(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SourceCode string `json:"source_code"`
			TargetCode string `json:"target_code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := b2bSvc.Connect(req.SourceCode, req.TargetCode)
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
// @Param body body b2b.HandshakeRequest true "握手请求"
// @Success 200 {object} b2b.HandshakeResponse "握手响应"
// @Router /api/b2b/handshake [post]
func HandleB2BHandshake(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req b2b.HandshakeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if b2bSvc == nil {
			http.Error(w, "B2B Service not initialized", http.StatusInternalServerError)
			return
		}

		resp, err := b2bSvc.HandleHandshake(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
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
func HandleMeshCall(b2bSvc b2b.B2BService) http.HandlerFunc {
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

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID (从 JWT Claims 中获取，假设是企业管理员在调用)
		// 这里简化逻辑，实际应从 ctx 中提取
		fromEntID := uint(1) // 默认模拟为 ID 1

		result, err := b2bSvc.CallRemoteTool(fromEntID, req.TargetEntID, req.ToolName, req.Arguments)
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
func HandleB2BRequestSkill(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TargetEntID uint   `json:"target_ent_id"`
			SkillName   string `json:"skill_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID
		// 这里假设从 context 或 JWT 中获取，简化起见先默认 ID 1
		fromEntID := uint(1)

		err := b2bSvc.RequestSkillSharing(fromEntID, req.TargetEntID, req.SkillName)
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
func HandleB2BApproveSkill(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SharingID uint   `json:"sharing_id"`
			Status    string `json:"status"` // approved, rejected, blocked
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := b2bSvc.ApproveSkillSharing(req.SharingID, req.Status)
		if err != nil {
			utils.SendJSONResponse(w, false, "Approval failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Skill sharing approved", nil)
	}
}

// HandleB2BListSkills 列出共享技能
func HandleB2BListSkills(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entID := uint(1) // 简化：从 ctx 获取
		role := r.URL.Query().Get("role")

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		sharings, err := b2bSvc.ListSkillSharings(entID, role)
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to list skills: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Success", sharings)
	}
}

// HandleB2BDispatchEmployee 处理员工外派申请
func HandleB2BDispatchEmployee(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EmployeeID  uint     `json:"employee_id"`
			TargetEntID uint     `json:"target_ent_id"`
			Permissions []string `json:"permissions"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		// 获取当前调用者的企业 ID
		sourceEntID := uint(1) // 简化：从 ctx 获取

		err := b2bSvc.DispatchEmployee(req.EmployeeID, sourceEntID, req.TargetEntID, req.Permissions)
		if err != nil {
			utils.SendJSONResponse(w, false, "Dispatch request failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Employee dispatch request sent", nil)
	}
}

// HandleB2BApproveDispatch 处理外派审批
func HandleB2BApproveDispatch(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DispatchID uint   `json:"dispatch_id"`
			Status     string `json:"status"` // approved, rejected
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "Invalid request body", nil)
			return
		}

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		err := b2bSvc.ApproveDispatch(req.DispatchID, req.Status)
		if err != nil {
			utils.SendJSONResponse(w, false, "Approval failed: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Dispatch request "+req.Status, nil)
	}
}

// HandleB2BListDispatches 列出外派记录
func HandleB2BListDispatches(b2bSvc b2b.B2BService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entID := uint(1) // 简化：从 ctx 获取
		role := r.URL.Query().Get("role")

		if b2bSvc == nil {
			utils.SendJSONResponse(w, false, "B2B Service not initialized", nil)
			return
		}

		dispatches, err := b2bSvc.ListDispatchedEmployees(entID, role)
		if err != nil {
			utils.SendJSONResponse(w, false, "Failed to list dispatches: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "Success", dispatches)
	}
}
