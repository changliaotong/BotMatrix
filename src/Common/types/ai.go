package types

import (
	"BotMatrix/common/models"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Manager 定义了 AI 处理模块需要的核心依赖接口
type Manager interface {
	GetGORMDB() *gorm.DB
	GetKnowledgeBase() KnowledgeBase
	GetAIService() AIService
	GetB2BService() B2BService
	GetCognitiveMemoryService() CognitiveMemoryService
	GetDigitalEmployeeService() DigitalEmployeeService
	GetDigitalEmployeeTaskService() DigitalEmployeeTaskService
	GetTaskManager() TaskManagerInterface
	GetMCPManager() MCPManagerInterface
	ValidateToken(token string) (*UserClaims, error)
}

// B2BService 企业间通信服务接口
type B2BService interface {
	VerifyB2BToken(tokenString string) (*models.Enterprise, error)
	CheckDispatchPermission(employeeID, targetOrgID uint, action string) (bool, error)
}

// CognitiveMemoryService 认知记忆服务接口
type CognitiveMemoryService interface {
	LearnFromURL(ctx context.Context, botID string, url string, category string) error
	LearnFromContent(ctx context.Context, botID string, content []byte, filename string, category string) error
	ConsolidateMemories(ctx context.Context, userID string, botID string, aiSvc AIService) error
	GetRelevantMemories(ctx context.Context, userID string, botID string, query string) ([]models.CognitiveMemory, error)
	GetRoleMemories(ctx context.Context, botID string) ([]models.CognitiveMemory, error)
	SaveMemory(ctx context.Context, memory *models.CognitiveMemory) error
	ForgetMemory(ctx context.Context, memoryID uint) error
	SearchMemories(ctx context.Context, botID string, query string, category string) ([]models.CognitiveMemory, error)
	SetEmbeddingService(svc any)
}

// DigitalEmployeeService 数字员工核心服务接口
type DigitalEmployeeService interface {
	GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error)
	RecordKpi(employeeID uint, metric string, score float64) error
	UpdateOnlineStatus(botID string, status string) error
	ConsumeSalary(botID string, tokens int64) error
	CheckSalaryLimit(botID string) (bool, error)
	UpdateSalary(botID string, salaryToken *int64, salaryLimit *int64) error
	AutoEvolve(employeeID uint) error
}

// DigitalEmployeeTaskService 数字员工任务服务接口
type DigitalEmployeeTaskService interface {
	CreateTask(ctx context.Context, task *models.DigitalEmployeeTask) error
	UpdateTaskStatus(ctx context.Context, executionID string, status string, progress int) error
	GetTaskByExecutionID(ctx context.Context, executionID string) (*models.DigitalEmployeeTask, error)
	AssignTask(ctx context.Context, executionID string, assigneeID uint) error
	PlanTask(ctx context.Context, executionID string) error
	ExecuteTask(ctx context.Context, executionID string) error
	ExecuteStep(ctx context.Context, executionID string, stepIndex int) error
	ApproveTask(ctx context.Context, executionID string) error
	CreateSubTask(ctx context.Context, parentExecutionID string, subTask *models.DigitalEmployeeTask) error
	RecordTaskResult(ctx context.Context, executionID string, result string, success bool) error
}

// Role 定义对话角色
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message 对话消息
type Message struct {
	Role       Role           `json:"role"`
	Content    any            `json:"content"`                // 改为 any 以支持 string 或 []ContentPart
	Name       string         `json:"name,omitempty"`         // 用于 Tool 角色
	ToolCallID string         `json:"tool_call_id,omitempty"` // 用于 Tool 角色
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`   // 用于 Assistant 角色发起调用
	Extras     map[string]any `json:"extras,omitempty"`       // 用于传递元数据 (如 SessionID, ExecutionID)
}

// ContentPart 消息内容部分 (多模态)
type ContentPart struct {
	Type     string         `json:"type"`                // "text" 或 "image_url"
	Text     string         `json:"text,omitempty"`      // 当 Type 为 "text" 时使用
	ImageURL *ImageURLValue `json:"image_url,omitempty"` // 当 Type 为 "image_url" 时使用
}

// ImageURLValue 图像 URL 详情
type ImageURLValue struct {
	URL    string `json:"url"`              // 可以是 http 链接或 base64 (data:image/jpeg;base64,...)
	Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}

// ToolCall 具体的工具调用请求
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // 总是 "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用详情
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON 字符串
}

// Tool 工具定义 (Function Definition)
type Tool struct {
	Type     string             `json:"type"` // 总是 "function"
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 函数定义详情
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema
}

// ChatRequest 对话请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	ID      string    `json:"id"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

// Choice 响应选项
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

// UsageInfo Token 消耗统计
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingRequest 向量请求
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"` // 改为 any 以支持 []string 或多模态 []map[string]any
}

// EmbeddingResponse 向量响应
type EmbeddingResponse struct {
	Data  []EmbeddingData `json:"data"`
	Model string          `json:"model"`
	Usage UsageInfo       `json:"usage"`
}

// EmbeddingData 向量数据
type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// SearchFilter 知识库搜索过滤条件
type SearchFilter struct {
	OwnerType string // system, user, group, bot
	OwnerID   string
	BotID     string
	Status    string // active, paused
}

// DocChunk 搜索返回的文档切片
type DocChunk struct {
	ID      string  `json:"id"`
	Content string  `json:"content"`
	Source  string  `json:"source"`
	Score   float64 `json:"score"`
	Title   string  `json:"title"`
	DocID   uint    `json:"doc_id"`
}

// ChatStreamResponse 流式响应增量
type ChatStreamResponse struct {
	ID      string         `json:"id"`
	Choices []StreamChoice `json:"choices"`
	Error   error          `json:"-"`
}

// StreamChoice 流式选项
type StreamChoice struct {
	Delta        MessageDelta `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

// MessageDelta 消息增量
type MessageDelta struct {
	Role      Role       `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// KnowledgeBase 知识库核心接口
type KnowledgeBase interface {
	Search(ctx context.Context, query string, limit int, filter *SearchFilter) ([]DocChunk, error)
}

// AIParserInterface defines the interface for the AI parser
type AIParserInterface interface {
	GetManifest() *SystemManifest
	GetSystemPrompt() string
	MatchSkillByRegex(input string) (*Capability, bool)
	MatchSkillByLLM(ctx context.Context, input string, modelID uint, parseCtx map[string]any) (*ParseResult, error)
	UpdateSkills(skills []Capability)
}

// AIActionType AI 解析的目标动作类型
type AIActionType string

const (
	AIActionCreateTask   AIActionType = "create_task"
	AIActionAdjustPolicy AIActionType = "adjust_policy"
	AIActionManageTags   AIActionType = "manage_tags"
	AIActionSystemQuery  AIActionType = "system_query"
	AIActionSkillCall    AIActionType = "skill_call"
	AIActionCancelTask   AIActionType = "cancel_task"
	AIActionBatch        AIActionType = "batch_task"
)

// ParseResult AI 解析结果
type ParseResult struct {
	DraftID    string         `json:"draft_id"`
	Intent     AIActionType   `json:"intent"`
	Summary    string         `json:"summary"`
	Data       any            `json:"data"`
	IsSafe     bool           `json:"is_safe"`
	Analysis   string         `json:"analysis"`
	SubActions []*ParseResult `json:"sub_actions"`
}

// ParseRequest AI 解析请求
type ParseRequest struct {
	Input      string         `json:"input"`
	ActionType AIActionType   `json:"action_type"` // 可选，明确指定意图
	Context    map[string]any `json:"context"`     // 上下文信息
}

// DispatcherInterface defines the interface for the task dispatcher
type DispatcherInterface interface {
	GetActions() []string
	ExecuteAction(ctx context.Context, action string, params any) (any, error)
}

// InterceptorManagerInterface defines the interface for the task interceptor manager
type InterceptorManagerInterface interface {
	GetInterceptors() []string
	GetStrategies() []models.Strategy
	GetStrategy(name string) (*models.Strategy, error)
	DeleteStrategy(name string)
}

// TaskManagerInterface defines the interface for the task manager
type TaskManagerInterface interface {
	GetAI() AIParserInterface
	GetDispatcher() DispatcherInterface
	GetInterceptors() InterceptorManagerInterface
	GetTagging() TaggingManagerInterface
	CreateTask(task *models.Task, isEnterprise bool) error
	GetStrategyConfig(name string, out any) bool
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// TaggingManagerInterface defines the interface for the tagging manager
type TaggingManagerInterface interface {
	AddTag(targetType, targetID, tagName string) error
	RemoveTag(targetType, targetID, tagName string) error
	GetTargetsByTags(targetType string, tags []string, logic string) ([]string, error)
	GetTagsByTarget(targetType, targetID string) ([]string, error)
}

// BotIdentity 机器人身份定义 (自举核心)
type BotIdentity struct {
	Name        string            `json:"name"`        // 机器人名称
	Role        string            `json:"role"`        // 核心角色定位
	Personality string            `json:"personality"` // 性格特征描述
	Knowledge   []string          `json:"knowledge"`   // 核心背景知识 (静态注入)
	HowTo       map[string]string `json:"how_to"`      // 核心功能操作指南 (静态注入)
}

// SystemManifest 系统功能清单
type SystemManifest struct {
	Version       string                `json:"version"`
	Identity      BotIdentity           `json:"identity"`     // 机器人身份
	Intents       map[string]string     `json:"intents"`      // 意图 -> 说明
	Actions       map[string]Capability `json:"actions"`      // 动作 -> 详情
	Triggers      map[string]Capability `json:"triggers"`     // 触发方式 -> 详情
	Skills        []Capability          `json:"skills"`       // 业务技能 (由 Worker 提供)
	GlobalRules   []string              `json:"global_rules"` // 全局约束
	KnowledgeBase KnowledgeBase         `json:"-"`            // 外部知识库接口
	RAGEnabled    bool                  `json:"rag_enabled"`
}

// GenerateTools 将 SystemManifest 转换为 OpenAI 格式的工具
func (s *SystemManifest) GenerateTools() []Tool {
	var tools []Tool
	// 将 Actions, Triggers, Skills 转换为 Tool 格式
	for _, cap := range s.Actions {
		tools = append(tools, cap.ToTool())
	}
	for _, cap := range s.Triggers {
		tools = append(tools, cap.ToTool())
	}
	for _, cap := range s.Skills {
		tools = append(tools, cap.ToTool())
	}
	return tools
}

// GenerateSystemPrompt 生成系统提示词
func (s *SystemManifest) GenerateSystemPrompt() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("你是 %s (%s)。\n", s.Identity.Name, s.Identity.Role))
	sb.WriteString(s.Identity.Personality + "\n\n")

	if len(s.GlobalRules) > 0 {
		sb.WriteString("全局规则:\n")
		for _, rule := range s.GlobalRules {
			sb.WriteString("- " + rule + "\n")
		}
		sb.WriteString("\n")
	}

	if len(s.Intents) > 0 {
		sb.WriteString("你可以处理以下意图:\n")
		for intent, desc := range s.Intents {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", intent, desc))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// MCPManagerInterface MCP 管理器接口
type MCPManagerInterface interface {
	GetToolsForContext(ctx context.Context, userID uint, orgID uint) ([]Tool, error)
	CallTool(ctx context.Context, fullName string, args map[string]any) (any, error)
	SetKnowledgeBase(kb KnowledgeBase)
	GetKnowledgeBase() KnowledgeBase
}

// AIService 定义任务系统需要的 AI 能力接口
type AIService interface {
	Chat(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error)
	ChatSimple(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatAgent(ctx context.Context, modelID uint, messages []Message, tools []Tool) (*ChatResponse, error)
	ChatStream(ctx context.Context, modelID uint, messages []Message, tools []Tool) (<-chan ChatStreamResponse, error)
	CreateEmbedding(ctx context.Context, modelID uint, input any) (*EmbeddingResponse, error)
	CreateEmbeddingSimple(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
	ExecuteTool(ctx context.Context, botID string, userID uint, orgID uint, toolCall ToolCall) (any, error)
	GetMCPManager() MCPManagerInterface
	GetProvider(id uint) (*models.AIProvider, error)
	DispatchIntent(msg InternalMessage) (string, error)
	ChatWithEmployee(employee *models.DigitalEmployee, msg InternalMessage, targetOrgID uint) (string, error)
}

// MCP (Model Context Protocol) 核心结构定义
// 参考: https://modelcontextprotocol.io/

// MCPTool 定义了 MCP 协议中的工具格式
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPResource 定义了 MCP 协议中的资源
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCPPrompt 定义了 MCP 协议中的提示词模板
type MCPPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Arguments   []MCPPromptArgument `json:"arguments,omitempty"`
}

// Capability 定义系统的一个功能原子
type Capability struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Category     string            `json:"category"` // 分类 (e.g., tools, entertainment, ai)
	Params       map[string]string `json:"params"`   // 参数名 -> 说明
	Required     []string          `json:"required"` // 必填参数列表
	Example      string            `json:"example"`  // 示例输入
	IsEnterprise bool              `json:"is_enterprise"`
	Regex        string            `json:"regex"`         // 正则触发器
	RiskLevel    string            `json:"risk_level"`    // 风险等级 (low, medium, high)
	DefaultRoles []string          `json:"default_roles"` // 默认允许的角色 (owner, admin, member)
}

// ToTool 将 Capability 转换为 LLM 可识别的工具定义
func (c *Capability) ToTool() Tool {
	properties := make(map[string]any)
	for name, desc := range c.Params {
		properties[name] = map[string]any{
			"type":        "string",
			"description": desc,
		}
	}

	return Tool{
		Type: "function",
		Function: FunctionDefinition{
			Name:        c.Name,
			Description: c.Description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   c.Required,
			},
		},
	}
}

// GenerateSkillGuide 生成技能的操作指南文本，用于 RAG 索引
func (c *Capability) GenerateSkillGuide() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("技能名称: %s\n", c.Name))
	sb.WriteString(fmt.Sprintf("分类: %s\n", c.Category))
	sb.WriteString(fmt.Sprintf("描述: %s\n", c.Description))

	if len(c.Params) > 0 {
		sb.WriteString("参数说明:\n")
		for name, desc := range c.Params {
			required := ""
			for _, r := range c.Required {
				if r == name {
					required = " (必填)"
					break
				}
			}
			sb.WriteString(fmt.Sprintf("- %s: %s%s\n", name, desc, required))
		}
	}

	if c.Example != "" {
		sb.WriteString(fmt.Sprintf("使用示例: %s\n", c.Example))
	}

	if c.Regex != "" {
		sb.WriteString(fmt.Sprintf("匹配规则: %s\n", c.Regex))
	}

	return sb.String()
}

// MCPPromptArgument 定义了 MCP 协议中的提示词模板参数
type MCPPromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// MCPServerScope 定义 MCP 服务器的可用范围
type MCPServerScope string

const (
	ScopeGlobal MCPServerScope = "global" // 全局可用
	ScopeOrg    MCPServerScope = "org"    // 组织/群组可用
	ScopeUser   MCPServerScope = "user"   // 仅个人可用
)

// MCPServerInfo 包含 MCP 服务器的元数据
type MCPServerInfo struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Scope       MCPServerScope `json:"scope"`
	OwnerID     uint           `json:"owner_id,omitempty"` // 所属用户或组织 ID
}

// MCPListToolsResponse tools/list 响应
type MCPListToolsResponse struct {
	Tools []MCPTool `json:"tools"`
}

// 将 MCP 工具转换为 OpenAI 格式的工具
func (t *MCPTool) ToOpenAITool() Tool {
	return Tool{
		Type: "function",
		Function: FunctionDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.InputSchema,
		},
	}
}

// MCPCallToolRequest tools/call 请求
type MCPCallToolRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// MCPCallToolResponse tools/call 响应
type MCPCallToolResponse struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
	Type string `json:"type"` // text, image, resource
	Text string `json:"text,omitempty"`
}

// MCPHost 定义了作为 MCP 客户端（Host）的行为
type MCPHost interface {
	// 列表发现
	ListTools(ctx context.Context, serverID string) ([]MCPTool, error)
	ListResources(ctx context.Context, serverID string) ([]MCPResource, error)
	ListPrompts(ctx context.Context, serverID string) ([]MCPPrompt, error)

	// 调用执行
	CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]any) (any, error)
	ReadResource(ctx context.Context, serverID string, uri string) (any, error)
	GetPrompt(ctx context.Context, serverID string, promptName string, arguments map[string]any) (string, error)
}
