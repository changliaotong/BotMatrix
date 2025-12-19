package main

import (
	"encoding/json"
	"log"
	"os"
)

// GlobalConfig 全局配置实例
var GlobalConfig = &AppConfig{
	WSPort:               ":3001",
	WebUIPort:            ":5000",
	StatsFile:            "stats.json",
	RedisAddr:            "192.168.0.126:6379",
	RedisPwd:             "redis_zsYik8",
	JWTSecret:            "botnexus_secret_key_for_jwt_token_generation",
	DefaultAdminPassword: "admin123",
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
)

const CONFIG_FILE = "config.json"

func init() {
	// 1. 设置默认值
	WS_PORT = GlobalConfig.WSPort
	WEBUI_PORT = GlobalConfig.WebUIPort
	STATS_FILE = GlobalConfig.StatsFile
	REDIS_ADDR = GlobalConfig.RedisAddr
	REDIS_PWD = GlobalConfig.RedisPwd
	JWT_SECRET = GlobalConfig.JWTSecret
	DEFAULT_ADMIN_PASSWORD = GlobalConfig.DefaultAdminPassword

	// 2. 尝试从文件加载
	loadConfigFromFile()

	// 3. 环境变量覆盖 (最高优先级)
	loadConfigFromEnv()

	// 4. 同步回 GlobalConfig
	syncToGlobalConfig()
}

func loadConfigFromFile() {
	if _, err := os.Stat(CONFIG_FILE); err == nil {
		data, err := os.ReadFile(CONFIG_FILE)
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
		log.Printf("[INFO] 已从 %s 加载配置", CONFIG_FILE)
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
}

func syncToGlobalConfig() {
	GlobalConfig.WSPort = WS_PORT
	GlobalConfig.WebUIPort = WEBUI_PORT
	GlobalConfig.StatsFile = STATS_FILE
	GlobalConfig.RedisAddr = REDIS_ADDR
	GlobalConfig.RedisPwd = REDIS_PWD
	GlobalConfig.JWTSecret = JWT_SECRET
	GlobalConfig.DefaultAdminPassword = DEFAULT_ADMIN_PASSWORD
}

func saveConfigToFile() error {
	syncToGlobalConfig()
	data, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CONFIG_FILE, data, 0644)
}
