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
	cfg *config.WeatherConfig
}

func (p *WeatherPlugin) Name() string {
	return "weather"
}

func (p *WeatherPlugin) Description() string {
	return "天气查询插件，支持城市天气查询"
}

func (p *WeatherPlugin) Version() string {
	return "1.0.0"
}

// NewWeatherPlugin 创建天气插件实例
func NewWeatherPlugin(cfg *config.WeatherConfig) *WeatherPlugin {
	return &WeatherPlugin{
		cfg: cfg,
	}
}

func (p *WeatherPlugin) Init(robot plugin.Robot) {
	log.Println("加载天气查询插件")

	// 处理天气查询命令
	robot.OnMessage(func(event *onebot.Event) error {
		if event.MessageType != "group" && event.MessageType != "private" {
			return nil
		}

		// 检查是否为天气查询命令
		msg := strings.TrimSpace(event.RawMessage)
		if !strings.HasPrefix(msg, "!weather ") && !strings.HasPrefix(msg, "!天气 ") {
			return nil
		}

		// 解析城市名称
		var city string
		if strings.HasPrefix(msg, "!weather ") {
			city = strings.TrimSpace(msg[9:])
		} else {
			city = strings.TrimSpace(msg[4:])
		}

		if city == "" {
			// 发送帮助信息
			helpMsg := "天气查询命令格式：\n!weather 城市名\n!天气 城市名\n例如：!weather 北京"
			p.sendMessage(robot, event, helpMsg)
			return nil
		}

		// 查询天气
		weatherInfo, err := p.getWeatherInfo(city)
		if err != nil {
			log.Printf("查询天气失败: %v\n", err)
			errorMsg := fmt.Sprintf("查询天气失败：%v", err)
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
	params := &onebot.SendMessageParams{
		GroupID: event.GroupID,
		UserID:  event.UserID,
		Message: message,
	}

	if _, err := robot.SendMessage(params); err != nil {
		log.Printf("发送消息失败: %v\n", err)
	}
}

// WeatherInfo 天气信息结构体
type WeatherInfo struct {
	Name string `json:"name"`
	Main struct {
		Temp     float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
		Pressure int     `json:"pressure"`
		Humidity int     `json:"humidity"`
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
		return nil, fmt.Errorf("天气API密钥未配置")
	}

	// 构建请求URL
	baseURL, err := url.Parse(p.cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("构建请求URL失败: %w", err)
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
		return nil, fmt.Errorf("请求天气API失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("天气API返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var weatherInfo WeatherInfo
	if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
		return nil, fmt.Errorf("解析天气API响应失败: %w", err)
	}

	return &weatherInfo, nil
}

// formatWeatherInfo 格式化天气信息
func (p *WeatherPlugin) formatWeatherInfo(info *WeatherInfo) string {
	// 检查天气数据是否完整
	if len(info.Weather) == 0 {
		return "无法获取完整的天气信息"
	}

	// 格式化输出
	weather := info.Weather[0]
	return fmt.Sprintf("当前天气信息\n城市: %s\n天气: %s (%s)\n温度: %.1f°C (体感温度: %.1f°C)\n最低温度: %.1f°C, 最高温度: %.1f°C\n湿度: %d%%\n气压: %d hPa\n风速: %.1f m/s\n风向: %d°\n云量: %d%%\n日出: %s\n日落: %s",
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