package plugins

import (
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"botworker/internal/config"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
)

// TranslatePlugin 翻译插件
type TranslatePlugin struct {
	cfg       *config.TranslateConfig
	cmdParser *CommandParser
}

func (p *TranslatePlugin) Name() string {
	return "translate"
}

func (p *TranslatePlugin) Description() string {
	return "翻译插件，支持中英文互译"
}

func (p *TranslatePlugin) Version() string {
	return "1.0.0"
}

// NewTranslatePlugin 创建翻译插件实例
func NewTranslatePlugin(cfg *config.TranslateConfig) *TranslatePlugin {
	return &TranslatePlugin{
		cfg:       cfg,
		cmdParser: NewCommandParser(),
	}
}

func (p *TranslatePlugin) Init(robot plugin.Robot) {
	log.Println("加载翻译插件")

	// 处理翻译命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为翻译命令
		var content string
		// 首先检查是否为带参数的翻译命令
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams("翻译|translate", "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 解析翻译内容
			content = strings.TrimSpace(params[0])
		} else {
			// 检查是否为不带参数的翻译命令（显示帮助信息）
			matchHelp, _ := p.cmdParser.MatchCommand("翻译|translate", event.RawMessage)
			if !matchHelp {
				return nil
			}
			// 发送帮助信息
			helpMsg := "翻译命令格式：\n/翻译 <文本> - 翻译指定文本\n/translate <文本> - 翻译指定文本\n例如：/translate Hello world"
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 进行翻译
		translation, err := p.translate(content)
		if err != nil {
			log.Printf("翻译失败: %v\n", err)
			errorMsg := fmt.Sprintf("翻译失败：%v", err)
			p.sendMessage(robot, event, errorMsg)
			return err
		}

		// 发送翻译结果
		translateMsg := fmt.Sprintf("翻译结果：\n原文：%s\n译文：%s", content, translation)
		p.sendMessage(robot, event, translateMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *TranslatePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// TranslateResponse 翻译API响应结构体
type TranslateResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Result string `json:"result"`
	} `json:"data"`
}

// translate 翻译文本
func (p *TranslatePlugin) translate(text string) (string, error) {
	// 检查API密钥是否配置
	if p.cfg.APIKey == "" {
		return "", fmt.Errorf("翻译API密钥未配置")
	}

	// 检查文本语言
	isChinese := p.IsChinese(text)

	// 构建请求URL
	baseURL := p.cfg.Endpoint
	params := url.Values{}
	params.Add("api-version", "3.0")
	if isChinese {
		params.Add("to", "en")
	} else {
		params.Add("to", "zh-Hans")
	}
	params.Add("from", "auto")

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: p.cfg.Timeout,
	}

	// 构建请求体
	requestBody := []map[string]string{
		{"Text": text},
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", baseURL+"?"+params.Encode(), strings.NewReader(string(requestBodyBytes)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", p.cfg.APIKey)
	req.Header.Set("Ocp-Apim-Subscription-Region", p.cfg.Region)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求翻译API失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("翻译API返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result []struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析翻译API响应失败: %w", err)
	}

	// 提取翻译结果
	if len(result) > 0 && len(result[0].Translations) > 0 {
		return result[0].Translations[0].Text, nil
	}

	return "", fmt.Errorf("无法获取翻译结果")
}

// IsChinese 检查文本是否为中文
func (p *TranslatePlugin) IsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

