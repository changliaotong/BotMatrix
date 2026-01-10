package common

import (
	"BotMatrix/common/bot"
	"BotMatrix/common/config"
	"BotMatrix/common/database"
	"BotMatrix/common/models"
	"BotMatrix/common/session"
	"BotMatrix/common/types"
	"BotMatrix/common/utils"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

// --- config.go ---
var GlobalConfig = config.GlobalConfig

func GetResolvedConfigPath() string {
	return config.GetResolvedConfigPath()
}

const CONFIG_FILE = config.CONFIG_FILE

// Redis Key 设计
const (
	REDIS_KEY_QUEUE_DEFAULT   = config.REDIS_KEY_QUEUE_DEFAULT
	REDIS_KEY_QUEUE_WORKER    = config.REDIS_KEY_QUEUE_WORKER
	REDIS_KEY_RATELIMIT_USER  = config.REDIS_KEY_RATELIMIT_USER
	REDIS_KEY_RATELIMIT_GROUP = config.REDIS_KEY_RATELIMIT_GROUP
	REDIS_KEY_IDEMPOTENCY     = config.REDIS_KEY_IDEMPOTENCY
	REDIS_KEY_SESSION_CONTEXT = config.REDIS_KEY_SESSION_CONTEXT
	REDIS_KEY_DYNAMIC_RULES   = config.REDIS_KEY_DYNAMIC_RULES
	REDIS_KEY_ACTION_QUEUE    = config.REDIS_KEY_ACTION_QUEUE
)

// --- db.go ---
func InitDB() error {
	db, err := database.InitDB(GlobalConfig)
	if err != nil {
		return err
	}
	GlobalManager.DB = db
	return nil
}

// --- gorm_compat.go ---
func InitGORM() error {
	db, err := database.InitGORM(GlobalConfig)
	if err != nil {
		return err
	}
	GlobalManager.GORMDB = db
	return nil
}

// --- gorm_manager.go ---
type GORMManager = database.GORMManager

func NewGORMManager() *GORMManager {
	return database.NewGORMManager(GlobalManager.GORMDB, nil)
}

// --- models & types ---
type UserGORM = models.UserInfo
type RoutingRuleGORM = models.RoutingRule

// type FissionConfigGORM = models.FissionConfigGORM // Deprecated/Missing
// type InvitationGORM = models.InvitationGORM // Deprecated/Missing
// type FissionTaskGORM = models.FissionTaskGORM // Deprecated/Missing
// type UserFissionRecordGORM = models.UserFissionRecordGORM // Deprecated/Missing
// type FissionRewardLogGORM = models.FissionRewardLogGORM // Deprecated/Missing
type GroupCacheGORM = models.GroupCache
type MemberCacheGORM = models.MemberCache
type FriendCacheGORM = models.FriendCache
type AIProviderGORM = models.AIProvider
type AIModelGORM = models.AIModel
type AIAgentGORM = models.AIAgent
type AIPromptTemplateGORM = models.AIPromptTemplate
type AIKnowledgeBaseGORM = models.AIKnowledgeBase
type AIUsageLogGORM = models.AIUsageLog
type AISkillGORM = models.AISkill
type AITrainingDataGORM = models.AITrainingData
type AIIntentGORM = models.AIIntent
type AIIntentRoutingGORM = models.AIIntentRouting
type GroupBotRoleGORM = models.GroupBotRole
type EnterpriseGORM = models.Enterprise

// type PlatformAccountGORM = models.PlatformAccountGORM // Check if exists
type DigitalEmployeeGORM = models.DigitalEmployee

// type EnterpriseMemberGORM = models.EnterpriseMemberGORM // Check if exists
// type B2BConnectionGORM = models.B2BConnectionGORM // Check if exists
type DigitalEmployeeKpiGORM = models.DigitalEmployeeKpi
type AISessionGORM = models.AISession
type AIChatMessageGORM = models.AIChatMessage
type BotEntityGORM = models.Member
type MessageLogGORM = models.MessageLog

type BotConfig = bot.BotConfig
type ConnectionHandler = bot.ConnectionHandler

type SessionContext = types.SessionContext
type SessionState = types.SessionState

// --- bot & connection ---
type ConnectionManager = bot.ConnectionManager

func NewConnectionManager() *ConnectionManager {
	return bot.NewConnectionManager()
}

type Manager = bot.Manager

var GlobalManager = bot.GlobalManager

func NewManager() *Manager {
	return bot.NewManager()
}

// --- utils & i18n & webui ---
type Translator = utils.Translator

func InitTranslator(dir, lang string) {
	utils.InitTranslator(dir, lang)
}

func T(lang string, key string, args ...any) string {
	return utils.T(lang, key, args...)
}

func GenerateRandomToken(length int) string {
	return utils.GenerateRandomToken(length)
}

func DecodeMapToStruct(m any, v any) error {
	return utils.DecodeMapToStruct(m, v)
}

func SendJSONResponse(w http.ResponseWriter, success bool, message string, data any) {
	utils.SendJSONResponse(w, success, message, data)
}

func SendJSONResponseWithCode(w http.ResponseWriter, success bool, message string, code string, data any) {
	utils.SendJSONResponseWithCode(w, success, message, code, data)
}

func ToString(v any) string {
	return utils.ToString(v)
}

func ToInt64(v any) int64 {
	return utils.ToInt64(v)
}

var Upgrader = utils.Upgrader

func ReadJSONWithNumber(conn *websocket.Conn, v any) error {
	return utils.ReadJSONWithNumber(conn, v)
}

// --- bot & connection ---
type BaseBot = bot.BaseBot

func NewBaseBot(port int) *BaseBot {
	return bot.NewBaseBot(port)
}

func NewTicker(seconds int) <-chan time.Time {
	return bot.NewTicker(seconds)
}

// --- types ---
type ApiResponse = types.ApiResponse
type InternalMessage = types.InternalMessage
type BotClient = types.BotClient
type WorkerClient = types.WorkerClient
type LogEntry = types.LogEntry
type ProcInfo = types.ProcInfo
type GroupInfo = types.GroupInfo
type MemberInfo = types.MemberInfo
type FriendInfo = types.FriendInfo
type User = types.User
type ConnectionStats = types.ConnectionStats
type BotStatDetail = types.BotStatDetail

type FissionUser = models.FissionUser

func InitDockerClient() (*client.Client, error) {
	return utils.InitDockerClient()
}

// --- session ---
type SessionStore = session.SessionStore
type RedisSessionStore = session.RedisSessionStore

func NewRedisSessionStore(client any) *RedisSessionStore {
	return session.NewRedisSessionStore(client)
}

func SessionKey(pluginID, groupID, userID string) string {
	return session.SessionKey(pluginID, groupID, userID)
}
