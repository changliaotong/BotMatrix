package plugins

import (
	"BotMatrix/common"
	"botworker/internal/config"
	"botworker/internal/onebot"
	"botworker/internal/plugin"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	return common.T("", "translate_plugin_desc|翻译插件，支持中英文互译")
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
	log.Println(common.T("", "translate_loaded|加载翻译插件"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理翻译命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "translate") {
				HandleFeatureDisabled(robot, event, "translate")
				return nil
			}
		}

		// 检查是否为翻译命令
		var content string
		// 首先检查是否为带参数的翻译命令
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "translate_cmd_translate|翻译|translate"), "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 解析翻译内容
			content = strings.TrimSpace(params[0])
		} else {
			// 检查是否为不带参数的翻译命令（显示帮助信息）
			matchHelp, _ := p.cmdParser.MatchCommand(common.T("", "translate_cmd_translate|翻译|translate"), event.RawMessage)
			if !matchHelp {
				return nil
			}
			// 发送帮助信息
			helpMsg := common.T("", "translate_help_msg|翻译命令格式：\n/翻译 <文本> - 翻译指定文本\n/translate <文本> - 翻译指定文本\n例如：/translate Hello world")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 进行翻译
		msg, err := p.doTranslate(content)
		if err != nil {
			log.Printf(common.T("", "translate_api_error_log|翻译失败: %v"), err)
			errorMsg := fmt.Sprintf(common.T("", "translate_api_error_msg|翻译失败：%v"), err)
			p.sendMessage(robot, event, errorMsg)
			return err
		}

		// 发送翻译结果
		p.sendMessage(robot, event, msg)

		return nil
	})
}

// GetSkills 报备插件技能
func (p *TranslatePlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "translate",
			Description: common.T("", "translate_skill_desc|翻译指定文本（中英互译）"),
			Usage:       "translate text=hello",
			Params: map[string]string{
				"text": common.T("", "translate_skill_param_text|待翻译的文本"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *TranslatePlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "translate":
		text := params["text"]
		if text == "" {
			return "", fmt.Errorf(common.T("", "translate_missing_param_text|缺少待翻译的文本参数"))
		}
		return p.doTranslate(text)
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// doTranslate 执行翻译逻辑
func (p *TranslatePlugin) doTranslate(content string) (string, error) {
	translation, err := p.translate(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(common.T("", "translate_result|翻译结果：\n原文：%s\n译文：%s"), content, translation), nil
}

// sendMessage 发送消息
func (p *TranslatePlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "translate_send_failed|发送翻译消息失败: %v"), message)
		return
	}
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "translate_send_failed|发送翻译消息失败: %v"), err)
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
		return "", fmt.Errorf(common.T("", "translate_api_key_not_set|Translate API Key not set"))
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
		return "", fmt.Errorf(common.T("", "translate_build_body_failed|Failed to build translation request body: %v"), err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", baseURL+"?"+params.Encode(), strings.NewReader(string(requestBodyBytes)))
	if err != nil {
		return "", fmt.Errorf(common.T("", "translate_create_request_failed|Failed to create translation request: %v"), err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", p.cfg.APIKey)
	req.Header.Set("Ocp-Apim-Subscription-Region", p.cfg.Region)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(common.T("", "translate_api_request_failed|Translate API request failed: %v"), err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(common.T("", "translate_api_error_status|Translate API returned error status code: %d"), resp.StatusCode)
	}

	// 解析响应
	var result []struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf(common.T("", "translate_parse_response_failed|Failed to parse translate API response: %v"), err)
	}

	// 提取翻译结果
	if len(result) > 0 && len(result[0].Translations) > 0 {
		return result[0].Translations[0].Text, nil
	}

	return "", fmt.Errorf(common.T("", "translate_no_result|No translation result found"))
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
