package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemInfoPlugin 系统信息插件
type SystemInfoPlugin struct {
	cmdParser *CommandParser
	startTime time.Time
}

func (p *SystemInfoPlugin) Name() string {
	return "system_info"
}

func (p *SystemInfoPlugin) Description() string {
	return "系统信息插件，查询服务器硬件、操作系统、软件版本等信息"
}

func (p *SystemInfoPlugin) Version() string {
	return "1.0.0"
}

// NewSystemInfoPlugin 创建系统信息插件实例
func NewSystemInfoPlugin() *SystemInfoPlugin {
	return &SystemInfoPlugin{
		cmdParser: NewCommandParser(),
		startTime: time.Now(),
	}
}

func (p *SystemInfoPlugin) Init(robot plugin.Robot) {
	log.Println("加载系统信息插件")

	// 处理系统信息命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为系统信息命令
		if match, _ := p.cmdParser.MatchCommand("系统信息|systeminfo|sysinfo", event.RawMessage); match {
			// 获取系统信息
			sysInfo := p.GetSystemInfo()
			p.sendMessage(robot, event, sysInfo)
		} else if match, _ := p.cmdParser.MatchCommand("系统信息|systeminfo|sysinfo", event.RawMessage); match {
			// 发送使用说明
			usage := "📊 系统信息命令使用说明:\n"
			usage += "====================\n"
			usage += "/系统信息 - 查看完整系统信息\n"
			usage += "/systeminfo - 查看完整系统信息\n"
			usage += "/sysinfo - 查看完整系统信息\n"
			p.sendMessage(robot, event, usage)
		}

		return nil
	})
}

// sendMessage 发送消息
func (p *SystemInfoPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// GetSystemInfo 获取系统信息
func (p *SystemInfoPlugin) GetSystemInfo() string {
	var info string

	// 基本信息
	info += "📊 系统信息\n"
	info += "====================\n\n"

	// 操作系统信息
	osInfo, err := host.Info()
	if err == nil {
		info += fmt.Sprintf("🖥️  操作系统: %s %s\n", osInfo.OS, osInfo.PlatformVersion)
		info += fmt.Sprintf("🏷️  内核版本: %s\n", osInfo.KernelVersion)
		info += fmt.Sprintf("🏭  主机名: %s\n", osInfo.Hostname)
		info += fmt.Sprintf("⏰  运行时间: %s\n\n", time.Since(p.startTime).Round(time.Second))
	}

	// CPU信息
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		cpuPercent, _ := cpu.Percent(time.Second, false)
		info += fmt.Sprintf("🧠 CPU: %s\n", cpuInfo[0].ModelName)
		info += fmt.Sprintf("⚡ 核心数: %d 物理核心, %d 逻辑核心\n", runtime.NumCPU(), runtime.NumCPU())
		info += fmt.Sprintf("🔥 CPU占用率: %.1f%%\n\n", cpuPercent[0])
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info += fmt.Sprintf("💾 总内存: %.2f GB\n", float64(memInfo.Total)/1024/1024/1024)
		info += fmt.Sprintf("📝 已用内存: %.2f GB (%.1f%%)\n", float64(memInfo.Used)/1024/1024/1024, memInfo.UsedPercent)
		info += fmt.Sprintf("🆓 可用内存: %.2f GB\n\n", float64(memInfo.Available)/1024/1024/1024)
	}

	// 磁盘信息
	diskInfo, err := disk.Usage("/")
	if err == nil {
		info += fmt.Sprintf("💿 磁盘总容量: %.2f GB\n", float64(diskInfo.Total)/1024/1024/1024)
		info += fmt.Sprintf("📂 已用磁盘: %.2f GB (%.1f%%)\n", float64(diskInfo.Used)/1024/1024/1024, diskInfo.UsedPercent)
		info += fmt.Sprintf("🗑️  空闲磁盘: %.2f GB\n\n", float64(diskInfo.Free)/1024/1024/1024)
	}

	// Go版本信息
	info += fmt.Sprintf("🐹 Go版本: %s\n", runtime.Version())
	info += fmt.Sprintf("🏗️  编译架构: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	// 进程信息
	info += fmt.Sprintf("🧵 当前进程ID: %d\n", os.Getpid())
	info += fmt.Sprintf("👥 线程数: %d\n\n", runtime.NumGoroutine())

	info += "====================\n"
	info += "💡 提示: 使用/系统info命令可查看更多详细信息"

	return info
}