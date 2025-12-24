package plugins

import (
	"BotMatrix/common"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
	return common.T("", "sysinfo_plugin_desc|系统信息插件，查询服务器硬件、操作系统、软件版本等信息")
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
	log.Println(common.T("", "sysinfo_plugin_loaded|加载系统信息插件"))

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
		} else if match, _ := p.cmdParser.MatchCommand("系统信息帮助", event.RawMessage); match {
			// 发送使用说明
			var sb strings.Builder
			sb.WriteString(common.T("", "sysinfo_usage_header|📊 系统信息命令使用说明:\n"))
			sb.WriteString("====================\n")
			sb.WriteString(common.T("", "sysinfo_usage_cmd1|/系统信息 - 查看完整系统信息\n"))
			sb.WriteString(common.T("", "sysinfo_usage_cmd2|/systeminfo - 查看完整系统信息\n"))
			sb.WriteString(common.T("", "sysinfo_usage_cmd3|/sysinfo - 查看完整系统信息\n"))
			p.sendMessage(robot, event, sb.String())
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
	var info strings.Builder

	// 基本信息
	info.WriteString(common.T("", "sysinfo_header|📊 系统信息\n"))
	info.WriteString("====================\n\n")

	// 操作系统信息
	osInfo, err := host.Info()
	if err == nil {
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_os|🖥️  操作系统: %s %s\n"), osInfo.OS, osInfo.PlatformVersion))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_kernel|🏷️  内核版本: %s\n"), osInfo.KernelVersion))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_hostname|🏭  主机名: %s\n"), osInfo.Hostname))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_uptime|⏰  运行时间: %s\n\n"), time.Since(p.startTime).Round(time.Second)))
	}

	// CPU信息
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		cpuPercent, _ := cpu.Percent(time.Second, false)
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_cpu_model|🧠 CPU: %s\n"), cpuInfo[0].ModelName))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_cpu_cores|⚡ 核心数: %d 物理核心, %d 逻辑核心\n"), runtime.NumCPU(), runtime.NumCPU()))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_cpu_usage|🔥 CPU占用率: %.1f%%\n\n"), cpuPercent[0]))
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_mem_total|💾 总内存: %.2f GB\n"), float64(memInfo.Total)/1024/1024/1024))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_mem_used|📝 已用内存: %.2f GB (%.1f%%)\n"), float64(memInfo.Used)/1024/1024/1024, memInfo.UsedPercent))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_mem_avail|🆓 可用内存: %.2f GB\n\n"), float64(memInfo.Available)/1024/1024/1024))
	}

	// 磁盘信息
	diskInfo, err := disk.Usage("/")
	if err == nil {
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_disk_total|💿 磁盘总容量: %.2f GB\n"), float64(diskInfo.Total)/1024/1024/1024))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_disk_used|📂 已用磁盘: %.2f GB (%.1f%%)\n"), float64(diskInfo.Used)/1024/1024/1024, diskInfo.UsedPercent))
		info.WriteString(fmt.Sprintf(common.T("", "sysinfo_disk_free|🗑️  空闲磁盘: %.2f GB\n\n"), float64(diskInfo.Free)/1024/1024/1024))
	}

	// Go版本信息
	info.WriteString(fmt.Sprintf(common.T("", "sysinfo_go_ver|🐹 Go版本: %s\n"), runtime.Version()))
	info.WriteString(fmt.Sprintf(common.T("", "sysinfo_arch|🏗️  编译架构: %s/%s\n\n"), runtime.GOOS, runtime.GOARCH))

	// 进程信息
	info.WriteString(fmt.Sprintf(common.T("", "sysinfo_pid|🧵 当前进程ID: %d\n"), os.Getpid()))
	info.WriteString(fmt.Sprintf(common.T("", "sysinfo_goroutines|👥 线程数: %d\n\n"), runtime.NumGoroutine()))

	info.WriteString("====================\n")
	info.WriteString(common.T("", "sysinfo_footer|💡 提示: 使用/系统info命令可查看更多详细信息"))

	return info.String()
}