package plugins

import (
	"BotMatrix/common/ai"
	"BotMatrix/common/ai/employee"
	"BotMatrix/common/ai/rag"
	"BotMatrix/common/plugin/core"
	"BotMatrix/common/types"
	"context"
	"encoding/json"
	"fmt"
)

// AIPlugin 提供分布式 AI 能力的内部插件
type AIPlugin struct {
	robot       core.Robot
	aiService   ai.AIService
	employeeSvc employee.DigitalEmployeeService
	memorySvc   employee.CognitiveMemoryService
	taskSvc     employee.DigitalEmployeeTaskService
}

func NewAIPlugin(aiService ai.AIService) *AIPlugin {
	return &AIPlugin{
		aiService: aiService,
	}
}

func (p *AIPlugin) SetEmployeeServices(emp employee.DigitalEmployeeService, mem employee.CognitiveMemoryService, task employee.DigitalEmployeeTaskService) {
	p.employeeSvc = emp
	p.memorySvc = mem
	p.taskSvc = task
}

func (p *AIPlugin) Name() string { return "AIPlugin" }
func (p *AIPlugin) Description() string {
	return "Provides distributed AI capabilities (Chat, Embedding, RAG)"
}
func (p *AIPlugin) Version() string { return "1.0.0" }

func (p *AIPlugin) Init(robot core.Robot) {
	p.robot = robot

	// 注册基础 AI 技能
	robot.HandleSkill("ai_chat", p.handleChat)
	robot.HandleSkill("ai_embedding", p.handleEmbedding)

	// 注册 RAG 技能
	robot.HandleSkill("ai_rag_search", p.handleRagSearch)

	// 注册 MCP 技能
	robot.HandleSkill("mcp_list_tools", p.handleMcpListTools)
	robot.HandleSkill("mcp_call_tool", p.handleMcpCallTool)

	// 注册数字员工技能
	robot.HandleSkill("employee_get_info", p.handleEmployeeGetInfo)
	robot.HandleSkill("employee_plan_task", p.handleEmployeePlanTask)
}

func (p *AIPlugin) GetSkills() []core.SkillCapability {
	return []core.SkillCapability{
		{
			Name:        "ai_chat",
			Description: "Execute a chat request using configured LLM",
			Usage:       "Internal use only",
			Params: map[string]string{
				"request": "JSON string of ai.ChatRequest",
			},
		},
		{
			Name:        "ai_embedding",
			Description: "Generate embeddings for given input",
			Usage:       "Internal use only",
			Params: map[string]string{
				"request": "JSON string of ai.EmbeddingRequest",
			},
		},
		{
			Name:        "ai_rag_search",
			Description: "Search knowledge base for relevant chunks",
			Usage:       "Internal use only",
			Params: map[string]string{
				"kb_id":  "Knowledge base ID",
				"query":  "Search query",
				"top_k":  "Number of results (optional)",
				"filter": "JSON string of ai.SearchFilter (optional)",
			},
		},
		{
			Name:        "mcp_list_tools",
			Description: "List available MCP tools for a context",
			Usage:       "Internal use only",
			Params: map[string]string{
				"user_id": "User ID",
				"org_id":  "Org ID",
			},
		},
		{
			Name:        "mcp_call_tool",
			Description: "Call an MCP tool",
			Usage:       "Internal use only",
			Params: map[string]string{
				"name": "Full tool name (server__tool)",
				"args": "JSON string of arguments",
			},
		},
		{
			Name:        "employee_get_info",
			Description: "Get digital employee information by bot ID",
			Usage:       "Internal use only",
			Params: map[string]string{
				"bot_id": "Bot ID",
			},
		},
		{
			Name:        "employee_plan_task",
			Description: "Plan a task for a digital employee",
			Usage:       "Internal use only",
			Params: map[string]string{
				"execution_id": "Task Execution ID",
			},
		},
	}
}

func (p *AIPlugin) handleEmployeeGetInfo(params map[string]string) (string, error) {
	if p.employeeSvc == nil {
		return "", fmt.Errorf("employee service not available")
	}
	botID := params["bot_id"]
	emp, err := p.employeeSvc.GetEmployeeByBotID(botID)
	if err != nil {
		return "", err
	}
	respBytes, _ := json.Marshal(emp)
	return string(respBytes), nil
}

func (p *AIPlugin) handleEmployeePlanTask(params map[string]string) (string, error) {
	if p.taskSvc == nil {
		return "", fmt.Errorf("task service not available")
	}
	execID := params["execution_id"]
	err := p.taskSvc.PlanTask(context.Background(), execID)
	if err != nil {
		return "", err
	}
	return "success", nil
}

func (p *AIPlugin) handleChat(params map[string]string) (string, error) {
	reqStr := params["request"]
	var req ai.ChatRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return "", fmt.Errorf("invalid chat request: %v", err)
	}

	resp, err := p.aiService.ChatSimple(context.Background(), req)
	if err != nil {
		return "", err
	}

	respBytes, _ := json.Marshal(resp)
	return string(respBytes), nil
}

func (p *AIPlugin) handleEmbedding(params map[string]string) (string, error) {
	reqStr := params["request"]
	var req ai.EmbeddingRequest
	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return "", fmt.Errorf("invalid embedding request: %v", err)
	}

	resp, err := p.aiService.CreateEmbeddingSimple(context.Background(), req)
	if err != nil {
		return "", err
	}

	respBytes, _ := json.Marshal(resp)
	return string(respBytes), nil
}

func (p *AIPlugin) handleRagSearch(params map[string]string) (string, error) {
	kbID := params["kb_id"]
	query := params["query"]
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	// 获取 KB 实例 (从 AIService 扩展)
	type RagCapable interface {
		GetKnowledgeBase() *rag.PostgresKnowledgeBase
	}
	rc, ok := p.aiService.(RagCapable)
	if !ok || rc.GetKnowledgeBase() == nil {
		return "", fmt.Errorf("RAG capability not available on this worker")
	}

	// 搜索逻辑
	limit := 5
	if l := params["top_k"]; l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	var filter ai.SearchFilter
	if f := params["filter"]; f != "" {
		json.Unmarshal([]byte(f), &filter)
	}

	// 执行搜索 (这里假设 kbID 已经匹配)
	chunks, err := rc.GetKnowledgeBase().Search(context.Background(), query, limit, &types.SearchFilter{
		OwnerType: "kb",
		OwnerID:   kbID,
	})
	if err != nil {
		return "", fmt.Errorf("search failed: %v", err)
	}

	respBytes, _ := json.Marshal(chunks)
	return string(respBytes), nil
}

func (p *AIPlugin) handleMcpListTools(params map[string]string) (string, error) {
	type McpCapable interface {
		GetMCPManager() *ai.MCPManager
	}
	mc, ok := p.aiService.(McpCapable)
	if !ok || mc.GetMCPManager() == nil {
		return "", fmt.Errorf("MCP capability not available on this worker")
	}

	var userID, orgID uint
	fmt.Sscanf(params["user_id"], "%d", &userID)
	fmt.Sscanf(params["org_id"], "%d", &orgID)

	tools, err := mc.GetMCPManager().GetToolsForContext(context.Background(), userID, orgID)
	if err != nil {
		return "", fmt.Errorf("list tools failed: %v", err)
	}

	respBytes, _ := json.Marshal(tools)
	return string(respBytes), nil
}

func (p *AIPlugin) handleMcpCallTool(params map[string]string) (string, error) {
	type McpCapable interface {
		GetMCPManager() *ai.MCPManager
	}
	mc, ok := p.aiService.(McpCapable)
	if !ok || mc.GetMCPManager() == nil {
		return "", fmt.Errorf("MCP capability not available on this worker")
	}

	name := params["name"]
	argsStr := params["args"]
	var args map[string]any
	if argsStr != "" {
		json.Unmarshal([]byte(argsStr), &args)
	}

	result, err := mc.GetMCPManager().CallTool(context.Background(), name, args)
	if err != nil {
		return "", fmt.Errorf("call tool failed: %v", err)
	}

	respBytes, _ := json.Marshal(result)
	return string(respBytes), nil
}
