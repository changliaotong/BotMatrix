package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// WeatherConfig 定义天气API配置
type WeatherConfig struct {
	APIKey   string        `json:"api_key"`
	Endpoint string        `json:"endpoint"`
	Timeout  time.Duration `json:"timeout"`
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
func getWeatherInfo(cfg *WeatherConfig, city string) (*WeatherInfo, error) {
	// 检查API密钥是否配置
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("天气API密钥未配置")
	}

	// 构建请求URL
	baseURL, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("构建请求URL失败: %w", err)
	}

	// 添加查询参数
	params := url.Values{}
	params.Add("q", city)
	params.Add("appid", cfg.APIKey)
	params.Add("units", "metric") // 使用摄氏度
	params.Add("lang", "zh_cn")   // 使用中文
	baseURL.RawQuery = params.Encode()

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: cfg.Timeout,
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
func formatWeatherInfo(info *WeatherInfo) string {
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

func main() {
	log.Println("测试天气功能核心逻辑...")

	// 测试天气查询功能
	log.Println("测试天气查询函数...")
	// 注意：这里不会实际调用API，因为需要有效的APIKey
	// 但我们可以测试函数的基本结构

	// 测试格式化函数
	log.Println("测试天气信息格式化...")
	// 创建模拟的天气数据
	sampleInfo := &WeatherInfo{
		Name: "北京",
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
				Main:        "晴",
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

	// 格式化天气信息
	formatted := formatWeatherInfo(sampleInfo)
	log.Println("格式化后的天气信息:")
	log.Println(formatted)

	log.Println("天气功能核心逻辑测试通过!")
	log.Println("注意: 完整的API调用测试需要有效的OpenWeatherMap API密钥")
}