package ai

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- AI Providers ---

// HandleListAIProviders 获取提供商列表
// @Summary 获取 AI 提供商列表
// @Description 获取所有已配置的 AI 模型提供商（如 OpenAI, Anthropic 等），API Key 会被脱敏
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "提供商列表"
// @Router /api/admin/ai/providers [get]
func HandleListAIProviders(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var providers []models.AIProvider
		db := m.GetGORMDB()
		if err := db.Find(&providers).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取提供商失败: "+err.Error(), nil)
			return
		}
		// 隐藏 API Key
		for i := range providers {
			if providers[i].APIKey != "" {
				providers[i].APIKey = "********"
			}
		}
		utils.SendJSONResponse(w, true, "", providers)
	}
}

// HandleSaveAIProvider 保存提供商 (新增或修改)
// @Summary 保存 AI 提供商
// @Description 新增或更新 AI 提供商配置。如果 API Key 为 "********" 则保持原值。
// @Tags AI Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.AIProviderGORM true "提供商信息"
// @Success 200 {object} utils.JSONResponse "保存成功"
// @Router /api/admin/ai/providers [post]
func HandleSaveAIProvider(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var provider models.AIProvider
		if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		db := m.GetGORMDB()
		var err error
		if provider.ID > 0 {
			// 如果是更新，且 APIKey 是 "********"，说明用户没有修改 Key，需要保留原有的 Key
			if provider.APIKey == "********" {
				var oldProvider models.AIProvider
				if err := db.First(&oldProvider, provider.ID).Error; err == nil {
					provider.APIKey = oldProvider.APIKey
				}
			}
			fmt.Printf("[DEBUG] Saving AI Provider %d (%s), Key length: %d\n", provider.ID, provider.Name, len(provider.APIKey))
			err = db.Save(&provider).Error
		} else {
			fmt.Printf("[DEBUG] Creating AI Provider (%s), Key length: %d\n", provider.Name, len(provider.APIKey))
			err = db.Create(&provider).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存提供商失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", provider)
	}
}

// HandleDeleteAIProvider 删除提供商
// @Summary 删除 AI 提供商
// @Description 根据 ID 删除指定的 AI 提供商配置
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "提供商 ID"
// @Success 200 {object} utils.JSONResponse "删除成功"
// @Router /api/admin/ai/providers/{id} [delete]
func HandleDeleteAIProvider(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/providers/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		db := m.GetGORMDB()
		if err := db.Delete(&models.AIProvider{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// --- AI Models ---

// HandleListAIModels 获取模型列表
// @Summary 获取 AI 模型列表
// @Description 获取所有已配置的 AI 模型，包含其关联的提供商信息
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "模型列表"
// @Router /api/admin/ai/models [get]
func HandleListAIModels(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var modelsList []models.AIModel
		db := m.GetGORMDB()
		if err := db.Preload("Provider").Find(&modelsList).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取模型失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", modelsList)
	}
}

// HandleSaveAIModel 保存模型
// @Summary 保存 AI 模型
// @Description 新增或更新 AI 模型配置
// @Tags AI Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.AIModelGORM true "模型信息"
// @Success 200 {object} utils.JSONResponse "保存成功"
// @Router /api/admin/ai/models [post]
func HandleSaveAIModel(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model models.AIModel
		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		db := m.GetGORMDB()
		var err error
		if model.ID > 0 {
			err = db.Save(&model).Error
		} else {
			err = db.Create(&model).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存模型失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", model)
	}
}

// HandleDeleteAIModel 删除模型
// @Summary 删除 AI 模型
// @Description 根据 ID 删除指定的 AI 模型配置
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "模型 ID"
// @Success 200 {object} utils.JSONResponse "删除成功"
// @Router /api/admin/ai/models/{id} [delete]
func HandleDeleteAIModel(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/models/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		db := m.GetGORMDB()
		if err := db.Delete(&models.AIModel{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// --- AI Agents ---

// HandleListAIAgents 获取智能体列表 (精简版)
// @Summary 获取智能体列表
// @Description 获取可用的 AI 智能体列表。非管理员仅能看到公开的和自己创建的。
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "智能体列表"
// @Router /api/ai/agents [get]
func HandleListAIAgents(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[DEBUG] HandleListAIAgents called")

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		isAdmin := false
		if claims != nil {
			userID = uint(claims.UserID)
			isAdmin = claims.IsAdmin
		}

		var agents []models.AIAgent
		// 获取所有字段，包含语音配置等，默认按使用量 (call_count) 降序排列
		db := m.GetGORMDB()
		query := db.Model(&models.AIAgent{}).Order("call_count DESC, created_at DESC")

		// 如果不是管理员，只返回公开的或自己创建的
		if !isAdmin {
			query = query.Where("is_public = ? OR owner_id = ?", true, userID)
		}

		if err := query.Find(&agents).Error; err != nil {
			fmt.Println("[DEBUG] HandleListAIAgents error:", err)
			utils.SendJSONResponse(w, false, "获取智能体失败: "+err.Error(), nil)
			return
		}

		// 打印所有获取到的智能体 ID 和名称，用于调试 ID 4 的归属
		fmt.Printf("[DB DEBUG] Listing all agents (total %d):\n", len(agents))
		for _, a := range agents {
			fmt.Printf("  - ID: %d, Name: '%s', CallCount: %d, IsVoice: %v\n", a.ID, a.Name, a.CallCount, a.IsVoice)
			if a.ID == 4 || strings.Contains(a.Name, "鲁迅") {
				fmt.Printf("    [DETAILED] VoiceID: '%s', VoiceName: '%s', Rate: %f\n", a.VoiceID, a.VoiceName, a.VoiceRate)
			}
		}

		fmt.Printf("[DEBUG] HandleListAIAgents returning %d agents\n", len(agents))
		utils.SendJSONResponse(w, true, "", agents)
	}
}

// HandleGetAIAgent 获取单个智能体详情
// @Summary 获取智能体详情
// @Description 根据 ID 获取 AI 智能体的详细配置，包含关联的模型信息
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "智能体 ID"
// @Success 200 {object} utils.JSONResponse "智能体详情"
// @Router /api/ai/agents/{id} [get]
func HandleGetAIAgent(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 自动从路径末尾获取 ID，兼容 /api/ai/agents/ 和 /api/admin/ai/agents/
		pathParts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
		if len(pathParts) == 0 {
			utils.SendJSONResponse(w, false, "无效的请求路径", nil)
			return
		}
		idStr := pathParts[len(pathParts)-1]
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的 ID: "+idStr, nil)
			return
		}

		var agent models.AIAgent
		db := m.GetGORMDB()
		if err := db.Preload("Model").Preload("Model.Provider").First(&agent, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取智能体详情失败: "+err.Error(), nil)
			return
		}

		// 打印详情获取过程
		if agent.ID == 4 || strings.Contains(agent.Name, "鲁迅") || agent.ID == uint(id) {
			fmt.Printf("[DB DEBUG] Fetching Agent Detail (ID: %d, Name: %s):\n", agent.ID, agent.Name)
			fmt.Printf("  - IsVoice: %v\n", agent.IsVoice)
			fmt.Printf("  - VoiceID: '%s'\n", agent.VoiceID)
			fmt.Printf("  - VoiceName: '%s'\n", agent.VoiceName)
			fmt.Printf("  - VoiceLang: '%s'\n", agent.VoiceLang)
			fmt.Printf("  - VoiceRate: %f\n", agent.VoiceRate)
		}

		// 检查权限：如果是私有的，只有所有者或管理员可以查看
		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		isAdmin := false
		if claims != nil {
			userID = uint(claims.UserID)
			isAdmin = claims.IsAdmin
		}

		if agent.Visibility != "public" && !isAdmin && agent.OwnerID != userID {
			utils.SendJSONResponse(w, false, "该智能体已设为私有，您没有访问权限", nil)
			return
		}

		// 安全处理：隐藏 API Key
		if agent.Model.Provider.APIKey != "" {
			agent.Model.Provider.APIKey = "********"
		}

		utils.SendJSONResponse(w, true, "", agent)
	}
}

// HandleSaveAIAgent 保存智能体
// @Summary 保存智能体
// @Description 新增或更新 AI 智能体配置
// @Tags AI Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.AIAgent true "智能体信息"
// @Success 200 {object} utils.JSONResponse "保存成功"
// @Router /api/admin/ai/agents [post]
func HandleSaveAIAgent(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var agent models.AIAgent
		if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		isAdmin := false
		if claims != nil {
			userID = uint(claims.UserID)
			isAdmin = claims.IsAdmin
		}

		db := m.GetGORMDB()
		var err error
		if agent.ID > 0 {
			// 检查权限：只有所有者或管理员可以修改
			var existing models.AIAgent
			if err := db.First(&existing, agent.ID).Error; err != nil {
				utils.SendJSONResponse(w, false, "未找到智能体", nil)
				return
			}
			if !isAdmin && existing.OwnerID != userID {
				utils.SendJSONResponse(w, false, "您没有修改此智能体的权限", nil)
				return
			}
			err = db.Save(&agent).Error
		} else {
			// 新增时设置 owner_id
			if agent.OwnerID == 0 {
				agent.OwnerID = userID
			}
			err = db.Create(&agent).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存智能体失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", agent)
	}
}

// HandleDeleteAIAgent 删除智能体
// @Summary 删除智能体
// @Description 根据 ID 删除指定的 AI 智能体配置
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "智能体 ID"
// @Success 200 {object} utils.JSONResponse "删除成功"
// @Router /api/admin/ai/agents/{id} [delete]
func HandleDeleteAIAgent(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/agents/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		db := m.GetGORMDB()
		if err := db.Delete(&models.AIAgent{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// HandleListAIUsageLogs 获取 AI 使用日志
// @Summary 获取 AI 使用日志
// @Description 获取最近的 AI 调用日志，用于审计和统计
// @Tags AI Management
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.JSONResponse "日志列表"
// @Router /api/admin/ai/logs [get]
func HandleListAIUsageLogs(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var logs []models.AIUsageLog
		// 默认返回最近 50 条记录
		db := m.GetGORMDB()
		if err := db.Order("id DESC").Limit(50).Find(&logs).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取日志失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", logs)
	}
}

// --- AI Trial (SSE) ---

// HandleAIChatStream 处理网页端流式试用
// @Summary AI 试用对话 (流式)
// @Description 在管理后台直接与指定的 AI 智能体对话，支持 SSE 流式返回
// @Tags AI Trial
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param body body object true "对话请求"
// @Success 200 {string} string "SSE Stream"
// @Router /api/ai/chat/stream [post]
func HandleAIChatStream(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] HandleAIChatStream called, Method: %s, Path: %s\n", r.Method, r.URL.Path)
		if r.Method != http.MethodPost {
			fmt.Printf("[DEBUG] Method not allowed: %s\n", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			AgentID       uint      `json:"agent_id"`
			SessionID     string    `json:"session_id"`
			Message       string    `json:"message"` // 兼容旧版或单条消息
			Messages      []Message `json:"messages"`
			ContextLength int       `json:"context_length"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Printf("[DEBUG] Decode error: %v\n", err)
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		// 如果 Messages 为空但 Message 有值，转换为 Messages
		if len(req.Messages) == 0 && req.Message != "" {
			req.Messages = []Message{
				{Role: RoleUser, Content: req.Message},
			}
		}

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		if claims != nil {
			userID = uint(claims.UserID)
		}

		fmt.Printf("[DEBUG] AgentID: %d, SessionID: %s, UserID: %d, Messages: %d\n", req.AgentID, req.SessionID, userID, len(req.Messages))

		// 获取 Agent 定义
		var agent models.AIAgent
		db := m.GetGORMDB()
		if err := db.First(&agent, req.AgentID).Error; err != nil {
			fmt.Printf("[DEBUG] Agent not found: %d\n", req.AgentID)
			utils.SendJSONResponse(w, false, "未找到智能体", nil)
			return
		}

		// 检查权限
		isAdmin := false
		if claims != nil {
			isAdmin = claims.IsAdmin
		}
		if agent.Visibility != "public" && !isAdmin && agent.OwnerID != userID {
			utils.SendJSONResponse(w, false, "该智能体已设为私有，您没有访问权限", nil)
			return
		}

		// 获取或创建会话
		sessionID := req.SessionID
		var session models.AISession
		if sessionID != "" {
			// 先尝试通过 session_id 查找
			db.Where("session_id = ?", sessionID).First(&session)
		}

		// 如果没找到，或者 sessionID 为空，则创建新会话
		if session.ID == 0 {
			if sessionID == "" {
				sessionID = uuid.New().String()
			}
			session = models.AISession{
				SessionID: sessionID,
				UserID:    userID,
				AgentID:   agent.ID,
				Status:    "active",
				Topic:     "", // 初始主题为空
			}
			// 使用 FirstOrCreate 或直接 Create，并处理可能存在的并发创建导致的唯一索引冲突
			if err := db.Where("session_id = ?", sessionID).FirstOrCreate(&session).Error; err != nil {
				fmt.Printf("[DEBUG] Session creation/fetch failed: %v\n", err)
				// 即使失败，我们也尝试继续，只要 sessionID 不为空
			}
		} else if session.UserID == 0 && userID != 0 {
			// 如果会话已存在但没绑定用户（可能是匿名时创建的），现在绑定上
			db.Model(&session).Update("user_id", userID)
			// 同时更新该会话下所有历史消息的用户ID，确保用户能看到自己的历史记录
			db.Model(&models.AIChatMessage{}).Where("session_id = ?", sessionID).Update("user_id", userID)
		}

		// 保存用户消息
		var lastUserContent string
		if len(req.Messages) > 0 {
			userMsg := req.Messages[len(req.Messages)-1]
			contentStr, _ := userMsg.Content.(string)
			lastUserContent = contentStr
			if err := db.Create(&models.AIChatMessage{
				SessionID: sessionID,
				UserID:    userID,
				Role:      string(RoleUser),
				Content:   contentStr,
			}).Error; err != nil {
				fmt.Printf("[DEBUG] Failed to save user message: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Saved user message for session %s, user %d\n", sessionID, userID)
			}

			// 如果是第一条消息，自动生成主题
			if session.Topic == "" {
				topic := contentStr
				if len(topic) > 50 {
					topic = topic[:47] + "..."
				}
				session.Topic = topic
				db.Model(&session).Update("topic", topic)
			}
		}

		// 获取 Model 和 Provider 详情
		var model models.AIModel
		modelID := agent.ModelID
		if modelID == 0 {
			// 如果智能体没绑定模型，尝试获取系统默认模型
			var defaultModels []models.AIModel
			if err := db.Where("is_default = ?", true).Limit(1).Find(&defaultModels).Error; err == nil && len(defaultModels) > 0 {
				modelID = defaultModels[0].ID
			} else {
				// 如果没有默认模型，尝试获取第一个可用模型
				if err := db.Limit(1).Find(&defaultModels).Error; err == nil && len(defaultModels) > 0 {
					modelID = defaultModels[0].ID
				}
			}
		}

		if modelID > 0 {
			if err := db.Preload("Provider").First(&model, modelID).Error; err == nil {
				fmt.Printf("[DEBUG] Model: %s, Provider: %s, BaseURL: %s\n", model.ModelName, model.Provider.Name, model.Provider.BaseURL)
			} else {
				fmt.Printf("[DEBUG] Model ID %d not found in database: %v\n", modelID, err)
			}
		} else {
			fmt.Printf("[DEBUG] No valid AI model found for agent %d\n", agent.ID)
			utils.SendJSONResponse(w, false, "智能体未配置 AI 模型且系统无可用模型", nil)
			return
		}

		// 增加智能体使用计数 (用于热门排序)
		db.Model(&agent).UpdateColumn("call_count", gorm.Expr("call_count + ?", 1))

		// 准备上下文消息
		fullMessages := []Message{}
		totalTokens := 0
		maxContextTokens := model.ContextSize
		if maxContextTokens <= 0 {
			maxContextTokens = 4096 // 默认值
		}

		// 预留 20% 的空间给 AI 回复，防止溢出
		tokenLimit := int(float64(maxContextTokens) * 0.8)

		// 1. 系统提示词 (最高优先级)
		if agent.SystemPrompt != "" {
			fullMessages = append(fullMessages, Message{
				Role:    RoleSystem,
				Content: agent.SystemPrompt,
			})
			// 粗略估算系统提示词 Token: 字符数 * 1.5
			totalTokens += int(float64(len([]rune(agent.SystemPrompt))) * 1.5)
		}

		// 2. 获取历史消息 (优先从数据库获取以确保准确性)
		limit := req.ContextLength
		if limit <= 0 {
			limit = 20 // 默认 20 条
		}

		var dbMessages []models.AIChatMessage
		db.Where("session_id = ?", sessionID).Order("id desc").Limit(limit).Find(&dbMessages)

		// 转换为 Message 格式并按时间正序排列
		var historyMessages []Message
		for i := len(dbMessages) - 1; i >= 0; i-- {
			historyMessages = append(historyMessages, Message{
				Role:    Role(dbMessages[i].Role),
				Content: dbMessages[i].Content,
			})
		}

		// 如果数据库里还没消息（可能是刚创建的会话），则使用请求中的消息
		if len(historyMessages) == 0 {
			historyMessages = req.Messages
		}

		var selectedMessages []Message
		// 倒序遍历，确保最近的消息被优先包含
		for i := len(historyMessages) - 1; i >= 0; i-- {
			msg := historyMessages[i]
			// 估算当前消息 Token: 字符数 * 1.5 (对中文更友好)
			contentStr, _ := msg.Content.(string)
			msgTokens := int(float64(len([]rune(contentStr))) * 1.5)

			if totalTokens+msgTokens > tokenLimit {
				fmt.Printf("[DEBUG] Context limit reached, discarding older messages. Current tokens: %d, Next msg tokens: %d, Limit: %d\n", totalTokens, msgTokens, tokenLimit)
				break
			}

			selectedMessages = append([]Message{msg}, selectedMessages...)
			totalTokens += msgTokens
		}

		fullMessages = append(fullMessages, selectedMessages...)

		fmt.Printf("[DEBUG] Final context: %d messages, estimated %d tokens (limit: %d)\n", len(fullMessages), totalTokens, tokenLimit)

		aiSvc := m.GetAIService()
		if aiSvc == nil {
			fmt.Println("[DEBUG] aiSvc is nil!")
			utils.SendJSONResponse(w, false, "AI 服务未初始化", nil)
			return
		}

		// 启动流式对话
		stream, err := aiSvc.ChatStream(r.Context(), modelID, fullMessages, nil)
		if err != nil {
			fmt.Printf("[DEBUG] ChatStream error: %v\n", err)
			utils.SendJSONResponse(w, false, "启动流式对话失败: "+err.Error(), nil)
			return
		}

		// 准备流式输出响应
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓存，确保流式立即发送

		flusher, ok := w.(http.Flusher)
		if !ok {
			utils.SendJSONResponse(w, false, "不支持流式响应", nil)
			return
		}

		fmt.Println("[DEBUG] Starting stream loop")
		var assistantContent string

		// 立即发送 session_id 给前端，确保即使 AI 响应延迟，前端也能锁定会话
		metaData, _ := json.Marshal(map[string]string{
			"session_id": sessionID,
		})
		fmt.Fprintf(w, "data: %s\n\n", string(metaData))
		flusher.Flush()

		// 循环发送流式数据
		for resp := range stream {
			if resp.Error != nil {
				errStr := resp.Error.Error()
				// 如果错误信息包含 HTML 标签，说明可能是被防火墙拦截或配置了错误的 URL
				if strings.Contains(errStr, "<html") || strings.Contains(errStr, "<!DOCTYPE") {
					fmt.Printf("[DEBUG] Stream error: Received HTML response instead of JSON. Check your Provider BaseURL.\n")
					fmt.Fprintf(w, "data: %s\n\n", "Error: 接口返回了 HTML 页面而非 JSON，请检查 AI 提供商的 BaseURL 配置是否正确（应为 API 地址而非网页地址）。")
				} else {
					fmt.Printf("[DEBUG] Stream error: %v\n", resp.Error)
					fmt.Fprintf(w, "data: %s\n\n", "Error: "+errStr)
				}
				flusher.Flush()
				break
			}

			if len(resp.Choices) > 0 {
				choice := resp.Choices[0]
				if choice.Delta.Content != "" {
					assistantContent += choice.Delta.Content
					data, _ := json.Marshal(map[string]string{
						"content": choice.Delta.Content,
					})
					fmt.Fprintf(w, "data: %s\n\n", string(data))
					flusher.Flush()
				}
			}
		}

		fmt.Println("[DEBUG] Stream loop finished")

		// 保存 AI 回复到数据库
		if assistantContent != "" {
			db := m.GetGORMDB()
			if err := db.Create(&models.AIChatMessage{
				SessionID: sessionID,
				UserID:    userID,
				Role:      string(RoleAssistant),
				Content:   assistantContent,
			}).Error; err != nil {
				fmt.Printf("[DEBUG] Failed to save assistant message: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Saved assistant message for session %s, user %d\n", sessionID, userID)
			}

			// 更新会话最后一条消息
			db.Model(&models.AISession{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
				"last_msg": assistantContent,
			})

			// 记录使用日志和计算收益扣除
			inputTokens := totalTokens
			outputTokens := int(float64(len(assistantContent)) * 1.2) // 粗略估算

			revenueDeducted := 0
			if agent.RevenueRate > 0 && userID > 0 && userID != agent.OwnerID {
				// 只有非所有者使用时才计算扣除
				// 计算逻辑：(input + output) / 1000 * RevenueRate
				revenueDeducted = int(float64(inputTokens+outputTokens) / 1000.0 * agent.RevenueRate)
				if revenueDeducted < 1 && agent.RevenueRate > 0 {
					revenueDeducted = 1 // 最少扣除 1 算力
				}
			}

			db.Create(&models.AIUsageLog{
				UserID:          userID,
				AgentID:         agent.ID,
				ModelName:       model.ModelName,
				ProviderType:    model.Provider.Type,
				InputTokens:     inputTokens,
				OutputTokens:    outputTokens,
				DurationMS:      0, // 暂时没统计
				Status:          "success",
				RevenueDeducted: revenueDeducted,
				CreatedAt:       time.Now(),
			})

			// TODO: 这里将来可以调用钱包服务扣除用户的算力并给智能体所有者增加收益
			if revenueDeducted > 0 {
				fmt.Printf("[REVENUE] User %d used agent %d, deducted %d points for owner %d\n", userID, agent.ID, revenueDeducted, agent.OwnerID)
			}
		} else if lastUserContent != "" {
			// 如果 AI 没回复，但也更新最后一条消息为用户的
			db := m.GetGORMDB()
			db.Model(&models.AISession{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
				"last_msg": lastUserContent,
			})
		}

		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}
}

// HandleGetAIChatHistory 获取智能体对话历史
// @Summary 获取对话历史
// @Description 获取指定会话的聊天历史记录，支持分页
// @Tags AI Trial
// @Produce json
// @Security BearerAuth
// @Param session_id query string true "会话 ID"
// @Param limit query int false "获取条数，默认 20"
// @Param before_id query int false "获取该 ID 之前的消息"
// @Success 200 {array} models.AIChatMessage "消息列表"
// @Router /api/ai/chat/history [get]
func HandleGetAIChatHistory(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			// 兼容旧版，如果没传 session_id，则尝试用 agent_id 生成默认 session_id
			agentIDStr := r.URL.Query().Get("agent_id")
			agentID, _ := strconv.Atoi(agentIDStr)
			if agentID > 0 {
				sessionID = fmt.Sprintf("agent:%d", agentID)
			}
		}

		if sessionID == "" {
			utils.SendJSONResponse(w, false, "无效的会话ID", nil)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit, _ := strconv.Atoi(limitStr)
		if limit <= 0 {
			limit = 20
		}

		beforeIDStr := r.URL.Query().Get("before_id")
		beforeID, _ := strconv.Atoi(beforeIDStr)

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		isAdmin := false
		if claims != nil {
			userID = uint(claims.UserID)
			isAdmin = claims.IsAdmin
		}

		var messages []models.AIChatMessage
		db := m.GetGORMDB()
		query := db.Where("session_id = ?", sessionID).Order("id desc").Limit(limit)

		// 检查会话归属
		if !isAdmin && !strings.HasPrefix(sessionID, "agent:") {
			var session models.AISession
			if err := db.Where("session_id = ?", sessionID).First(&session).Error; err == nil {
				if session.UserID != 0 && session.UserID != userID {
					utils.SendJSONResponse(w, false, "无权访问该会话", nil)
					return
				}
			}
		}

		if beforeID > 0 {
			query = query.Where("id < ?", beforeID)
		}

		if err := query.Find(&messages).Error; err != nil {
			fmt.Printf("[DEBUG] Get history error: %v\n", err)
			utils.SendJSONResponse(w, false, "获取历史记录失败: "+err.Error(), nil)
			return
		}

		fmt.Printf("[DEBUG] Retrieved %d messages for session %s, user %d\n", len(messages), sessionID, userID)

		// 翻转数组，使其按时间正序排列返回给前端
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}

		utils.SendJSONResponse(w, true, "", messages)
	}
}

// HandleGetRecentSessions 获取用户最近的 AI 对话会话
// @Summary 获取最近会话
// @Description 获取当前登录用户最近的 AI 对话会话列表
// @Tags AI Trial
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.AISession "会话列表"
// @Router /api/ai/chat/sessions [get]
func HandleGetRecentSessions(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		if claims != nil {
			userID = uint(claims.UserID)
		}

		if userID == 0 {
			utils.SendJSONResponse(w, false, "未登录", nil)
			return
		}

		var sessions []models.AISession
		// 获取最近的 50 个会话，并关联智能体信息
		db := m.GetGORMDB()
		err := db.Preload("Agent").Where("user_id = ?", userID).Order("updated_at desc").Limit(50).Find(&sessions).Error
		if err != nil {
			utils.SendJSONResponse(w, false, "获取会话列表失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", sessions)
	}
}

// --- Cognitive Memory Autonomous Learning ---

// HandleBotLearnURL 从 URL 自动学习并存入记忆
// @Summary 从 URL 自动学习
// @Description 提供一个 URL，让数字员工抓取并分析其中的知识，存入其认知记忆
// @Tags AI Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "学习请求 (bot_id, url, category)"
// @Success 200 {object} utils.JSONResponse "学习任务已提交"
// @Router /api/ai/memory/learn/url [post]
func HandleBotLearnURL(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BotID    string `json:"bot_id"`
			URL      string `json:"url"`
			Category string `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		if req.BotID == "" || req.URL == "" {
			utils.SendJSONResponse(w, false, "bot_id 和 url 不能为空", nil)
			return
		}

		memorySvc := m.GetCognitiveMemoryService()
		if memorySvc == nil {
			utils.SendJSONResponse(w, false, "认知记忆服务未初始化", nil)
			return
		}

		// 异步执行，防止阻塞前端
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			err := memorySvc.LearnFromURL(ctx, req.BotID, req.URL, req.Category)
			if err != nil {
				fmt.Printf("[Memory] LearnFromURL failed: %v\n", err)
			}
		}()

		utils.SendJSONResponse(w, true, "抓取学习任务已异步提交，请稍后在记忆库中查看结果", nil)
	}
}

// HandleBotLearnFile 从上传的文件学习并存入记忆
// @Summary 从文件学习
// @Description 上传 PDF/Excel/Docx 等文件，让数字员工分析其中的知识，存入其认知记忆
// @Tags AI Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param bot_id formData string true "机器人 ID"
// @Param category formData string false "知识分类"
// @Param file formData file true "知识文件"
// @Success 200 {object} utils.JSONResponse "学习任务已提交"
// @Router /api/ai/memory/learn/file [post]
func HandleBotLearnFile(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 multipart/form-data
		err := r.ParseMultipartForm(32 << 20) // 32MB max
		if err != nil {
			utils.SendJSONResponse(w, false, "解析表单失败: "+err.Error(), nil)
			return
		}

		botID := r.FormValue("bot_id")
		category := r.FormValue("category")
		if botID == "" {
			utils.SendJSONResponse(w, false, "bot_id 不能为空", nil)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			utils.SendJSONResponse(w, false, "读取文件失败: "+err.Error(), nil)
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			utils.SendJSONResponse(w, false, "读取文件内容失败: "+err.Error(), nil)
			return
		}

		memorySvc := m.GetCognitiveMemoryService()
		if memorySvc == nil {
			utils.SendJSONResponse(w, false, "认知记忆服务未初始化", nil)
			return
		}

		// 异步执行
		filename := header.Filename
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			err := memorySvc.LearnFromContent(ctx, botID, content, filename, category)
			if err != nil {
				fmt.Printf("[Memory] LearnFromContent failed: %v\n", err)
			}
		}()

		utils.SendJSONResponse(w, true, "文件分析学习任务已异步提交", nil)
	}
}

// HandleBotConsolidateMemories 手动触发记忆固化
// @Summary 触发记忆固化
// @Description 手动触发对指定机器人/用户的记忆进行总结与去重
// @Tags AI Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "固化请求 (bot_id, user_id)"
// @Success 200 {object} utils.JSONResponse "固化任务已提交"
// @Router /api/ai/memory/consolidate [post]
func HandleBotConsolidateMemories(m Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BotID  string `json:"bot_id"`
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		memorySvc := m.GetCognitiveMemoryService()
		if memorySvc == nil {
			utils.SendJSONResponse(w, false, "认知记忆服务未初始化", nil)
			return
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			err := memorySvc.ConsolidateMemories(ctx, req.UserID, req.BotID, m.GetAIService())
			if err != nil {
				fmt.Printf("[Memory] ConsolidateMemories failed: %v\n", err)
			}
		}()

		utils.SendJSONResponse(w, true, "记忆固化任务已异步提交", nil)
	}
}
