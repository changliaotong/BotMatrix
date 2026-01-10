package config

import (
	"BotMatrix/common/bot"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Config 定义应用程序配置结构
type Config struct {
	bot.BotConfig

	// Worker唯一标识
	WorkerID string `json:"worker_id"`

	// HTTP服务器配置
	HTTP HTTPConfig `json:"http"`

	// WebSocket服务器配置
	WebSocket WebSocketConfig `json:"websocket"`

	// 日志配置
	Log LogConfig `json:"log"`

	// 插件配置
	Plugin PluginConfig `json:"plugin"`

	// 数据库配置
	Database DatabaseConfig `json:"database"`

	// Redis配置
	Redis RedisConfig `json:"redis"`

	// 天气API配置
	Weather WeatherConfig `json:"weather"`

	// 翻译API配置
	Translate TranslateConfig `json:"translate"`

	// AI配置
	AI AIConfig `json:"ai"`

	// 技能系统开关
	EnableSkill bool `json:"enable_skill"`
}

// HTTPConfig 定义HTTP服务器配置
type HTTPConfig struct {
	// 监听地址，如 ":8081"
	Addr string `json:"addr"`

	// 读取超时时间
	ReadTimeout time.Duration `json:"read_timeout"`

	// 写入超时时间
	WriteTimeout time.Duration `json:"write_timeout"`
}

// WebSocketConfig 定义WebSocket服务器配置
type WebSocketConfig struct {
	// 监听地址，如 ":8080"
	Addr string `json:"addr"`

	// 读取超时时间
	ReadTimeout time.Duration `json:"read_timeout"`

	// 写入超时时间
	WriteTimeout time.Duration `json:"write_timeout"`

	// Pong超时时间
	PongTimeout time.Duration `json:"pong_timeout"`

	// 检查来源，允许所有来源设置为true
	CheckOrigin bool `json:"check_origin"`
}

// JSONConfig 用于解析JSON的中间结构
type JSONConfig struct {
	LogPort   int    `json:"log_port"`
	BotToken  string `json:"bot_token"`
	NexusAddr string `json:"nexus_addr"`
	WorkerID  string `json:"worker_id"`

	HTTP struct {
		Addr         string      `json:"addr"`
		ReadTimeout  interface{} `json:"read_timeout"`
		WriteTimeout interface{} `json:"write_timeout"`
	} `json:"http"`

	WebSocket struct {
		Addr         string      `json:"addr"`
		ReadTimeout  interface{} `json:"read_timeout"`
		WriteTimeout interface{} `json:"write_timeout"`
		PongTimeout  interface{} `json:"pong_timeout"`
		CheckOrigin  bool        `json:"check_origin"`
	} `json:"websocket"`

	Log struct {
		Level string `json:"level"`
		File  string `json:"file"`
	} `json:"log"`

	Plugin struct {
		Dir     string   `json:"dir"`
		DevDirs []string `json:"dev_dirs"` // 新增：开发目录列表
		Enabled []string `json:"enabled"`
	} `json:"plugin"`

	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`

	Weather struct {
		APIKey   string      `json:"api_key"`
		Endpoint string      `json:"endpoint"`
		Timeout  interface{} `json:"timeout"`
		Mock     bool        `json:"mock"`
	} `json:"weather"`

	Translate struct {
		APIKey   string      `json:"api_key"`
		Endpoint string      `json:"endpoint"`
		Timeout  interface{} `json:"timeout"`
		Region   string      `json:"region"`
	} `json:"translate"`

	AI struct {
		APIKey          string      `json:"api_key"`
		Endpoint        string      `json:"endpoint"`
		Model           string      `json:"model"`
		Timeout         interface{} `json:"timeout"`
		OfficialGroupID string      `json:"official_group_id"`
	} `json:"ai"`

	EnableSkill bool `json:"enable_skill"`
}

// LogConfig 定义日志配置
type LogConfig struct {
	// 日志级别：debug, info, warn, error
	Level string `json:"level"`

	// 日志文件路径，为空则输出到控制台
	File string `json:"file"`
}

// PluginConfig 定义插件配置
type PluginConfig struct {
	// 插件目录
	Dir string `json:"dir"`

	// 开发目录列表
	DevDirs []string `json:"dev_dirs"`

	// 启用的插件列表
	Enabled []string `json:"enabled"`
}

// DatabaseConfig 定义数据库配置
type DatabaseConfig struct {
	// 数据库服务器地址
	Host string `json:"host"`

	// 数据库服务器端口
	Port int `json:"port"`

	// 数据库用户名
	User string `json:"user"`

	// 数据库密码
	Password string `json:"password"`

	// 数据库名称
	DBName string `json:"dbname"`

	// SSL连接模式
	SSLMode string `json:"sslmode"`
}

// RedisConfig 定义Redis配置
type RedisConfig struct {
	// Redis服务器地址
	Host string `json:"host"`

	// Redis服务器端口
	Port int `json:"port"`

	// Redis密码
	Password string `json:"password"`

	// Redis数据库编号
	DB int `json:"db"`

	// Stream配置
	Stream StreamConfig `json:"stream"`
}

// StreamConfig 定义Redis Stream配置
type StreamConfig struct {
	// 要消费的 Stream 列表
	Streams []string `json:"streams"`
	// 消费者组名称
	Group string `json:"group"`
	// 消费者名称
	Consumer string `json:"consumer"`
	// 每次读取的消息数量
	BatchSize int64 `json:"batch_size"`
	// 阻塞读取时间
	BlockTime time.Duration `json:"block_time"`
}

// WeatherConfig 定义天气API配置
type WeatherConfig struct {
	// 天气API密钥
	APIKey string `json:"api_key"`

	// 天气API端点
	Endpoint string `json:"endpoint"`

	// 天气API超时时间
	Timeout time.Duration `json:"timeout"`

	// 是否启用模拟数据
	Mock bool `json:"mock"`
}

// TranslateConfig 定义翻译API配置
type TranslateConfig struct {
	// 翻译API密钥
	APIKey string `json:"api_key"`

	// 翻译API端点
	Endpoint string `json:"endpoint"`

	// 翻译API超时时间
	Timeout time.Duration `json:"timeout"`

	// 翻译API区域
	Region string `json:"region"`
}

// AIConfig 定义AI问答配置
type AIConfig struct {
	APIKey          string        `json:"api_key"`
	Endpoint        string        `json:"endpoint"`
	Model           string        `json:"model"`
	Timeout         time.Duration `json:"timeout"`
	OfficialGroupID string        `json:"official_group_id"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Addr:         ":8081",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		WebSocket: WebSocketConfig{
			Addr:         ":8080",
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 10 * time.Second,
			PongTimeout:  60 * time.Second,
			CheckOrigin:  true,
		},
		Log: LogConfig{
			Level: "info",
			File:  "",
		},
		Plugin: PluginConfig{
			Dir:     "plugins/worker",
			Enabled: []string{},
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			DBName:   "bot",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Weather: WeatherConfig{
			APIKey:   "",
			Endpoint: "https://api.openweathermap.org/data/2.5/weather",
			Timeout:  10 * time.Second,
			Mock:     false,
		},
		Translate: TranslateConfig{
			APIKey:   "",
			Endpoint: "https://api.cognitive.microsofttranslator.com/translate",
			Timeout:  10 * time.Second,
			Region:   "eastus",
		},
		AI: AIConfig{
			APIKey:          "",
			Endpoint:        "",
			Model:           "",
			Timeout:         15 * time.Second,
			OfficialGroupID: "",
		},
		BotConfig: bot.BotConfig{
			LogPort: 8082,
		},
		EnableSkill: false, // 默认关闭技能系统
	}
}

// LoadConfig 加载配置文件，如果文件不存在则使用默认配置
func LoadConfig(configPath string) (*Config, error) {
	// 获取默认配置
	config := DefaultConfig()

	// 如果配置文件路径为空，直接返回默认配置
	if configPath == "" {
		return config, nil
	}

	// 打开配置文件
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件: %w", err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %w", err)
	}

	// 先解析到 JSONConfig 中间结构
	var jsonCfg JSONConfig
	if err := json.Unmarshal(content, &jsonCfg); err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %w", err)
	}

	// 使用通用的更新逻辑
	return UpdateConfigFromJSON(config, content)
}

func UpdateConfigFromJSON(config *Config, content []byte) (*Config, error) {
	// 先解析到 JSONConfig 中间结构
	var jsonCfg JSONConfig
	if err := json.Unmarshal(content, &jsonCfg); err != nil {
		return nil, err
	}

	// 更新 WorkerID
	if jsonCfg.WorkerID != "" {
		config.WorkerID = jsonCfg.WorkerID
	}

	// 更新HTTP配置
	if jsonCfg.HTTP.Addr != "" {
		config.HTTP.Addr = jsonCfg.HTTP.Addr
	}
	if d, ok := parseDuration(jsonCfg.HTTP.ReadTimeout); ok {
		config.HTTP.ReadTimeout = d
	}
	if d, ok := parseDuration(jsonCfg.HTTP.WriteTimeout); ok {
		config.HTTP.WriteTimeout = d
	}

	// 更新WebSocket配置
	if jsonCfg.WebSocket.Addr != "" {
		config.WebSocket.Addr = jsonCfg.WebSocket.Addr
	}
	if d, ok := parseDuration(jsonCfg.WebSocket.ReadTimeout); ok {
		config.WebSocket.ReadTimeout = d
	}
	if d, ok := parseDuration(jsonCfg.WebSocket.WriteTimeout); ok {
		config.WebSocket.WriteTimeout = d
	}
	if d, ok := parseDuration(jsonCfg.WebSocket.PongTimeout); ok {
		config.WebSocket.PongTimeout = d
	}

	// 更新CheckOrigin配置
	config.WebSocket.CheckOrigin = jsonCfg.WebSocket.CheckOrigin

	// 更新日志配置
	if jsonCfg.Log.Level != "" {
		config.Log.Level = jsonCfg.Log.Level
	}
	if jsonCfg.Log.File != "" {
		config.Log.File = jsonCfg.Log.File
	}

	// 更新插件配置
	if jsonCfg.Plugin.Dir != "" {
		config.Plugin.Dir = jsonCfg.Plugin.Dir
	}
	if len(jsonCfg.Plugin.Enabled) > 0 {
		config.Plugin.Enabled = jsonCfg.Plugin.Enabled
	}
	if len(jsonCfg.Plugin.DevDirs) > 0 {
		config.Plugin.DevDirs = jsonCfg.Plugin.DevDirs
	}

	// 更新数据库配置
	if jsonCfg.Database.Host != "" {
		config.Database.Host = jsonCfg.Database.Host
	}
	if jsonCfg.Database.Port != 0 {
		config.Database.Port = jsonCfg.Database.Port
	}
	if jsonCfg.Database.User != "" {
		config.Database.User = jsonCfg.Database.User
	}
	if jsonCfg.Database.Password != "" {
		config.Database.Password = jsonCfg.Database.Password
	}
	if jsonCfg.Database.DBName != "" {
		config.Database.DBName = jsonCfg.Database.DBName
	}
	if jsonCfg.Database.SSLMode != "" {
		config.Database.SSLMode = jsonCfg.Database.SSLMode
	}

	// 更新Redis配置
	if jsonCfg.Redis.Host != "" {
		config.Redis.Host = jsonCfg.Redis.Host
	}
	if jsonCfg.Redis.Port != 0 {
		config.Redis.Port = jsonCfg.Redis.Port
	}
	if jsonCfg.Redis.Password != "" {
		config.Redis.Password = jsonCfg.Redis.Password
	}
	config.Redis.DB = jsonCfg.Redis.DB

	config.Redis.Stream.Streams = jsonCfg.Redis.Stream.Streams
	config.Redis.Stream.Group = jsonCfg.Redis.Stream.Group
	config.Redis.Stream.Consumer = jsonCfg.Redis.Stream.Consumer
	config.Redis.Stream.BatchSize = jsonCfg.Redis.Stream.BatchSize
	if config.Redis.Stream.BatchSize == 0 {
		config.Redis.Stream.BatchSize = 10 // 默认值
	}

	if d, ok := parseDuration(jsonCfg.Redis.Stream.BlockTime); ok {
		config.Redis.Stream.BlockTime = d
	}
	if config.Redis.Stream.BlockTime == 0 {
		config.Redis.Stream.BlockTime = 2 * time.Second // 默认值
	}

	// 更新天气API配置
	if jsonCfg.Weather.APIKey != "" {
		config.Weather.APIKey = jsonCfg.Weather.APIKey
	}
	if jsonCfg.Weather.Endpoint != "" {
		config.Weather.Endpoint = jsonCfg.Weather.Endpoint
	}
	if d, ok := parseDuration(jsonCfg.Weather.Timeout); ok {
		config.Weather.Timeout = d
	}
	config.Weather.Mock = jsonCfg.Weather.Mock

	// 更新翻译API配置
	if jsonCfg.Translate.APIKey != "" {
		config.Translate.APIKey = jsonCfg.Translate.APIKey
	}
	if jsonCfg.Translate.Endpoint != "" {
		config.Translate.Endpoint = jsonCfg.Translate.Endpoint
	}
	if d, ok := parseDuration(jsonCfg.Translate.Timeout); ok {
		config.Translate.Timeout = d
	}
	if jsonCfg.Translate.Region != "" {
		config.Translate.Region = jsonCfg.Translate.Region
	}

	// 更新AI配置
	if jsonCfg.AI.APIKey != "" {
		config.AI.APIKey = jsonCfg.AI.APIKey
	}
	if jsonCfg.AI.Endpoint != "" {
		config.AI.Endpoint = jsonCfg.AI.Endpoint
	}
	if jsonCfg.AI.Model != "" {
		config.AI.Model = jsonCfg.AI.Model
	}
	if d, ok := parseDuration(jsonCfg.AI.Timeout); ok {
		config.AI.Timeout = d
	}
	if jsonCfg.AI.OfficialGroupID != "" {
		config.AI.OfficialGroupID = jsonCfg.AI.OfficialGroupID
	}

	// 更新 BotConfig 字段
	if jsonCfg.LogPort != 0 {
		config.LogPort = jsonCfg.LogPort
	}
	if jsonCfg.BotToken != "" {
		config.BotToken = jsonCfg.BotToken
	}
	if jsonCfg.NexusAddr != "" {
		config.NexusAddr = jsonCfg.NexusAddr
	}

	// 更新技能开关
	config.EnableSkill = jsonCfg.EnableSkill

	return config, nil
}

// parseDuration 解析 interface{} 类型的持续时间（支持 string 和 float64）
func parseDuration(v interface{}) (time.Duration, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case string:
		if d, err := time.ParseDuration(val); err == nil {
			return d, true
		}
	case float64:
		return time.Duration(val), true
	}
	return 0, false
}

// LoadFromCLI 从命令行参数加载配置
func LoadFromCLI() (*Config, string, error) {
	// 定义命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	httpAddr := flag.String("http-addr", "", "HTTP服务器监听地址")
	wsAddr := flag.String("ws-addr", "", "WebSocket服务器监听地址")
	logLevel := flag.String("log-level", "", "日志级别")
	logPort := flag.Int("log-port", 0, "Web UI 端口")

	// 数据库配置命令行参数
	dbHost := flag.String("db-host", "", "数据库服务器地址")
	dbPort := flag.Int("db-port", 0, "数据库服务器端口")
	dbUser := flag.String("db-user", "", "数据库用户名")
	dbPassword := flag.String("db-password", "", "数据库密码")
	dbName := flag.String("db-name", "", "数据库名称")
	dbSSLMode := flag.String("db-sslmode", "", "数据库SSL连接模式")

	// Redis配置命令行参数
	redisHost := flag.String("redis-host", "", "Redis服务器地址")
	redisPort := flag.Int("redis-port", 0, "Redis服务器端口")
	redisPassword := flag.String("redis-password", "", "Redis密码")
	redisDB := flag.Int("redis-db", 0, "Redis数据库编号")

	// 天气API配置命令行参数
	weatherAPIKey := flag.String("weather-api-key", "", "天气API密钥")
	weatherEndpoint := flag.String("weather-endpoint", "", "天气API端点")

	// 技能系统命令行参数
	enableSkill := flag.Bool("enable-skill", false, "是否启用技能系统")

	// 解析命令行参数
	flag.Parse()

	// 加载配置文件
	config, err := LoadConfig(*configPath)
	if err != nil {
		return nil, "", err
	}

	// 命令行参数优先级高于配置文件
	if *enableSkill {
		config.EnableSkill = true
	} else if os.Getenv("ENABLE_SKILL") == "true" {
		config.EnableSkill = true
	}
	if *httpAddr != "" {
		config.HTTP.Addr = *httpAddr
	}

	if *wsAddr != "" {
		config.WebSocket.Addr = *wsAddr
	}

	if *logLevel != "" {
		config.Log.Level = *logLevel
	}
	if *logPort != 0 {
		config.LogPort = *logPort
	}

	// 数据库配置命令行参数处理
	if *dbHost != "" {
		config.Database.Host = *dbHost
	}

	if *dbPort != 0 {
		config.Database.Port = *dbPort
	}

	if *dbUser != "" {
		config.Database.User = *dbUser
	}

	if *dbPassword != "" {
		config.Database.Password = *dbPassword
	}

	if *dbName != "" {
		config.Database.DBName = *dbName
	}

	if *dbSSLMode != "" {
		config.Database.SSLMode = *dbSSLMode
	}

	// Redis配置命令行参数处理
	if *redisHost != "" {
		config.Redis.Host = *redisHost
	}

	if *redisPort != 0 {
		config.Redis.Port = *redisPort
	}

	if *redisPassword != "" {
		config.Redis.Password = *redisPassword
	}

	if *redisDB != 0 {
		config.Redis.DB = *redisDB
	}

	// 天气配置命令行参数处理
	if *weatherAPIKey != "" {
		config.Weather.APIKey = *weatherAPIKey
	}
	if *weatherEndpoint != "" {
		config.Weather.Endpoint = *weatherEndpoint
	}

	return config, *configPath, nil
}
