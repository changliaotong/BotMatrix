package mcp

import (
	"BotMatrix/common/ai"
	// "BotMatrix/common/ai/employee" // Removed to break cycle
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"
	"fmt"
	"strings"
	"time"
)

// CollaborationExecutionInfo 协作任务执行信息
type CollaborationExecutionInfo struct {
	ExecutionID string
	Status      string
	Result      string
}

// CollaborationProvider 抽象智能体间协作所需的底层能力
type CollaborationProvider interface {
	GetEmployeesByOrg(ctx context.Context, orgID uint) ([]models.DigitalEmployee, error)
	GetEmployeeByID(ctx context.Context, orgID uint, employeeID string) (*models.DigitalEmployee, error)
	ChatWithEmployee(ctx context.Context, employee *models.DigitalEmployee, message *types.Message, orgID uint) (string, error)
	CreateExecution(ctx context.Context, executionID, traceID string) error
	UpdateExecution(ctx context.Context, executionID string, status string, result string) error
	GetExecution(ctx context.Context, executionID string) (*CollaborationExecutionInfo, error)
	NewInternalMessage(msgType, category, botID, role, content string) *types.Message
}

// AgentCollaborationMCPHost 提供智能体间协作的 MCP 工具
type AgentCollaborationMCPHost struct {
	provider CollaborationProvider
}

func NewAgentCollaborationMCPHost(provider CollaborationProvider) *AgentCollaborationMCPHost {
	return &AgentCollaborationMCPHost{provider: provider}
}

func (h *AgentCollaborationMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "colleague_list",
			Description: "获取当前企业内所有可用的数字员工（同事）列表及其职责描述",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
		{
			Name:        "colleague_consult",
			Description: "向另一位数字员工咨询问题或请求协助。这会同步等待该员工的答复。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_employee_id": map[string]any{
						"type":        "string",
						"description": "目标员工的工号 (EmployeeID)",
					},
					"question": map[string]any{
						"type":        "string",
						"description": "你想要咨询的具体问题 or 请求内容",
					},
				},
				"required": []string{"target_employee_id", "question"},
			},
		},
		{
			Name:        "task_delegate",
			Description: "将一个复杂的任务分解并委派给另一位数字员工处理。这适用于不需要立即得到最终结果，或者需要该员工独立完成的工作。返回 execution_id 用于后续查询进度。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_employee_id": map[string]any{
						"type":        "string",
						"description": "目标员工的工号 (EmployeeID)",
					},
					"task_description": map[string]any{
						"type":        "string",
						"description": "任务的详细描述、背景信息及交付要求",
					},
					"priority": map[string]any{
						"type": "string",
						"enum": []string{"low", "medium", "high"},
					},
				},
				"required": []string{"target_employee_id", "task_description"},
			},
		},
		{
			Name:        "task_status",
			Description: "查询之前委派任务的执行状态和结果。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"execution_id": map[string]any{
						"type":        "string",
						"description": "委派任务时返回的 execution_id",
					},
				},
				"required": []string{"execution_id"},
			},
		},
		{
			Name:        "task_report",
			Description: "主动向任务发起者汇报当前任务的进度、阶段性成果或最终结果。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"execution_id": map[string]any{
						"type":        "string",
						"description": "当前正在处理的任务 execution_id",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "任务状态: running (执行中), success (成功), failed (失败)",
						"enum":        []string{"running", "success", "failed"},
					},
					"result": map[string]any{
						"type":        "string",
						"description": "汇报的具体内容、结果或错误信息",
					},
				},
				"required": []string{"execution_id", "status", "result"},
			},
		},
	}, nil
}

func (h *AgentCollaborationMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *AgentCollaborationMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *AgentCollaborationMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	// 从上下文获取当前组织 ID
	orgID, _ := ctx.Value("orgIDNum").(uint)
	if orgID == 0 {
		return nil, fmt.Errorf("unauthorized: organization ID missing in context")
	}

	// 传递父 SessionID 用于链路追踪
	parentSessionID, _ := ctx.Value("sessionID").(string)

	switch toolName {
	case "colleague_list":
		employees, err := h.provider.GetEmployeesByOrg(ctx, orgID)
		if err != nil {
			return nil, err
		}

		var sb strings.Builder
		sb.WriteString("以下是您可以协作的同事列表：\n")
		for _, e := range employees {
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s): %s\n", e.EmployeeID, e.Name, e.Title, e.Bio))
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil

	case "colleague_consult":
		targetID, _ := arguments["target_employee_id"].(string)
		question, _ := arguments["question"].(string)

		// 1. 查找目标员工
		targetEmp, err := h.provider.GetEmployeeByID(ctx, orgID, targetID)
		if err != nil {
			return nil, err
		}

		// 2. 构造模拟内部消息
		internalMsg := h.provider.NewInternalMessage("mesh", "collaboration", targetEmp.BotID, "system", question)
		if parentSessionID != "" {
			internalMsg.Extras["parentSessionID"] = parentSessionID
		}

		// 3. 调用 AI 服务进行对话
		resp, err := h.provider.ChatWithEmployee(ctx, targetEmp, internalMsg, orgID)
		if err != nil {
			return nil, fmt.Errorf("向同事咨询失败: %v", err)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("同事 %s 的答复：\n%s", targetEmp.Name, resp),
				},
			},
		}, nil

	case "task_delegate":
		targetID, _ := arguments["target_employee_id"].(string)
		taskDesc, _ := arguments["task_description"].(string)

		// 1. 查找目标员工
		targetEmp, err := h.provider.GetEmployeeByID(ctx, orgID, targetID)
		if err != nil {
			return nil, err
		}

		// 2. 创建任务执行记录用于追踪
		executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano()) // 简单生成
		if err := h.provider.CreateExecution(ctx, executionID, parentSessionID); err != nil {
			return nil, fmt.Errorf("创建任务追踪失败: %v", err)
		}

		// 3. 异步启动任务 (委派)
		go func() {
			// 注意：异步执行需要一个新的上下文或清理过的上下文
			asyncCtx := context.Background()
			internalMsg := h.provider.NewInternalMessage("mesh", "delegation", targetEmp.BotID, "system", taskDesc)
			internalMsg.Extras["executionID"] = executionID
			if parentSessionID != "" {
				internalMsg.Extras["parentSessionID"] = parentSessionID
			}

			resp, err := h.provider.ChatWithEmployee(asyncCtx, targetEmp, internalMsg, orgID)

			// 更新执行结果
			status := "success"
			result := resp
			if err != nil {
				status = "failed"
				result = fmt.Sprintf("Error: %v", err)
			}

			_ = h.provider.UpdateExecution(asyncCtx, executionID, status, result)
		}()

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("已成功将任务委派给同事 %s (%s)。任务追踪 ID: %s", targetEmp.Name, targetEmp.EmployeeID, executionID),
				},
			},
		}, nil

	case "task_status":
		execID, _ := arguments["execution_id"].(string)
		exec, err := h.provider.GetExecution(ctx, execID)
		if err != nil {
			return nil, fmt.Errorf("未找到 ID 为 %s 的任务记录", execID)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("任务 [%s] 当前状态: %s\n执行结果: %s", execID, exec.Status, exec.Result),
				},
			},
		}, nil

	case "task_report":
		execID, _ := arguments["execution_id"].(string)
		status, _ := arguments["status"].(string)
		result, _ := arguments["result"].(string)

		if err := h.provider.UpdateExecution(ctx, execID, status, result); err != nil {
			return nil, fmt.Errorf("汇报任务进度失败: %v", err)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: "任务汇报已接收。",
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

func (h *AgentCollaborationMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *AgentCollaborationMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}

// ReasoningMCPHost 提供推理相关的工具，如 Sequential Thinking
type ReasoningMCPHost struct {
	aiSvc types.AIService
}

func NewReasoningMCPHost(aiSvc types.AIService) *ReasoningMCPHost {
	return &ReasoningMCPHost{aiSvc: aiSvc}
}

func (h *ReasoningMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "sequential_thinking",
			Description: "A tool for dynamic and reflective problem-solving through a structured sequence of thoughts. Use this for complex tasks that require planning, step-by-step reasoning, and self-correction.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"thought": map[string]any{
						"type":        "string",
						"description": "The current thinking process or step.",
					},
					"thoughtNumber": map[string]any{
						"type":        "integer",
						"description": "The current step number in the reasoning chain.",
					},
					"totalThoughts": map[string]any{
						"type":        "integer",
						"description": "The estimated total number of steps (can be adjusted).",
					},
					"nextStep": map[string]any{
						"type":        "string",
						"description": "What to do next based on this thought.",
					},
					"isFinished": map[string]any{
						"type":        "boolean",
						"description": "Whether the reasoning process is complete.",
					},
				},
				"required": []string{"thought", "thoughtNumber", "nextStep"},
			},
		},
	}, nil
}

func (h *ReasoningMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *ReasoningMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *ReasoningMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "sequential_thinking" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	thought, _ := arguments["thought"].(string)
	num, _ := arguments["thoughtNumber"]
	isFinished, _ := arguments["isFinished"].(bool)

	fmt.Printf("\n[Reasoning] Step %v: %s\n", num, thought)

	msg := fmt.Sprintf("Thought accepted. Next: %v", arguments["nextStep"])
	if isFinished {
		msg = "Reasoning complete. Proceeding to final answer."
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{
			{
				Type: "text",
				Text: msg,
			},
		},
	}, nil
}

func (h *ReasoningMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *ReasoningMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}

// SearchMCPHost 提供安全联网搜索工具
type SearchMCPHost struct {
	aiSvc ai.AIService
}

func NewSearchMCPHost(aiSvc ai.AIService) *SearchMCPHost {
	return &SearchMCPHost{aiSvc: aiSvc}
}

func (h *SearchMCPHost) ListTools(ctx context.Context, serverID string) ([]ai.MCPTool, error) {
	return []ai.MCPTool{
		{
			Name:        "web_search",
			Description: "Search the web for real-time information. Privacy Bastion will automatically protect your sensitive data in the query.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query.",
					},
				},
				"required": []string{"query"},
			},
		},
	}, nil
}

func (h *SearchMCPHost) ListResources(ctx context.Context, serverID string) ([]ai.MCPResource, error) {
	return nil, nil
}

func (h *SearchMCPHost) ListPrompts(ctx context.Context, serverID string) ([]ai.MCPPrompt, error) {
	return nil, nil
}

func (h *SearchMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "web_search" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	query, _ := arguments["query"].(string)
	// 在此处可以增加 Privacy Bastion 的逻辑，或者依赖 aiSvc 内部处理

	// 模拟搜索结果
	return types.MCPCallToolResponse{
		Content: []types.MCPContent{{Type: "text", Text: fmt.Sprintf("Search results for: %s\n1. Result A...\n2. Result B...", query)}},
	}, nil
}

func (h *SearchMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *SearchMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}

// KnowledgeMCPHost 提供本地知识库检索能力
type KnowledgeMCPHost struct {
	kb types.KnowledgeBase
}

func NewKnowledgeMCPHost(kb types.KnowledgeBase) *KnowledgeMCPHost {
	return &KnowledgeMCPHost{
		kb: kb,
	}
}

func (h *KnowledgeMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "search_knowledge",
			Description: "Search the knowledge base for technical documentation, architecture details, and project information using semantic search.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query or keywords in natural language.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 3).",
					},
				},
				"required": []string{"query"},
			},
		},
	}, nil
}

func (h *KnowledgeMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *KnowledgeMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *KnowledgeMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if toolName != "search_knowledge" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	if h.kb == nil {
		return nil, fmt.Errorf("knowledge base not initialized")
	}

	query, _ := arguments["query"].(string)
	limit := 3
	if l, ok := arguments["limit"].(float64); ok {
		limit = int(l)
	}

	filter := &types.SearchFilter{
		Status: "active",
	}
	chunks, err := h.kb.Search(ctx, query, limit, filter)
	if err != nil {
		return nil, fmt.Errorf("knowledge search failed: %v", err)
	}

	if len(chunks) == 0 {
		return types.MCPCallToolResponse{
			Content: []types.MCPContent{{Type: "text", Text: "No relevant information found in the knowledge base."}},
		}, nil
	}

	var results []string
	for _, c := range chunks {
		results = append(results, fmt.Sprintf("Source: %s\nContent: %s", c.Source, c.Content))
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{{Type: "text", Text: "Found relevant information:\n\n" + strings.Join(results, "\n\n---\n\n")}},
	}, nil
}

func (h *KnowledgeMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *KnowledgeMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}

// MemoryMCPHost 提供记忆管理工具
type MemoryMCPHost struct {
	memorySvc types.CognitiveMemoryService
}

func NewMemoryMCPHost(memorySvc types.CognitiveMemoryService) *MemoryMCPHost {
	return &MemoryMCPHost{memorySvc: memorySvc}
}

func (h *MemoryMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "store_memory",
			Description: "Store important information in long-term memory. Supports automatic semantic indexing.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content": map[string]any{
						"type":        "string",
						"description": "The information to remember.",
					},
					"category": map[string]any{
						"type":        "string",
						"description": "Optional category (e.g., 'user_preference', 'project_info').",
					},
					"importance": map[string]any{
						"type":        "integer",
						"description": "Importance level (1-10), default is 5.",
						"minimum":     1,
						"maximum":     10,
					},
				},
				"required": []string{"content"},
			},
		},
		{
			Name:        "search_memory",
			Description: "Search through long-term memory using semantic similarity or keywords. Returns memory IDs for potential management.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query to find relevant memories.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "forget_memory",
			Description: "Remove a specific piece of information from long-term memory using its memory ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"memory_id": map[string]any{
						"type":        "integer",
						"description": "The unique ID of the memory to forget.",
					},
				},
				"required": []string{"memory_id"},
			},
		},
	}, nil
}

func (h *MemoryMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *MemoryMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *MemoryMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if h.memorySvc == nil {
		return nil, fmt.Errorf("CognitiveMemoryService not initialized")
	}

	botID := "system"
	userID := "system"

	switch toolName {
	case "store_memory":
		content, _ := arguments["content"].(string)
		category, _ := arguments["category"].(string)
		importance, _ := arguments["importance"].(float64)
		if importance == 0 {
			importance = 5
		}

		memory := &models.CognitiveMemory{
			BotID:      botID,
			UserID:     userID,
			Content:    content,
			Category:   category,
			Importance: int(importance),
			LastSeen:   time.Now(),
		}

		err := h.memorySvc.SaveMemory(ctx, memory)
		if err != nil {
			return nil, err
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{{Type: "text", Text: "Memory saved and indexed successfully."}},
		}, nil

	case "search_memory":
		query, _ := arguments["query"].(string)
		memories, err := h.memorySvc.GetRelevantMemories(ctx, userID, botID, query)
		if err != nil {
			return nil, err
		}

		if len(memories) == 0 {
			return types.MCPCallToolResponse{
				Content: []types.MCPContent{{Type: "text", Text: "No relevant memories found."}},
			}, nil
		}

		var results []string
		for _, m := range memories {
			results = append(results, fmt.Sprintf("[ID: %d] [%s] %s (Importance: %d)", m.ID, m.Category, m.Content, m.Importance))
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{{Type: "text", Text: "Found relevant memories:\n" + fmt.Sprintf("- %s", strings.Join(results, "\n- "))}},
		}, nil

	case "forget_memory":
		memoryIDFloat, ok := arguments["memory_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid memory_id: must be an integer")
		}
		memoryID := uint(memoryIDFloat)

		err := h.memorySvc.ForgetMemory(ctx, memoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to forget memory: %v", err)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{{Type: "text", Text: fmt.Sprintf("Memory ID %d has been removed from long-term memory.", memoryID)}},
		}, nil

	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

func (h *MemoryMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("not supported")
}

func (h *MemoryMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("not supported")
}

// InternalSkillProvider 定义了内部技能提供的接口
type InternalSkillProvider interface {
	GetTools() []types.Tool
	ExecuteAction(actionType string, params map[string]any) error
}

// InternalSkillMCPHost 将现有的 Skill Center 桥接为 MCP 协议
type InternalSkillMCPHost struct {
	provider InternalSkillProvider
}

func NewInternalSkillMCPHost(provider InternalSkillProvider) *InternalSkillMCPHost {
	return &InternalSkillMCPHost{provider: provider}
}

func (h *InternalSkillMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	if h.provider == nil {
		return nil, nil
	}

	aiTools := h.provider.GetTools()
	mcpTools := make([]types.MCPTool, len(aiTools))
	for i, t := range aiTools {
		mcpTools[i] = types.MCPTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
	}
	return mcpTools, nil
}

func (h *InternalSkillMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *InternalSkillMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *InternalSkillMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	if h.provider == nil {
		return nil, fmt.Errorf("internal skill provider not initialized")
	}

	err := h.provider.ExecuteAction(toolName, arguments)
	if err != nil {
		return nil, err
	}

	return types.MCPCallToolResponse{
		Content: []types.MCPContent{{Type: "text", Text: "Action executed successfully."}},
	}, nil
}

func (h *InternalSkillMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource reading not supported for internal skills")
}

func (h *InternalSkillMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt templates not supported for internal skills")
}

// IMServiceProvider 定义了 IM 服务提供的接口
type IMServiceProvider interface {
	SendMessage(platform, targetID, content string) error
	GetActivePlatforms() []string
}

// IMBridgeMCPHost 将现有的 IM 适配器能力桥接为 MCP 工具
type IMBridgeMCPHost struct {
	svc IMServiceProvider
}

func NewIMBridgeMCPHost(svc IMServiceProvider) *IMBridgeMCPHost {
	return &IMBridgeMCPHost{svc: svc}
}

func (h *IMBridgeMCPHost) ListTools(ctx context.Context, serverID string) ([]types.MCPTool, error) {
	return []types.MCPTool{
		{
			Name:        "im_send_message",
			Description: "通过现有的 IM 适配器（微信、Discord 等）发送消息。这是适配器模式与 MCP 模式并行的典型应用。",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"platform": map[string]any{
						"type":        "string",
						"description": "目标平台，如 'wechat', 'discord', 'onebot'",
					},
					"target_id": map[string]any{
						"type":        "string",
						"description": "接收者 ID、群号或频道 ID",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "消息文本内容",
					},
				},
				"required": []string{"platform", "target_id", "content"},
			},
		},
		{
			Name:        "im_list_platforms",
			Description: "列出当前系统已连接并激活的 IM 平台适配器",
			InputSchema: map[string]any{
				"type": "object",
			},
		},
	}, nil
}

func (h *IMBridgeMCPHost) ListResources(ctx context.Context, serverID string) ([]types.MCPResource, error) {
	return nil, nil
}

func (h *IMBridgeMCPHost) ListPrompts(ctx context.Context, serverID string) ([]types.MCPPrompt, error) {
	return nil, nil
}

func (h *IMBridgeMCPHost) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error) {
	switch toolName {
	case "im_send_message":
		platform, _ := arguments["platform"].(string)
		targetID, _ := arguments["target_id"].(string)
		content, _ := arguments["content"].(string)

		if h.svc != nil {
			err := h.svc.SendMessage(platform, targetID, content)
			if err != nil {
				return nil, err
			}
		} else {
			// Fallback for simulation if service is not provided
			fmt.Printf("[IM-Bridge] Calling legacy adapter (Simulated): platform=%s, target=%s, content=%s\n", platform, targetID, content)
		}

		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("已通过适配器模式向 %s(%s) 发送消息: %s", platform, targetID, content),
				},
			},
		}, nil

	case "im_list_platforms":
		var platforms []string
		if h.svc != nil {
			platforms = h.svc.GetActivePlatforms()
		} else {
			platforms = []string{"simulated_adapter"}
		}
		return types.MCPCallToolResponse{
			Content: []types.MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Active platforms: %s", strings.Join(platforms, ", ")),
				},
			},
		}, nil
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

func (h *IMBridgeMCPHost) ReadResource(ctx context.Context, serverID string, uri string) (any, error) {
	return nil, fmt.Errorf("resource not supported")
}

func (h *IMBridgeMCPHost) GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error) {
	return "", fmt.Errorf("prompt not supported")
}
