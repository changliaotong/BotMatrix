package app

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- AI Providers ---

// HandleListAIProviders 获取提供商列表
func HandleListAIProviders(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var providers []models.AIProviderGORM
		if err := m.GORMDB.Find(&providers).Error; err != nil {
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
func HandleSaveAIProvider(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var provider models.AIProviderGORM
		if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		var err error
		if provider.ID > 0 {
			// 如果是更新，且 APIKey 是 "********"，说明用户没有修改 Key，需要保留原有的 Key
			if provider.APIKey == "********" {
				var oldProvider models.AIProviderGORM
				if err := m.GORMDB.First(&oldProvider, provider.ID).Error; err == nil {
					provider.APIKey = oldProvider.APIKey
				}
			}
			fmt.Printf("[DEBUG] Saving AI Provider %d (%s), Key length: %d\n", provider.ID, provider.Name, len(provider.APIKey))
			err = m.GORMDB.Save(&provider).Error
		} else {
			fmt.Printf("[DEBUG] Creating AI Provider (%s), Key length: %d\n", provider.Name, len(provider.APIKey))
			err = m.GORMDB.Create(&provider).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存提供商失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", provider)
	}
}

// HandleDeleteAIProvider 删除提供商
func HandleDeleteAIProvider(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/providers/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		if err := m.GORMDB.Delete(&models.AIProviderGORM{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// --- AI Models ---

// HandleListAIModels 获取模型列表
func HandleListAIModels(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var modelsList []models.AIModelGORM
		if err := m.GORMDB.Preload("Provider").Find(&modelsList).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取模型失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", modelsList)
	}
}

// HandleSaveAIModel 保存模型
func HandleSaveAIModel(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model models.AIModelGORM
		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		var err error
		if model.ID > 0 {
			err = m.GORMDB.Save(&model).Error
		} else {
			err = m.GORMDB.Create(&model).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存模型失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", model)
	}
}

// HandleDeleteAIModel 删除模型
func HandleDeleteAIModel(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/models/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		if err := m.GORMDB.Delete(&models.AIModelGORM{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// --- AI Agents ---

// HandleListAIAgents 获取智能体列表 (精简版)
func HandleListAIAgents(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[DEBUG] HandleListAIAgents called")

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		isAdmin := false
		if claims != nil {
			userID = uint(claims.UserID)
			isAdmin = claims.IsAdmin
		}

		var agents []models.AIAgentGORM
		// 获取所有字段，包含语音配置等，默认按使用量 (call_count) 降序排列
		query := m.GORMDB.Model(&models.AIAgentGORM{}).Order("call_count DESC, created_at DESC")

		// 如果不是管理员，只返回公开的或自己创建的
		if !isAdmin {
			query = query.Where("visibility = ? OR owner_id = ?", "public", userID)
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
func HandleGetAIAgent(m *Manager) http.HandlerFunc {
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

		var agent models.AIAgentGORM
		if err := m.GORMDB.Preload("Model").Preload("Model.Provider").First(&agent, id).Error; err != nil {
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

		if agent.Visibility == "private" && !isAdmin && agent.OwnerID != userID {
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
func HandleSaveAIAgent(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var agent models.AIAgentGORM
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

		var err error
		if agent.ID > 0 {
			// 检查权限：只有所有者或管理员可以修改
			var existing models.AIAgentGORM
			if err := m.GORMDB.First(&existing, agent.ID).Error; err != nil {
				utils.SendJSONResponse(w, false, "未找到智能体", nil)
				return
			}
			if !isAdmin && existing.OwnerID != userID {
				utils.SendJSONResponse(w, false, "没有权限修改此智能体", nil)
				return
			}
			// 保持 OwnerID 不变，除非是管理员显式修改（目前前端没这功能）
			agent.OwnerID = existing.OwnerID
			err = m.GORMDB.Save(&agent).Error
		} else {
			// 新建时设置所有者
			agent.OwnerID = userID
			err = m.GORMDB.Create(&agent).Error
		}

		if err != nil {
			utils.SendJSONResponse(w, false, "保存智能体失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "保存成功", agent)
	}
}

// HandleDeleteAIAgent 删除智能体
func HandleDeleteAIAgent(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/ai/agents/")
		id, _ := strconv.Atoi(idStr)
		if id == 0 {
			utils.SendJSONResponse(w, false, "无效的ID", nil)
			return
		}

		if err := m.GORMDB.Delete(&models.AIAgentGORM{}, id).Error; err != nil {
			utils.SendJSONResponse(w, false, "删除失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "删除成功", nil)
	}
}

// HandleListAIUsageLogs 获取 AI 使用日志
func HandleListAIUsageLogs(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var logs []models.AIUsageLogGORM
		// 默认返回最近 50 条记录
		if err := m.GORMDB.Order("id DESC").Limit(50).Find(&logs).Error; err != nil {
			utils.SendJSONResponse(w, false, "获取日志失败: "+err.Error(), nil)
			return
		}
		utils.SendJSONResponse(w, true, "", logs)
	}
}

// --- AI Trial (SSE) ---

// HandleAIChatStream 处理网页端流式试用
func HandleAIChatStream(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] HandleAIChatStream called, Method: %s, Path: %s\n", r.Method, r.URL.Path)
		if r.Method != http.MethodPost {
			fmt.Printf("[DEBUG] Method not allowed: %s\n", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			AgentID       uint         `json:"agent_id"`
			SessionID     string       `json:"session_id"`
			Message       string       `json:"message"` // 兼容旧版或单条消息
			Messages      []ai.Message `json:"messages"`
			ContextLength int          `json:"context_length"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Printf("[DEBUG] Decode error: %v\n", err)
			utils.SendJSONResponse(w, false, "请求格式错误", nil)
			return
		}

		// 如果 Messages 为空但 Message 有值，转换为 Messages
		if len(req.Messages) == 0 && req.Message != "" {
			req.Messages = []ai.Message{
				{Role: ai.RoleUser, Content: req.Message},
			}
		}

		claims, _ := r.Context().Value(types.UserClaimsKey).(*types.UserClaims)
		userID := uint(0)
		if claims != nil {
			userID = uint(claims.UserID)
		}

		fmt.Printf("[DEBUG] AgentID: %d, SessionID: %s, UserID: %d, Messages: %d\n", req.AgentID, req.SessionID, userID, len(req.Messages))

		// 获取 Agent 定义
		var agent models.AIAgentGORM
		if err := m.GORMDB.First(&agent, req.AgentID).Error; err != nil {
			fmt.Printf("[DEBUG] Agent not found: %d\n", req.AgentID)
			utils.SendJSONResponse(w, false, "未找到智能体", nil)
			return
		}

		// 检查权限
		isAdmin := false
		if claims != nil {
			isAdmin = claims.IsAdmin
		}
		if agent.Visibility == "private" && !isAdmin && agent.OwnerID != userID {
			utils.SendJSONResponse(w, false, "该智能体已设为私有，您没有访问权限", nil)
			return
		}

		// 获取或创建会话
		sessionID := req.SessionID
		var session models.AISessionGORM
		if sessionID != "" {
			// 先尝试通过 session_id 查找
			m.GORMDB.Where("session_id = ?", sessionID).First(&session)
		}

		// 如果没找到，或者 sessionID 为空，则创建新会话
		if session.ID == 0 {
			if sessionID == "" {
				sessionID = uuid.New().String()
			}
			session = models.AISessionGORM{
				SessionID: sessionID,
				UserID:    userID,
				AgentID:   agent.ID,
				Status:    "active",
				Topic:     "", // 初始主题为空
			}
			// 使用 FirstOrCreate 或直接 Create，并处理可能存在的并发创建导致的唯一索引冲突
			if err := m.GORMDB.Where("session_id = ?", sessionID).FirstOrCreate(&session).Error; err != nil {
				fmt.Printf("[DEBUG] Session creation/fetch failed: %v\n", err)
				// 即使失败，我们也尝试继续，只要 sessionID 不为空
			}
		} else if session.UserID == 0 && userID != 0 {
			// 如果会话已存在但没绑定用户（可能是匿名时创建的），现在绑定上
			m.GORMDB.Model(&session).Update("user_id", userID)
			// 同时更新该会话下所有历史消息的用户ID，确保用户能看到自己的历史记录
			m.GORMDB.Model(&models.AIChatMessageGORM{}).Where("session_id = ?", sessionID).Update("user_id", userID)
		}

		// 保存用户消息
		var lastUserContent string
		if len(req.Messages) > 0 {
			userMsg := req.Messages[len(req.Messages)-1]
			contentStr, _ := userMsg.Content.(string)
			lastUserContent = contentStr
			if err := m.GORMDB.Create(&models.AIChatMessageGORM{
				SessionID: sessionID,
				UserID:    userID,
				Role:      string(ai.RoleUser),
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
				m.GORMDB.Model(&session).Update("topic", topic)
			}
		}

		// 获取 Model 和 Provider 详情
		var model models.AIModelGORM
		modelID := agent.ModelID
		if modelID == 0 {
			// 如果智能体没绑定模型，尝试获取系统默认模型
			var defaultModels []models.AIModelGORM
			if err := m.GORMDB.Where("is_default = ?", true).Limit(1).Find(&defaultModels).Error; err == nil && len(defaultModels) > 0 {
				modelID = defaultModels[0].ID
			} else {
				// 如果没有默认模型，尝试获取第一个可用模型
				if err := m.GORMDB.Limit(1).Find(&defaultModels).Error; err == nil && len(defaultModels) > 0 {
					modelID = defaultModels[0].ID
				}
			}
		}

		if modelID > 0 {
			if err := m.GORMDB.Preload("Provider").First(&model, modelID).Error; err == nil {
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
		m.GORMDB.Model(&agent).UpdateColumn("call_count", gorm.Expr("call_count + ?", 1))

		// 准备上下文消息
		fullMessages := []ai.Message{}
		totalTokens := 0
		maxContextTokens := model.ContextSize
		if maxContextTokens <= 0 {
			maxContextTokens = 4096 // 默认值
		}

		// 预留 20% 的空间给 AI 回复，防止溢出
		tokenLimit := int(float64(maxContextTokens) * 0.8)

		// 1. 系统提示词 (最高优先级)
		if agent.SystemPrompt != "" {
			fullMessages = append(fullMessages, ai.Message{
				Role:    ai.RoleSystem,
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

		var dbMessages []models.AIChatMessageGORM
		m.GORMDB.Where("session_id = ?", sessionID).Order("id desc").Limit(limit).Find(&dbMessages)

		// 转换为 ai.Message 格式并按时间正序排列
		var historyMessages []ai.Message
		for i := len(dbMessages) - 1; i >= 0; i-- {
			historyMessages = append(historyMessages, ai.Message{
				Role:    ai.Role(dbMessages[i].Role),
				Content: dbMessages[i].Content,
			})
		}

		// 如果数据库里还没消息（可能是刚创建的会话），则使用请求中的消息
		if len(historyMessages) == 0 {
			historyMessages = req.Messages
		}

		var selectedMessages []ai.Message
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

			selectedMessages = append([]ai.Message{msg}, selectedMessages...)
			totalTokens += msgTokens
		}

		fullMessages = append(fullMessages, selectedMessages...)

		fmt.Printf("[DEBUG] Final context: %d messages, estimated %d tokens (limit: %d)\n", len(fullMessages), totalTokens, tokenLimit)

		if m.AIIntegrationService == nil {
			fmt.Println("[DEBUG] m.AIIntegrationService is nil!")
			utils.SendJSONResponse(w, false, "AI 服务未初始化", nil)
			return
		}

		// 启动流式对话
		stream, err := m.AIIntegrationService.ChatStream(r.Context(), modelID, fullMessages, nil)
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
			if err := m.GORMDB.Create(&models.AIChatMessageGORM{
				SessionID: sessionID,
				UserID:    userID,
				Role:      string(ai.RoleAssistant),
				Content:   assistantContent,
			}).Error; err != nil {
				fmt.Printf("[DEBUG] Failed to save assistant message: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Saved assistant message for session %s, user %d\n", sessionID, userID)
			}

			// 更新会话最后一条消息
			m.GORMDB.Model(&models.AISessionGORM{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
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

			m.GORMDB.Create(&models.AIUsageLogGORM{
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
			m.GORMDB.Model(&models.AISessionGORM{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
				"last_msg": lastUserContent,
			})
		}

		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}
}

// HandleGetAIChatHistory 获取智能体对话历史
func HandleGetAIChatHistory(m *Manager) http.HandlerFunc {
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

		var messages []models.AIChatMessageGORM
		query := m.GORMDB.Where("session_id = ?", sessionID).Order("id desc").Limit(limit)

		// 检查会话归属
		if !isAdmin && !strings.HasPrefix(sessionID, "agent:") {
			var session models.AISessionGORM
			if err := m.GORMDB.Where("session_id = ?", sessionID).First(&session).Error; err == nil {
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
func HandleGetRecentSessions(m *Manager) http.HandlerFunc {
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

		var sessions []models.AISessionGORM
		// 获取最近的 50 个会话，并关联智能体信息
		err := m.GORMDB.Preload("Agent").Where("user_id = ?", userID).Order("updated_at desc").Limit(50).Find(&sessions).Error
		if err != nil {
			utils.SendJSONResponse(w, false, "获取会话列表失败: "+err.Error(), nil)
			return
		}

		utils.SendJSONResponse(w, true, "", sessions)
	}
}
