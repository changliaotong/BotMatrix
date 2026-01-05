package config

import (
	"encoding/json"
	"log"
	"os"
)

// GlobalConfig is the global configuration instance
var GlobalConfig = &AppConfig{}

const CONFIG_FILE = "config.json"

// Backward compatibility constants
var (
	WS_PORT      = ":8080"
	WEBUI_PORT   = ":5000"
	ENABLE_SKILL = true
	REDIS_ADDR   = "localhost:6379"
	REDIS_PWD    = ""
)

// Redis Key constants
const (
	REDIS_KEY_QUEUE_DEFAULT    = "botmatrix:queue:default"
	REDIS_KEY_QUEUE_WORKER     = "botmatrix:queue:worker:%s"
	REDIS_KEY_RATELIMIT_USER   = "botmatrix:ratelimit:user:%s"
	REDIS_KEY_RATELIMIT_GROUP  = "botmatrix:ratelimit:group:%s"
	REDIS_KEY_IDEMPOTENCY      = "botmatrix:msg:idempotency:%s"
	REDIS_KEY_SESSION_CONTEXT  = "botmatrix:session:%s:%s" // platform:user_id
	REDIS_KEY_DYNAMIC_RULES    = "botmatrix:rules:routing"
	REDIS_KEY_ACTION_QUEUE     = "botmatrix:actions"
	REDIS_KEY_CONFIG_RATELIMIT = "botmatrix:config:ratelimit"
	REDIS_KEY_CONFIG_TTL       = "botmatrix:config:ttl"
)

// InitConfig initializes the global configuration
func InitConfig(path string) error {
	resolvedPath := CONFIG_FILE
	if path != "" {
		resolvedPath = path
	}

	loadConfigFromFile(resolvedPath)
	loadConfigFromEnv()

	// Sync backward compatibility constants
	if GlobalConfig.WSPort != "" {
		WS_PORT = GlobalConfig.WSPort
	}
	if GlobalConfig.WebUIPort != "" {
		WEBUI_PORT = GlobalConfig.WebUIPort
	}
	ENABLE_SKILL = GlobalConfig.EnableSkill
	if GlobalConfig.RedisAddr != "" {
		REDIS_ADDR = GlobalConfig.RedisAddr
	}
	if GlobalConfig.RedisPwd != "" {
		REDIS_PWD = GlobalConfig.RedisPwd
	}

	return nil
}

func loadConfigFromFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: Config file %s not found, using default/env values", path)
		return
	}

	if err := json.Unmarshal(data, GlobalConfig); err != nil {
		log.Printf("Error: Failed to parse config file: %v", err)
	} else {
		log.Printf("Successfully loaded config from %s", path)
	}
}

func loadConfigFromEnv() {
	if val := os.Getenv("WS_PORT"); val != "" {
		GlobalConfig.WSPort = val
	}
	if val := os.Getenv("WEBUI_PORT"); val != "" {
		GlobalConfig.WebUIPort = val
	}
	if val := os.Getenv("REDIS_ADDR"); val != "" {
		GlobalConfig.RedisAddr = val
	}
	if val := os.Getenv("REDIS_PWD"); val != "" {
		GlobalConfig.RedisPwd = val
	}
	if val := os.Getenv("JWT_SECRET"); val != "" {
		GlobalConfig.JWTSecret = val
	}
	if val := os.Getenv("PG_HOST"); val != "" {
		GlobalConfig.PGHost = val
	}
	// ... add other env vars as needed
}

// GetResolvedConfigPath returns the absolute path to the config file
func GetResolvedConfigPath() string {
	return CONFIG_FILE
}

// SaveConfig persists configuration to disk
func SaveConfig(cfg *AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CONFIG_FILE, data, 0644)
}
