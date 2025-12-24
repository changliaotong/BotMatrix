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
)

// Config represents the bot configuration
type Config struct {
	Networks []NetworkConfig `json:"networks"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ManagerUrl string `json:"manager_url"`
	SelfId     string `json:"self_id"`
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

	// Set default values
	if len(config.Networks) == 0 {
		config.Networks = append(config.Networks, NetworkConfig{
			ManagerUrl: "ws://localhost:3001",
			SelfId:     "", // Will be set by server
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
		}
	}

	// Start multiple bot instances for each network configuration
	for _, networkConfig := range config.Networks {
		bot := core.NewWxBot(networkConfig.ManagerUrl, networkConfig.SelfId, &ConsoleCallback{})
		bot.Start()
	}
}
