package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"botmatrix/wxbotgo/core"
)

// Config represents the bot configuration
type Config struct {
	Networks []NetworkConfig `json:"networks"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ManagerUrl    string `json:"manager_url"`
	SelfId        string `json:"self_id"`
	ReportSelfMsg bool   `json:"report_self_msg"` // 是否上报自身消息
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

	// Set default values for each network configuration
	for i := range config.Networks {
		// Set default ManagerUrl if empty
		if config.Networks[i].ManagerUrl == "" {
			config.Networks[i].ManagerUrl = "ws://localhost:3001"
		}
		// Set default ReportSelfMsg to true if not specified
		// config.Networks[i].ReportSelfMsg is already false by default in Go
	}

	// Override with environment variables if present
	if envManagerUrl := os.Getenv("MANAGER_URL"); envManagerUrl != "" && len(config.Networks) > 0 {
		config.Networks[0].ManagerUrl = envManagerUrl
	}
	if envSelfId := os.Getenv("BOT_SELF_ID"); envSelfId != "" && len(config.Networks) > 0 {
		config.Networks[0].SelfId = envSelfId
	}

	if envReportSelfMsg := os.Getenv("REPORT_SELF_MSG"); envReportSelfMsg != "" && len(config.Networks) > 0 {
		report, _ := strconv.ParseBool(envReportSelfMsg)
		config.Networks[0].ReportSelfMsg = report
	}

	// Set default values if no networks configured
	if len(config.Networks) == 0 {
		config.Networks = append(config.Networks, NetworkConfig{
			ManagerUrl:    "ws://localhost:3001",
			SelfId:        "", // Will be set by server
			ReportSelfMsg: true,
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
	fmt.Printf("\n========================================\n")
	fmt.Printf("        WeChat Bot Login Required        \n")
	fmt.Printf("========================================\n\n")

	// 1. 显示二维码链接，方便用户复制或使用浏览器打开
	fmt.Printf("📱 QR Code Login Link: %s\n\n", urlContent)
	fmt.Printf("💡 Tip: You can copy this link to your browser to open the QR code.\n\n")

	// urlContent is like https://login.weixin.qq.com/l/IsOwU-8wNA==
	// Image URL is https://login.weixin.qq.com/qrcode/IsOwU-8wNA==

	parts := strings.Split(urlContent, "/l/")
	if len(parts) != 2 {
		fmt.Println("❌ Could not parse UUID from URL to download image.")
		fmt.Println("ℹ️  Please use the QR code link above to scan.")
		fmt.Printf("========================================\n\n")
		return
	}
	uuid := parts[1]
	imgUrl := "https://login.weixin.qq.com/qrcode/" + uuid

	// 2. 下载并保存二维码图片
	fmt.Printf("📥 Downloading QR Code image ...\n")
	resp, err := http.Get(imgUrl)
	if err != nil {
		fmt.Printf("❌ Failed to download QR code: %v\n", err)
		fmt.Println("ℹ️  Please use the QR code link above to scan.")
		fmt.Printf("========================================\n\n")
		return
	}
	defer resp.Body.Close()

	fileName := "qrcode.png"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		fmt.Println("ℹ️  Please use the QR code link above to scan.")
		fmt.Printf("========================================\n\n")
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to save QR code: %v\n", err)
		fmt.Println("ℹ️  Please use the QR code link above to scan.")
		fmt.Printf("========================================\n\n")
		return
	}

	path, _ := os.Getwd()
	fullPath := fmt.Sprintf("%s%c%s", path, os.PathSeparator, fileName)
	fmt.Printf("\n✅ QR Code saved to: %s\n", fullPath)
	fmt.Println("📤 The image file is being opened automatically...")
	fmt.Println()
	fmt.Println("📝 Please open the QR code image and scan it with WeChat to log in.")
	fmt.Printf("========================================\n\n")

	// Try to open the file automatically
	openFile(fileName)
}

func openFile(filePath string) {
	var err error
	openSuccess := false
	var openMethod string

	// 尝试自动打开图片文件
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", filePath).Start()
		openSuccess = err == nil
		openMethod = "Linux default app"
	case "windows":
		// 在Windows下，尝试多种方法打开图片文件
		// 方法1：尝试使用画图工具打开，因为画图工具肯定能打开图片
		fmt.Println("💡 Trying to open with Paint...")
		err = exec.Command("cmd", "/c", "start", "", "/min", "mspaint.exe", filePath).Start()
		if err == nil {
			openSuccess = true
			openMethod = "Paint"
		} else {
			// 方法2：尝试使用默认程序打开
			fmt.Println("💡 Trying to open with default app...")
			err = exec.Command("cmd", "/c", "start", "", "/min", filePath).Start()
			openSuccess = err == nil
			openMethod = "Default app"
		}
	case "darwin":
		err = exec.Command("open", filePath).Start()
		openSuccess = err == nil
		openMethod = "macOS default app"
	default:
		err = fmt.Errorf("unsupported platform")
		openSuccess = false
		openMethod = "Unknown"
	}

	// 无论是否成功，都显示清晰的提示
	fmt.Printf("\n📌 QR Code Image Location: %s\n", filePath)

	if openSuccess {
		// 如果自动打开成功，提示用户扫码，并显示使用了哪种方法
		fmt.Printf("✅ Image file opened automatically with %s. Please scan the QR code.\n", openMethod)
		fmt.Println("💡 If you don't see the image, check your taskbar for minimized windows.")
	} else {
		// 如果自动打开失败，提供手动操作指导
		fmt.Println("❌ Auto-open failed. Please try one of these methods:")
		fmt.Println("   1. Open the file manually with an image viewer")
		fmt.Println("   2. Copy the QR code link to your browser")
		fmt.Println("   3. Use Windows Paint: Start → Paint → File → Open")
	}

	fmt.Println("📝 Remember to confirm login on your WeChat app after scanning!")
}

func main() {
	// Parse command line arguments
	managerUrl := flag.String("manager-url", "", "Bot manager WebSocket URL")
	selfId := flag.String("self-id", "", "Bot self ID")
	reportSelfMsg := flag.Bool("report-self-msg", true, "Report self messages to server")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Using default configuration due to error: %v\n", err)
		config = &Config{
			Networks: []NetworkConfig{
				{
					ManagerUrl:    "ws://localhost:3001",
					SelfId:        "", // Will be set by server
					ReportSelfMsg: *reportSelfMsg,
				},
			},
		}
	}

	// Override with command line arguments if provided
	for i := range config.Networks {
		if *managerUrl != "" {
			config.Networks[i].ManagerUrl = *managerUrl
		}
		if *selfId != "" {
			config.Networks[i].SelfId = *selfId
		}
		// Command line arguments override config file
		config.Networks[i].ReportSelfMsg = *reportSelfMsg
	}

	// Start multiple bot instances for each network configuration
	for _, networkConfig := range config.Networks {
		bot := core.NewWxBot(networkConfig.ManagerUrl, networkConfig.SelfId, &ConsoleCallback{})
		bot.ReportSelfMsg = networkConfig.ReportSelfMsg
		bot.Start()
	}
}
