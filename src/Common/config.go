package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// GlobalConfig 全局配置实例
var GlobalConfig = &AppConfig{
	WSPort:               "",
	WebUIPort:            "",
	StatsFile:            "",
	RedisAddr:            "",
	RedisPwd:             "",
	JWTSecret:            "",
	DefaultAdminPassword: "",
	PGHost:               "",
	PGPort:               0,
	PGUser:               "",
	PGPassword:           "",
	PGDBName:             "",
	PGSSLMode:            "",
	EnableSkill:          false,
	LogLevel:             "",
	AutoReply:            false,
}

// 注意：为了兼容现有代码，保留这些全局变量
var (
	WS_PORT                string
	WEBUI_PORT             string
	STATS_FILE             string
	REDIS_ADDR             string
	REDIS_PWD              string
	JWT_SECRET             string
	DEFAULT_ADMIN_PASSWORD string
	PG_HOST                string
	PG_PORT                int
	PG_USER                string
	PG_PASSWORD            string
	PG_DBNAME              string
	PG_SSLMODE             string
	ENABLE_SKILL           bool
	LOG_LEVEL              string
	AUTO_REPLY             bool
	AZURE_TRANSLATE_KEY    string
	AZURE_TRANSLATE_END    string
	AZURE_TRANSLATE_REG    string
)

var resolvedConfigPath = CONFIG_FILE

const CONFIG_FILE = "config.json"

// Redis Key 设计
const (
	// 消息队列 (支持多 Worker 横向扩展)
	REDIS_KEY_QUEUE_DEFAULT = "botmatrix:queue:default"
	REDIS_KEY_QUEUE_WORKER  = "botmatrix:queue:worker:%s"

	// 限流 (防刷、防滥用)
	REDIS_KEY_RATELIMIT_USER  = "botmatrix:ratelimit:user:%s"
	REDIS_KEY_RATELIMIT_GROUP = "botmatrix:ratelimit:group:%s"

	// 幂等与去重 (防重复回复)
	REDIS_KEY_IDEMPOTENCY = "botmatrix:msg:idempotency:%s"

	// 会话上下文与状态缓存 (支持 TTL)
	REDIS_KEY_SESSION_CONTEXT = "botmatrix:session:%s:%s" // platform:user_id

	// 动态路由规则
	REDIS_KEY_DYNAMIC_RULES = "botmatrix:rules:routing"

	// 动态限流配置
	REDIS_KEY_CONFIG_RATELIMIT = "botmatrix:config:ratelimit"

	// 动态 TTL 配置
	REDIS_KEY_CONFIG_TTL = "botmatrix:config:ttl"
)

func init() {
	// 1. 设置默认值
	WS_PORT = GlobalConfig.WSPort
	WEBUI_PORT = GlobalConfig.WebUIPort
	STATS_FILE = GlobalConfig.StatsFile
	REDIS_ADDR = GlobalConfig.RedisAddr
	REDIS_PWD = GlobalConfig.RedisPwd
	JWT_SECRET = GlobalConfig.JWTSecret
	DEFAULT_ADMIN_PASSWORD = GlobalConfig.DefaultAdminPassword
	PG_HOST = GlobalConfig.PGHost
	PG_PORT = GlobalConfig.PGPort
	PG_USER = GlobalConfig.PGUser
	PG_PASSWORD = GlobalConfig.PGPassword
	PG_DBNAME = GlobalConfig.PGDBName
	PG_SSLMODE = GlobalConfig.PGSSLMode
	ENABLE_SKILL = GlobalConfig.EnableSkill
	LOG_LEVEL = GlobalConfig.LogLevel
	AUTO_REPLY = GlobalConfig.AutoReply

	// 2. 尝试从文件加载
	loadConfigFromFile()

	// 3. 环境变量覆盖 (最高优先级)
	loadConfigFromEnv()

	// 4. 同步回 GlobalConfig
	syncToGlobalConfig()
	absPath, _ := filepath.Abs(resolvedConfigPath)
	log.Printf("[INFO] 系统配置初始化完成: LogLevel=%s, AutoReply=%v, WebUIPort=%s, ConfigFile=%s", GlobalConfig.LogLevel, GlobalConfig.AutoReply, GlobalConfig.WebUIPort, absPath)
}

func loadConfigFromFile() {
	// 尝试加载配置文件
	// 首先检查当前目录下的 config.json
	// 如果不存在，检查可执行文件所在目录下的 config.json

	targetFile := CONFIG_FILE
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		// 尝试从可执行文件目录查找
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			potentialFile := filepath.Join(exeDir, CONFIG_FILE)
			if _, err := os.Stat(potentialFile); err == nil {
				targetFile = potentialFile
			}
		}
	}

	absPath, _ := filepath.Abs(targetFile)
	resolvedConfigPath = absPath
	log.Printf("[INFO] 配置文件解析路径: %s", absPath)

	if _, err := os.Stat(targetFile); err == nil {
		log.Printf("[INFO] 正在从配置文件加载: %s", absPath)
		data, err := os.ReadFile(targetFile)
		if err != nil {
			log.Printf("[WARN] 读取配置文件失败: %v", err)
			return
		}
		var fileConfig AppConfig
		if err := json.Unmarshal(data, &fileConfig); err != nil {
			log.Printf("[WARN] 解析配置文件失败: %v", err)
			return
		}

		if fileConfig.WSPort != "" {
			WS_PORT = fileConfig.WSPort
		}
		if fileConfig.WebUIPort != "" {
			WEBUI_PORT = fileConfig.WebUIPort
		}
		if fileConfig.StatsFile != "" {
			STATS_FILE = fileConfig.StatsFile
		}
		if fileConfig.RedisAddr != "" {
			REDIS_ADDR = fileConfig.RedisAddr
		}
		if fileConfig.RedisPwd != "" {
			REDIS_PWD = fileConfig.RedisPwd
		}
		if fileConfig.JWTSecret != "" {
			JWT_SECRET = fileConfig.JWTSecret
		}
		if fileConfig.DefaultAdminPassword != "" {
			DEFAULT_ADMIN_PASSWORD = fileConfig.DefaultAdminPassword
		}
		if fileConfig.PGHost != "" {
			PG_HOST = fileConfig.PGHost
		}
		if fileConfig.PGPort != 0 {
			PG_PORT = fileConfig.PGPort
		}
		if fileConfig.PGUser != "" {
			PG_USER = fileConfig.PGUser
		}
		if fileConfig.PGPassword != "" {
			PG_PASSWORD = fileConfig.PGPassword
		}
		if fileConfig.PGDBName != "" {
			PG_DBNAME = fileConfig.PGDBName
		}
		if fileConfig.PGSSLMode != "" {
			PG_SSLMODE = fileConfig.PGSSLMode
		}
		if fileConfig.LogLevel != "" {
			LOG_LEVEL = fileConfig.LogLevel
		}
		if fileConfig.AzureTranslateKey != "" {
			AZURE_TRANSLATE_KEY = fileConfig.AzureTranslateKey
		}
		if fileConfig.AzureTranslateEndpoint != "" {
			AZURE_TRANSLATE_END = fileConfig.AzureTranslateEndpoint
		}
		if fileConfig.AzureTranslateRegion != "" {
			AZURE_TRANSLATE_REG = fileConfig.AzureTranslateRegion
		}
		// EnableSkill 和 AutoReply 只要在配置文件中存在就覆盖默认值
		ENABLE_SKILL = fileConfig.EnableSkill
		AUTO_REPLY = fileConfig.AutoReply
		log.Printf("[INFO] 已从 %s 加载配置", CONFIG_FILE)

		// 同步到 GlobalConfig
		syncToGlobalConfig()
	}
}

func loadConfigFromEnv() {
	if v := os.Getenv("WS_PORT"); v != "" {
		WS_PORT = v
	}
	if v := os.Getenv("WEBUI_PORT"); v != "" {
		WEBUI_PORT = v
	}
	if v := os.Getenv("STATS_FILE"); v != "" {
		STATS_FILE = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		REDIS_ADDR = v
	}
	if v := os.Getenv("REDIS_PWD"); v != "" {
		REDIS_PWD = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		JWT_SECRET = v
	}
	if v := os.Getenv("DEFAULT_ADMIN_PASSWORD"); v != "" {
		DEFAULT_ADMIN_PASSWORD = v
	}
	if v := os.Getenv("PG_HOST"); v != "" {
		PG_HOST = v
	}
	if v := os.Getenv("PG_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &PG_PORT)
	}
	if v := os.Getenv("PG_USER"); v != "" {
		PG_USER = v
	}
	if v := os.Getenv("PG_PASSWORD"); v != "" {
		PG_PASSWORD = v
	}
	if v := os.Getenv("PG_DBNAME"); v != "" {
		PG_DBNAME = v
	}
	if v := os.Getenv("PG_SSLMODE"); v != "" {
		PG_SSLMODE = v
	}
	if v := os.Getenv("ENABLE_SKILL"); v != "" {
		ENABLE_SKILL = (v == "true" || v == "1")
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		LOG_LEVEL = v
	}
	if v := os.Getenv("AUTO_REPLY"); v != "" {
		AUTO_REPLY = v == "true"
	}
	if v := os.Getenv("AZURE_TRANSLATE_KEY"); v != "" {
		AZURE_TRANSLATE_KEY = v
	}
	if v := os.Getenv("AZURE_TRANSLATE_ENDPOINT"); v != "" {
		AZURE_TRANSLATE_END = v
	}
	if v := os.Getenv("AZURE_TRANSLATE_REGION"); v != "" {
		AZURE_TRANSLATE_REG = v
	}
}

// 同步到 GlobalConfig
	syncToGlobalConfig()
}

func syncToGlobalConfig() {
	GlobalConfig.WSPort = WS_PORT
	GlobalConfig.WebUIPort = WEBUI_PORT
	GlobalConfig.StatsFile = STATS_FILE
	GlobalConfig.RedisAddr = REDIS_ADDR
	GlobalConfig.RedisPwd = REDIS_PWD
	GlobalConfig.JWTSecret = JWT_SECRET
	GlobalConfig.DefaultAdminPassword = DEFAULT_ADMIN_PASSWORD
	GlobalConfig.PGHost = PG_HOST
	GlobalConfig.PGPort = PG_PORT
	GlobalConfig.PGUser = PG_USER
	GlobalConfig.PGPassword = PG_PASSWORD
	GlobalConfig.PGDBName = PG_DBNAME
	GlobalConfig.PGSSLMode = PG_SSLMODE
	GlobalConfig.EnableSkill = ENABLE_SKILL
	GlobalConfig.LogLevel = LOG_LEVEL
	GlobalConfig.AutoReply = AUTO_REPLY
	GlobalConfig.AzureTranslateKey = AZURE_TRANSLATE_KEY
	GlobalConfig.AzureTranslateEndpoint = AZURE_TRANSLATE_END
	GlobalConfig.AzureTranslateRegion = AZURE_TRANSLATE_REG
}

func GetResolvedConfigPath() string {
	absPath, _ := filepath.Abs(resolvedConfigPath)
	return absPath
}

// SaveConfigToFile 保存配置到文件
func SaveConfigToFile() error {
	syncToGlobalConfig()
	data, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Marshal config failed: %v", err)
		return err
	}
	absPath, _ := filepath.Abs(resolvedConfigPath)
	log.Printf("[INFO] Saving config to file: %s", absPath)
	err = os.WriteFile(resolvedConfigPath, data, 0644)
	if err != nil {
		log.Printf("[ERROR] Write config file failed: %v", err)
	} else {
		log.Printf("[INFO] Config file saved successfully. Content length: %d", len(data))
	}
	return err
}

// SaveConfig 保存 Manager 的配置并持久化
func (m *Manager) SaveConfig() error {
	if m.Config != nil {
		WS_PORT = m.Config.WSPort
		WEBUI_PORT = m.Config.WebUIPort
		STATS_FILE = m.Config.StatsFile
		REDIS_ADDR = m.Config.RedisAddr
		REDIS_PWD = m.Config.RedisPwd
		JWT_SECRET = m.Config.JWTSecret
		DEFAULT_ADMIN_PASSWORD = m.Config.DefaultAdminPassword
		PG_HOST = m.Config.PGHost
		PG_PORT = m.Config.PGPort
		PG_USER = m.Config.PGUser
		PG_PASSWORD = m.Config.PGPassword
		PG_DBNAME = m.Config.PGDBName
		PG_SSLMODE = m.Config.PGSSLMode
		ENABLE_SKILL = m.Config.EnableSkill
		LOG_LEVEL = m.Config.LogLevel
		AUTO_REPLY = m.Config.AutoReply
	}
	return SaveConfigToFile()
}
