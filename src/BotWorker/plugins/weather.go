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
	"time"
)

type WeatherPlugin struct {
	cfg       *config.WeatherConfig
	cmdParser *CommandParser
}

func (p *WeatherPlugin) Name() string {
	return "weather"
}

func (p *WeatherPlugin) Description() string {
	return common.T("", "weather_plugin_desc|🌤️ 天气查询插件，支持全球城市天气实时查询")
}

func (p *WeatherPlugin) Version() string {
	return "1.0.0"
}

// GetSkills 报备插件技能
func (p *WeatherPlugin) GetSkills() []plugin.SkillCapability {
	return []plugin.SkillCapability{
		{
			Name:        "get_weather",
			Description: common.T("", "weather_skill_get_weather_desc|查询指定城市的天气信息"),
			Usage:       "get_weather city=北京",
			Params: map[string]string{
				"city": common.T("", "weather_skill_param_city|城市名称"),
			},
		},
	}
}

// HandleSkill 处理技能调用
func (p *WeatherPlugin) HandleSkill(robot plugin.Robot, event *onebot.Event, skillName string, params map[string]string) (string, error) {
	switch skillName {
	case "get_weather":
		city := params["city"]
		if city == "" {
			return "", fmt.Errorf(common.T("", "weather_missing_city|❌ 请提供城市名称"))
		}
		weatherInfo, err := p.getWeatherInfo(city)
		if err != nil {
			return "", err
		}
		return p.formatWeatherInfo(weatherInfo), nil
	default:
		return "", fmt.Errorf("unknown skill: %s", skillName)
	}
}

// NewWeatherPlugin 创建天气插件实例
func NewWeatherPlugin(cfg *config.WeatherConfig) *WeatherPlugin {
	return &WeatherPlugin{
		cfg:       cfg,
		cmdParser: NewCommandParser(),
	}
}

func (p *WeatherPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "weather_plugin_loaded|✅ 天气插件已加载"))

	// 注册技能处理器
	skills := p.GetSkills()
	for _, skill := range skills {
		skillName := skill.Name
		robot.HandleSkill(skillName, func(params map[string]string) (string, error) {
			return p.HandleSkill(robot, nil, skillName, params)
		})
	}

	// 处理天气查询命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		if event.MessageType == "group" {
			groupIDStr := fmt.Sprintf("%d", event.GroupID)
			if !IsFeatureEnabledForGroup(GlobalDB, groupIDStr, "weather") {
				HandleFeatureDisabled(robot, event, "weather")
				return nil
			}
		}

		// 使用命令解析器检查并解析天气查询命令
		var city string
		// 首先检查是否为带参数的天气查询命令
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "weather_cmd_query|天气"), "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 提取城市名称
			city = strings.TrimSpace(params[0])
		} else {
			// 检查是否为帮助请求（不带参数）
			matchHelp, _ := p.cmdParser.MatchCommand(common.T("", "weather_cmd_query|天气"), event.RawMessage)
			if !matchHelp {
				return nil
			}

			// 发送帮助信息
			helpMsg := common.T("", "weather_help_msg|💡 天气查询使用方法：\n输入 “天气 [城市名]” 即可查询，例如：“天气 北京”")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		if city == "" {
			// 发送帮助信息
			helpMsg := common.T("", "weather_help_msg|💡 天气查询使用方法：\n输入 “天气 [城市名]” 即可查询，例如：“天气 北京”")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 查询天气
		weatherInfo, err := p.getWeatherInfo(city)
		if err != nil {
			log.Printf(common.T("", "weather_query_failed_log|❌ 天气查询失败：%v"), err)
			errorMsg := fmt.Sprintf(common.T("", "weather_query_failed_msg|❌ 天气查询失败：%v"), err)
			p.sendMessage(robot, event, errorMsg)
			return err
		}

		// 格式化天气信息
		weatherMsg := p.formatWeatherInfo(weatherInfo)

		// 发送天气信息
		p.sendMessage(robot, event, weatherMsg)

		return nil
	})
}

// sendMessage 发送消息
func (p *WeatherPlugin) sendMessage(robot plugin.Robot, event *onebot.Event, message string) {
	if robot == nil || event == nil {
		log.Printf(common.T("", "weather_send_failed_log|❌ 发送天气消息失败：%s"), message)
		return
	}

	_, err := SendTextReply(robot, event, message)
	if err != nil {
		log.Printf(common.T("", "weather_send_failed_log|❌ 发送天气消息失败：%v"), err)
	}
}

// WeatherInfo 天气信息结构体
type WeatherInfo struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Sys struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
}

// getWeatherInfo 获取天气信息
func (p *WeatherPlugin) getWeatherInfo(city string) (*WeatherInfo, error) {
	// 如果启用了模拟数据，或者API密钥为空且城市名为"模拟"或"mock"
	if p.cfg.Mock || (p.cfg.APIKey == "" && (city == "模拟" || strings.ToLower(city) == "mock")) {
		return p.getMockWeatherInfo(city), nil
	}

	// 检查API密钥是否配置
	if p.cfg.APIKey == "" {
		return nil, fmt.Errorf(common.T("", "weather_api_key_not_set|❌ 未配置天气API Key"))
	}

	// 构建请求URL
	baseURL, err := url.Parse(p.cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf(common.T("", "weather_build_url_failed|❌ 构建天气请求URL失败：%v"), err)
	}

	// 添加查询参数
	params := url.Values{}
	params.Add("q", city)
	params.Add("appid", p.cfg.APIKey)
	params.Add("units", "metric") // 使用摄氏度
	params.Add("lang", "zh_cn")   // 使用中文
	baseURL.RawQuery = params.Encode()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: p.cfg.Timeout,
	}

	// 发送请求
	resp, err := client.Get(baseURL.String())
	if err != nil {
		return nil, fmt.Errorf(common.T("", "weather_api_request_failed|❌ 请求天气API失败：%v"), err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(common.T("", "weather_api_error_status|❌ 天气API返回错误状态码：%d"), resp.StatusCode)
	}

	// 解析响应
	var weatherInfo WeatherInfo
	if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
		return nil, fmt.Errorf(common.T("", "weather_parse_response_failed|❌ 解析天气响应失败：%v"), err)
	}

	return &weatherInfo, nil
}

// formatWeatherInfo 格式化天气信息
func (p *WeatherPlugin) formatWeatherInfo(info *WeatherInfo) string {
	// 检查天气数据是否完整
	if len(info.Weather) == 0 {
		return common.T("", "weather_incomplete_info|❌ 天气信息不完整")
	}

	// 格式化输出
	weather := info.Weather[0]
	return fmt.Sprintf(common.T("", "weather_info_format|🌤️ 城市：%s\n☁️ 天气：%s (%s)\n🌡️ 温度：%.1f°C (体感 %.1f°C)\n❄️ 最低：%.1f°C / 🔥 最高：%.1f°C\n💧 湿度：%d%%\n🌬️ 风速：%.1f m/s (风向 %d°)\n☁️ 云量：%d%%\n🌅 日出：%s / 🌇 日落：%s"),
		info.Name,
		weather.Main,
		weather.Description,
		info.Main.Temp,
		info.Main.FeelsLike,
		info.Main.TempMin,
		info.Main.TempMax,
		info.Main.Humidity,
		info.Wind.Speed,
		info.Wind.Deg,
		info.Clouds.All,
		time.Unix(info.Sys.Sunrise, 0).Format("15:04"),
		time.Unix(info.Sys.Sunset, 0).Format("15:04"),
	)
}

// getMockWeatherInfo 返回模拟的天气信息
func (p *WeatherPlugin) getMockWeatherInfo(city string) *WeatherInfo {
	if city == "模拟" || strings.ToLower(city) == "mock" {
		city = "北京"
	}

	return &WeatherInfo{
		Name: city + " (模拟数据)",
		Main: struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			TempMin   float64 `json:"temp_min"`
			TempMax   float64 `json:"temp_max"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
		}{
			Temp:      25.5,
			FeelsLike: 26.8,
			TempMin:   20.0,
			TempMax:   30.0,
			Pressure:  1013,
			Humidity:  65,
		},
		Weather: []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		}{
			{
				Main:        "Clear",
				Description: "晴朗",
				Icon:        "01d",
			},
		},
		Wind: struct {
			Speed float64 `json:"speed"`
			Deg   int     `json:"deg"`
		}{
			Speed: 3.5,
			Deg:   180,
		},
		Clouds: struct {
			All int `json:"all"`
		}{
			All: 10,
		},
		Sys: struct {
			Sunrise int64 `json:"sunrise"`
			Sunset  int64 `json:"sunset"`
		}{
			Sunrise: time.Now().Add(-6 * time.Hour).Unix(),
			Sunset:  time.Now().Add(6 * time.Hour).Unix(),
		},
	}
}
