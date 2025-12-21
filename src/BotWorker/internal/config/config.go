package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Config 定义应用程序配置结构
type Config struct {
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

// jsonConfig 用于解析JSON的中间结构
type jsonConfig struct {
	HTTP struct {
		Addr         string `json:"addr"`
		ReadTimeout  string `json:"read_timeout"`
		WriteTimeout string `json:"write_timeout"`
	} `json:"http"`

	WebSocket struct {
		Addr         string `json:"addr"`
		ReadTimeout  string `json:"read_timeout"`
		WriteTimeout string `json:"write_timeout"`
		PongTimeout  string `json:"pong_timeout"`
		CheckOrigin  bool   `json:"check_origin"`
	} `json:"websocket"`

	Log struct {
		Level string `json:"level"`
		File  string `json:"file"`
	} `json:"log"`

	Plugin struct {
		Dir     string   `json:"dir"`
		Enabled []string `json:"enabled"`
	} `json:"plugin"`

	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
	} `json:"database"`

	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`

	Weather struct {
		APIKey   string `json:"api_key"`
		Endpoint string `json:"endpoint"`
		Timeout  string `json:"timeout"`
	} `json:"weather"`

	Translate struct {
		APIKey   string `json:"api_key"`
		Endpoint string `json:"endpoint"`
		Timeout  string `json:"timeout"`
		Region   string `json:"region"`
	} `json:"translate"`
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
}

// WeatherConfig 定义天气API配置
type WeatherConfig struct {
	// 天气API密钥
	APIKey string `json:"api_key"`

	// 天气API端点
	Endpoint string `json:"endpoint"`

	// 天气API超时时间
	Timeout time.Duration `json:"timeout"`
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
			Dir:     "plugins",
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
		},
		Translate: TranslateConfig{
			APIKey:   "",
			Endpoint: "https://api.cognitive.microsofttranslator.com/translate",
			Timeout:  10 * time.Second,
			Region:   "eastus",
		},
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

	// 先解析到jsonConfig中间结构
	var jsonCfg jsonConfig
	if err := json.Unmarshal(content, &jsonCfg); err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %w", err)
	}

	// 更新HTTP配置
	if jsonCfg.HTTP.Addr != "" {
		config.HTTP.Addr = jsonCfg.HTTP.Addr
	}
	if jsonCfg.HTTP.ReadTimeout != "" {
		if readTimeout, err := time.ParseDuration(jsonCfg.HTTP.ReadTimeout); err == nil {
			config.HTTP.ReadTimeout = readTimeout
		}
	}
	if jsonCfg.HTTP.WriteTimeout != "" {
		if writeTimeout, err := time.ParseDuration(jsonCfg.HTTP.WriteTimeout); err == nil {
			config.HTTP.WriteTimeout = writeTimeout
		}
	}

	// 更新WebSocket配置
	if jsonCfg.WebSocket.Addr != "" {
		config.WebSocket.Addr = jsonCfg.WebSocket.Addr
	}
	if jsonCfg.WebSocket.ReadTimeout != "" {
		if readTimeout, err := time.ParseDuration(jsonCfg.WebSocket.ReadTimeout); err == nil {
			config.WebSocket.ReadTimeout = readTimeout
		}
	}
	if jsonCfg.WebSocket.WriteTimeout != "" {
		if writeTimeout, err := time.ParseDuration(jsonCfg.WebSocket.WriteTimeout); err == nil {
			config.WebSocket.WriteTimeout = writeTimeout
		}
	}
	if jsonCfg.WebSocket.PongTimeout != "" {
		if pongTimeout, err := time.ParseDuration(jsonCfg.WebSocket.PongTimeout); err == nil {
			config.WebSocket.PongTimeout = pongTimeout
		}
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
	if jsonCfg.Redis.DB != 0 {
		config.Redis.DB = jsonCfg.Redis.DB
	}

	// 更新天气API配置
	if jsonCfg.Weather.APIKey != "" {
		config.Weather.APIKey = jsonCfg.Weather.APIKey
	}
	if jsonCfg.Weather.Endpoint != "" {
		config.Weather.Endpoint = jsonCfg.Weather.Endpoint
	}
	if jsonCfg.Weather.Timeout != "" {
		if timeout, err := time.ParseDuration(jsonCfg.Weather.Timeout); err == nil {
			config.Weather.Timeout = timeout
		}
	}

	// 更新翻译API配置
	if jsonCfg.Translate.APIKey != "" {
		config.Translate.APIKey = jsonCfg.Translate.APIKey
	}
	if jsonCfg.Translate.Endpoint != "" {
		config.Translate.Endpoint = jsonCfg.Translate.Endpoint
	}
	if jsonCfg.Translate.Timeout != "" {
		if timeout, err := time.ParseDuration(jsonCfg.Translate.Timeout); err == nil {
			config.Translate.Timeout = timeout
		}
	}
	if jsonCfg.Translate.Region != "" {
		config.Translate.Region = jsonCfg.Translate.Region
	}

	return config, nil
}

// LoadFromCLI 从命令行参数加载配置
func LoadFromCLI() (*Config, string, error) {
	// 定义命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	httpAddr := flag.String("http-addr", "", "HTTP服务器监听地址")
	wsAddr := flag.String("ws-addr", "", "WebSocket服务器监听地址")
	logLevel := flag.String("log-level", "", "日志级别")

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

	// 解析命令行参数
	flag.Parse()

	// 加载配置文件
	config, err := LoadConfig(*configPath)
	if err != nil {
		return nil, "", err
	}

	// 命令行参数优先级高于配置文件
	if *httpAddr != "" {
		config.HTTP.Addr = *httpAddr
	}

	if *wsAddr != "" {
		config.WebSocket.Addr = *wsAddr
	}

	if *logLevel != "" {
		config.Log.Level = *logLevel
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
