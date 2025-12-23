package plugins

import (
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
	return common.T("", "weather_plugin_desc")
}

func (p *WeatherPlugin) Version() string {
	return "1.0.0"
}

// NewWeatherPlugin 创建天气插件实例
func NewWeatherPlugin(cfg *config.WeatherConfig) *WeatherPlugin {
	return &WeatherPlugin{
		cfg:       cfg,
		cmdParser: NewCommandParser(),
	}
}

func (p *WeatherPlugin) Init(robot plugin.Robot) {
	log.Println(common.T("", "weather_plugin_loaded"))

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
		matchWithParams, _, params := p.cmdParser.MatchCommandWithParams(common.T("", "weather_cmd_query"), "(.+)", event.RawMessage)
		if matchWithParams && len(params) == 1 {
			// 提取城市名称
			city = strings.TrimSpace(params[0])
		} else {
			// 检查是否为帮助请求（不带参数）
			matchHelp, _ := p.cmdParser.MatchCommand(common.T("", "weather_cmd_query"), event.RawMessage)
			if !matchHelp {
				return nil
			}

			// 发送帮助信息
			helpMsg := common.T("", "weather_help_msg")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		if city == "" {
			// 发送帮助信息
			helpMsg := common.T("", "weather_help_msg")
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 查询天气
		weatherInfo, err := p.getWeatherInfo(city)
		if err != nil {
			log.Printf(common.T("", "weather_query_failed_log"), err)
			errorMsg := fmt.Sprintf(common.T("", "weather_query_failed_msg"), err)
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
	if _, err := SendTextReply(robot, event, message); err != nil {
		log.Printf(common.T("", "weather_send_failed_log"), err)
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
	// 检查API密钥是否配置
	if p.cfg.APIKey == "" {
		return nil, fmt.Errorf(common.T("", "weather_api_key_not_set"))
	}

	// 构建请求URL
	baseURL, err := url.Parse(p.cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf(common.T("", "weather_build_url_failed"), err)
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
		return nil, fmt.Errorf(common.T("", "weather_api_request_failed"), err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(common.T("", "weather_api_error_status"), resp.StatusCode)
	}

	// 解析响应
	var weatherInfo WeatherInfo
	if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
		return nil, fmt.Errorf(common.T("", "weather_parse_response_failed"), err)
	}

	return &weatherInfo, nil
}

// formatWeatherInfo 格式化天气信息
func (p *WeatherPlugin) formatWeatherInfo(info *WeatherInfo) string {
	// 检查天气数据是否完整
	if len(info.Weather) == 0 {
		return common.T("", "weather_incomplete_info")
	}

	// 格式化输出
	weather := info.Weather[0]
	return fmt.Sprintf(common.T("", "weather_info_format"),
		info.Name,
		weather.Main,
		weather.Description,
		info.Main.Temp,
		info.Main.FeelsLike,
		info.Main.TempMin,
		info.Main.TempMax,
		info.Main.Humidity,
		info.Main.Pressure,
		info.Wind.Speed,
		info.Wind.Deg,
		info.Clouds.All,
		time.Unix(info.Sys.Sunrise, 0).Format("15:04"),
		time.Unix(info.Sys.Sunset, 0).Format("15:04"),
	)
}
