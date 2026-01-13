package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// GlobalConfig is the global configuration instance
var GlobalConfig = &AppConfig{}

const CONFIG_FILE = "config.json"

// Backward compatibility constants
var (
	WS_PORT      = "0.0.0.0:8080"
	WEBUI_PORT   = "0.0.0.0:5000"
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
	resolvedPath := path
	if resolvedPath == "" {
		// Try current directory
		if _, err := os.Stat(CONFIG_FILE); err == nil {
			resolvedPath = CONFIG_FILE
			log.Printf("[DEBUG] Found config in current dir: %s", resolvedPath)
		} else {
			// Try project root (assuming we are in a subfolder of BotMatrix)
			cwd, _ := os.Getwd()
			log.Printf("[DEBUG] CWD: %s", cwd)
			// Look for BotMatrix root
			rootPath := findProjectRoot(cwd)
			log.Printf("[DEBUG] RootPath: %s", rootPath)
			if rootPath != "" {
				testPath := filepath.Join(rootPath, CONFIG_FILE)
				if _, err := os.Stat(testPath); err == nil {
					resolvedPath = testPath
					log.Printf("[DEBUG] Found config in root: %s", resolvedPath)
				}
			}
		}
	}

	if resolvedPath == "" {
		resolvedPath = CONFIG_FILE // fallback
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

// GetResolvedConfigPath returns the actual path of the config file being used
func GetResolvedConfigPath() string {
	// If GlobalConfig was loaded from a specific file, we should have stored it
	// For now, return the default location logic
	if _, err := os.Stat(CONFIG_FILE); err == nil {
		return CONFIG_FILE
	}
	cwd, _ := os.Getwd()
	rootPath := findProjectRoot(cwd)
	if rootPath != "" {
		testPath := filepath.Join(rootPath, CONFIG_FILE)
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}
	}
	return CONFIG_FILE
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Always save to the default CONFIG_FILE location
	if err := os.WriteFile(CONFIG_FILE, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	log.Printf("Successfully saved config to %s", CONFIG_FILE)
	return nil
}

func findProjectRoot(startDir string) string {
	curr := startDir
	for {
		if _, err := os.Stat(filepath.Join(curr, "go.mod")); err == nil {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
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
		// Map nested database config to flat fields if necessary
		if GlobalConfig.Database.Host != "" {
			if GlobalConfig.PGHost == "" {
				GlobalConfig.PGHost = GlobalConfig.Database.Host
			}
			if GlobalConfig.PGPort == 0 {
				GlobalConfig.PGPort = GlobalConfig.Database.Port
			}
			if GlobalConfig.PGUser == "" {
				GlobalConfig.PGUser = GlobalConfig.Database.User
			}
			if GlobalConfig.PGPassword == "" {
				GlobalConfig.PGPassword = GlobalConfig.Database.Password
			}
			if GlobalConfig.PGDBName == "" {
				GlobalConfig.PGDBName = GlobalConfig.Database.DBName
			}
			if GlobalConfig.PGSSLMode == "" {
				GlobalConfig.PGSSLMode = GlobalConfig.Database.SSLMode
			}
		}
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
	if val := os.Getenv("AI_EMBEDDING_MODEL"); val != "" {
		GlobalConfig.AIEmbeddingModel = val
	}
}
