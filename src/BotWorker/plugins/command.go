package plugins

import (
	"regexp"
	"strings"
)

// CommandParser 命令解析器结构体
type CommandParser struct{}

// NewCommandParser 创建命令解析器实例
func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

// MatchCommand 匹配无参数命令
// pattern: 命令模式，如 "(points|积分)"
// message: 输入消息
// 返回值: 是否匹配，命令名称
func (p *CommandParser) MatchCommand(pattern string, message string) (bool, string) {
	// 构建正则表达式：支持可选/前缀和任意空格
	regPattern := regexp.MustCompile(`^(?:/\s*)?(` + pattern + `)\s*$`)
	matches := regPattern.FindStringSubmatch(message)
	if len(matches) < 2 {
		return false, ""
	}
	return true, matches[1]
}

// MatchCommandWithParams 匹配带参数命令
// pattern: 命令模式，如 "(打赏|reward)"
// paramPattern: 参数模式，如 "(\\S+)\\s+(\\S+)"
// message: 输入消息
// 返回值: 是否匹配，命令名称，参数列表
func (p *CommandParser) MatchCommandWithParams(pattern string, paramPattern string, message string) (bool, string, []string) {
	// 构建正则表达式：支持可选/前缀和任意空格
	regPattern := regexp.MustCompile(`^(?:/\s*)?(` + pattern + `)\s+` + paramPattern + `\s*$`)
	matches := regPattern.FindStringSubmatch(message)
	if len(matches) < 2 {
		return false, "", nil
	}
	return true, matches[1], matches[2:]
}

// MatchCommandWithSingleParam 匹配带单个参数的命令
// pattern: 命令模式，如 "(猜拳|rock)"
// message: 输入消息
// 返回值: 是否匹配，命令名称，参数
func (p *CommandParser) MatchCommandWithSingleParam(pattern string, message string) (bool, string, string) {
	// 构建正则表达式：支持可选/前缀和任意空格
	regPattern := regexp.MustCompile(`^(?:/\s*)?(` + pattern + `)\s+(.+)\s*$`)
	matches := regPattern.FindStringSubmatch(message)
	if len(matches) != 3 {
		return false, "", ""
	}
	return true, matches[1], strings.TrimSpace(matches[2])
}

// IsCommand 判断消息是否为命令（以/开头或匹配指定命令模式）
// pattern: 命令模式，如 "(points|积分|猜拳|rock)"
// message: 输入消息
// 返回值: 是否为命令
func (p *CommandParser) IsCommand(pattern string, message string) bool {
	// 构建正则表达式：支持可选/前缀和任意空格
	regPattern := regexp.MustCompile(`^(?:/\s*)?(` + pattern + `)`)
	return regPattern.MatchString(message)
}

// GetCommandPrefix 获取命令前缀（如果有）
// message: 输入消息
// 返回值: 命令前缀（/或空）
func (p *CommandParser) GetCommandPrefix(message string) string {
	// 检查是否以/开头
	if strings.HasPrefix(strings.TrimSpace(message), "/") {
		return "/"
	}
	return ""
}

// ExtractCommand 提取命令部分
// message: 输入消息
// 返回值: 命令部分（去除前缀和空格）
func (p *CommandParser) ExtractCommand(message string) string {
	// 去除可选的/前缀和空格
	cmd := strings.TrimSpace(message)
	if strings.HasPrefix(cmd, "/") {
		cmd = strings.TrimSpace(cmd[1:])
	}
	// 提取第一个单词作为命令
	parts := strings.Fields(cmd)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
