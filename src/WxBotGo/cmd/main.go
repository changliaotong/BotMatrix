package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"botmatrix/wxbotgo/core"
)

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
	managerUrl := os.Getenv("MANAGER_URL")
	selfId := os.Getenv("BOT_SELF_ID")

	if managerUrl == "" {
		managerUrl = "ws://localhost:3001"
	}
	if selfId == "" {
		selfId = "1098299491"
	}

	bot := core.NewWxBot(managerUrl, selfId, &ConsoleCallback{})
	bot.Start()
}
