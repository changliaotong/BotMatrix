package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"botmatrix/wxbotgo/core"

	"github.com/gin-gonic/gin"
)

// Config represents the bot configuration
type Config struct {
	Networks   []NetworkConfig   `json:"networks"`
	HTTPs      []HTTPConfig      `json:"https"`
	WebSockets []WebSocketConfig `json:"websockets"`
	Logging    LoggingConfig     `json:"logging"`
	Features   FeaturesConfig    `json:"features"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ManagerUrl string `json:"manager_url"`
	SelfId     string `json:"self_id"`
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
}

// WebSocketConfig WebSocket 服务器配置
type WebSocketConfig struct {
	Enabled bool   `json:"enabled"`
	Port    string `json:"port"`
	Host    string `json:"host"`
	Path    string `json:"path"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// FeaturesConfig 功能配置
type FeaturesConfig struct {
	AutoLogin     bool `json:"auto_login"`
	QRCodeSave    bool `json:"qr_code_save"`
	AutoReconnect bool `json:"auto_reconnect"`
}

// LoadConfig reads configuration from config.json
func LoadConfig() (*Config, error) {
	configPath := filepath.Join(".", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Override with environment variables if present
	if envManagerUrl := os.Getenv("MANAGER_URL"); envManagerUrl != "" && len(config.Networks) > 0 {
		config.Networks[0].ManagerUrl = envManagerUrl
	}
	if envSelfId := os.Getenv("BOT_SELF_ID"); envSelfId != "" && len(config.Networks) > 0 {
		config.Networks[0].SelfId = envSelfId
	}
	if envHTTPPort := os.Getenv("HTTP_PORT"); envHTTPPort != "" && len(config.HTTPs) > 0 {
		config.HTTPs[0].Port = envHTTPPort
	}
	if envWSPort := os.Getenv("WS_PORT"); envWSPort != "" && len(config.WebSockets) > 0 {
		config.WebSockets[0].Port = envWSPort
	}

	// Set default values
	if len(config.Networks) == 0 {
		config.Networks = append(config.Networks, NetworkConfig{
			ManagerUrl: "ws://localhost:3001",
			SelfId:     "", // Will be set by server
		})
	}
	if len(config.HTTPs) == 0 {
		config.HTTPs = append(config.HTTPs, HTTPConfig{
			Enabled: true,
			Port:    "8080",
			Host:    "0.0.0.0",
		})
	}
	if len(config.WebSockets) == 0 {
		config.WebSockets = append(config.WebSockets, WebSocketConfig{
			Enabled: true,
			Port:    "3001",
			Host:    "0.0.0.0",
			Path:    "/ws",
		})
	}

	return &config, nil
}

// SaveConfig writes configuration to config.json
func SaveConfig(config *Config) error {
	configPath := filepath.Join(".", "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Console Callback
type ConsoleCallback struct{}

func (c *ConsoleCallback) OnLog(msg string) {
	fmt.Println(msg)
}

func (c *ConsoleCallback) OnQrCode(urlContent string) {
	fmt.Printf("Scan QR Code Link: %s\n", urlContent)

	// urlContent is like https://login.weixin.qq.com/l/IsOwU-8wNA==
	// Image URL is https://login.weixin.qq.com/qrcode/IsOwU-8wNA==

	parts := strings.Split(urlContent, "/l/")
	if len(parts) != 2 {
		fmt.Println("Could not parse UUID from URL to download image.")
		return
	}
	uuid := parts[1]
	imgUrl := "https://login.weixin.qq.com/qrcode/" + uuid

	fmt.Printf("Downloading QR Code image from %s ...\n", imgUrl)
	resp, err := http.Get(imgUrl)
	if err != nil {
		fmt.Printf("Failed to download QR code: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fileName := "qrcode.png"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save QR code: %v\n", err)
		return
	}

	path, _ := os.Getwd()
	fullPath := fmt.Sprintf("%s%c%s", path, os.PathSeparator, fileName)
	fmt.Printf("\n[SUCCESS] QR Code saved to: %s\n", fullPath)
	fmt.Println("Please open this image file to scan.")

	// Try to open the file automatically
	openFile(fileName)
}

func openFile(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Could not open file automatically: %v\n", err)
	}
}

func main() {
	// Load legacy configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Using default configuration due to error: %v\n", err)
		config = &Config{
			Networks: []NetworkConfig{
				{
					ManagerUrl: "ws://localhost:3001",
					SelfId:     "", // Will be set by server
				},
			},
			HTTPs: []HTTPConfig{
				{
					Enabled: true,
					Port:    "8080",
					Host:    "0.0.0.0",
				},
			},
			WebSockets: []WebSocketConfig{
				{
					Enabled: true,
					Port:    "3001",
					Host:    "0.0.0.0",
					Path:    "/ws",
				},
			},
		}
	}

	// Setup WebUI
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve QR code image
	r.GET("/qrcode", func(c *gin.Context) {
		c.File("qrcode.png")
	})

	// Configuration page
	r.GET("/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "config.html", config)
	})

	// Save configuration
	r.POST("/save", func(c *gin.Context) {
		// Update network configuration
		if len(config.Networks) > 0 {
			config.Networks[0].ManagerUrl = c.PostForm("manager_url")
			config.Networks[0].SelfId = c.PostForm("self_id")
		}

		// Update HTTP configuration
		if len(config.HTTPs) > 0 {
			config.HTTPs[0].Enabled = c.PostForm("http_enabled") == "on"
			config.HTTPs[0].Host = c.PostForm("http_host")
			config.HTTPs[0].Port = c.PostForm("http_port")
		}

		// Update WebSocket configuration
		if len(config.WebSockets) > 0 {
			config.WebSockets[0].Enabled = c.PostForm("ws_enabled") == "on"
			config.WebSockets[0].Host = c.PostForm("ws_host")
			config.WebSockets[0].Port = c.PostForm("ws_port")
			config.WebSockets[0].Path = c.PostForm("ws_path")
		}

		// Update logging configuration
		config.Logging.Level = c.PostForm("log_level")
		config.Logging.File = c.PostForm("log_file")

		// Update features configuration
		config.Features.AutoLogin = c.PostForm("auto_login") == "on"
		config.Features.QRCodeSave = c.PostForm("qr_code_save") == "on"
		config.Features.AutoReconnect = c.PostForm("auto_reconnect") == "on"

		// Save configuration to file
		if err := SaveConfig(config); err != nil {
			c.String(http.StatusInternalServerError, "Failed to save configuration: %v", err)
			return
		}

		c.String(http.StatusOK, "Configuration saved successfully!")
	})

	// Start WebUI server in background
	for _, httpConfig := range config.HTTPs {
		if httpConfig.Enabled {
			go func(cfg HTTPConfig) {
				fmt.Printf("Starting HTTP server on %s:%s...\n", cfg.Host, cfg.Port)
				if err := r.Run(cfg.Host + ":" + cfg.Port); err != nil {
					fmt.Printf("Failed to start HTTP server on %s:%s: %v\n", cfg.Host, cfg.Port, err)
				}
			}(httpConfig)
		}
	}

	// Start multiple bot instances for each network configuration
	for _, networkConfig := range config.Networks {
		bot := core.NewWxBot(networkConfig.ManagerUrl, networkConfig.SelfId, &ConsoleCallback{})
		bot.Start()
	}
}
