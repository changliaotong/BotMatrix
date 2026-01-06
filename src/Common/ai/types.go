package ai

import (
	"BotMatrix/common/models"
	"BotMatrix/common/types"
	"context"

	"gorm.io/gorm"
)

// Re-export types from common/types for convenience and backward compatibility if needed,
// but it's better to use them directly.
type Role = types.Role

const (
	RoleSystem    = types.RoleSystem
	RoleUser      = types.RoleUser
	RoleAssistant = types.RoleAssistant
	RoleTool      = types.RoleTool
)

type Message = types.Message
type ToolCall = types.ToolCall
type FunctionCall = types.FunctionCall
type Tool = types.Tool
type ChatRequest = types.ChatRequest
type ChatResponse = types.ChatResponse
type Choice = types.Choice
type ChatStreamResponse = types.ChatStreamResponse
type EmbeddingRequest = types.EmbeddingRequest
type EmbeddingResponse = types.EmbeddingResponse
type EmbeddingData = types.EmbeddingData
type UsageInfo = types.UsageInfo
type FunctionDefinition = types.FunctionDefinition
type SearchFilter = types.SearchFilter
type DocChunk = types.DocChunk
type KnowledgeBase = types.KnowledgeBase
type MCPManagerInterface = types.MCPManagerInterface

type MCPTool = types.MCPTool
type MCPResource = types.MCPResource
type MCPPrompt = types.MCPPrompt
type MCPPromptArgument = types.MCPPromptArgument
type MCPServerScope = types.MCPServerScope

const (
	ScopeGlobal = types.ScopeGlobal
	ScopeOrg    = types.ScopeOrg
	ScopeUser   = types.ScopeUser
)

type MCPServerInfo = types.MCPServerInfo
type MCPListToolsResponse = types.MCPListToolsResponse
type MCPCallToolRequest = types.MCPCallToolRequest
type MCPCallToolResponse = types.MCPCallToolResponse
type MCPContent = types.MCPContent
type MCPHost = types.MCPHost
type MCPManager = types.MCPManager
type RegisteredServer = types.RegisteredServer
type PrivacyFilter = types.PrivacyFilter
type MaskContext = types.MaskContext
type SensitiveType = types.SensitiveType

const (
	SensitivePhone  = types.SensitivePhone
	SensitiveEmail  = types.SensitiveEmail
	SensitiveIDCard = types.SensitiveIDCard
	SensitiveIP     = types.SensitiveIP
	SensitiveCustom = types.SensitiveCustom
)

type Capability = types.Capability

// AIServiceProvider AI 服务提供者接口 (由 Nexus 或 Worker 实现)
type AIService = types.AIService
type AIServiceProvider interface {
	SyncSkillCall(ctx context.Context, skillName string, params map[string]any) (any, error)
	GetWorkers() []types.WorkerInfo
	CheckPermission(ctx context.Context, botID string, userID uint, orgID uint, skillName string) (bool, error)
	GetGORMDB() *gorm.DB
	GetKnowledgeBase() KnowledgeBase
	GetManifest() *SystemManifest
	IsDigitalEmployeeEnabled() bool
}

type ContentPart = types.ContentPart
type ImageURLValue = types.ImageURLValue

// Client AI 客户端接口
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamResponse, error)
	CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
	GetEmployeeByBotID(botID string) (*models.DigitalEmployee, error)
	PlanTask(ctx context.Context, executionID string) error
}

type BotIdentity = types.BotIdentity
type SystemManifest = types.SystemManifest
type Manager = types.Manager
