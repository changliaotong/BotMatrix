using Newtonsoft.Json.Linq;
using System.Net;
using sz84.Core.MetaDatas;
using sz84.Infrastructure.Utils;

namespace sz84.Infrastructure.Tools
{
    public class Weather : MetaData<Weather>
    {
        public override string TableName => "Weather";
        public override string KeyField => "Id";

        public static async Task<Dictionary<string, string>> GetWeatherAsync(IEnumerable<string> citys)
        {
            Dictionary<string, string> res = [];
            foreach (var cityName in citys)
            {
                string weatherInfo = GetWhere("WeatherInfo", $"CityName = {cityName.Quotes()} AND ABS(DATEDIFF(HOUR, GETDATE(), InsertDate)) < 5", "Id DESC");
                if (weatherInfo.IsNull() && !weatherInfo.In("没有此位置的天气资料", "天气预报功能暂时不能使用"))
                {
                    weatherInfo = await GetWeatherAsync(cityName);
                    Append(cityName, weatherInfo);
                }
                res.TryAdd(cityName, weatherInfo);
            }
            return res;
        }

        public static int Append(string cityName, string weather)
        {
            return Insert([
                            new Cov("CityName", cityName),
                            new Cov("WeatherInfo", weather),
                        ]);
        }

        public static int GetDisplayWidth(string text)
        {
            int width = 0;
            foreach (char c in text)
            {
                // 判断是否全角（中日韩字符）
                if (c >= 0x4E00 && c <= 0x9FA5)
                    width += 3;
                else
                    width += 1;
            }
            return width;
        }

        public static string PadRightWide(string text, int totalWidth)
        {
            int w = GetDisplayWidth(text);
            int spaces = totalWidth - w;

            return text + new string(' ', Math.Max(0, spaces));
        }

        public static readonly Dictionary<string, string> WeatherEmoji = new()
        {
            ["晴"] = "☀️",
            ["多云"] = "⛅",
            ["阴"] = "☁️",
            ["小雨"] = "🌦️",
            ["中雨"] = "🌧️",
            ["大雨"] = "🌧️💧",
            ["暴雨"] = "⛈️",
            ["雷阵雨"] = "⛈️⚡",
            ["小雪"] = "🌨️",
            ["中雪"] = "🌨️",
            ["大雪"] = "❄️❄️",
            ["暴雪"] = "🌨️❄️",
            ["雾"] = "🌫️",
            ["霾"] = "🌫️",
            ["沙尘"] = "🌪️",
            ["台风"] = "🌀"
        };

        public static async Task<string> GetWeatherAsync(string cityName)
        {
            var encodedCityName = WebUtility.UrlEncode(cityName);
            var url = $"https://restapi.amap.com/v3/weather/weatherInfo?key=5fd93c8870028ba274e66ab20d8c4a7d&city={encodedCityName}&extensions=all&output=json";
            var response = await url.GetUrlDataAsync(); 

            if (string.IsNullOrEmpty(response))
            {
                return "天气预报功能暂时不能使用";
            }

            try
            {
                var weatherData = JObject.Parse(response);
                var forecasts = JArray.Parse(weatherData["forecasts"]!.ToString());

                if (forecasts.Count == 0)
                {
                    return "没有此位置的天气资料";
                }

                var reportTime = string.Empty;
                var res = string.Empty;

                foreach (var forecast in forecasts)
                {
                    var forecastData = JObject.Parse(forecast.ToString());
                    cityName = forecastData["city"]!.ToString();
                    string? province = forecastData["province"]!.ToString();
                    reportTime = forecastData["reporttime"]!.ToString();

                    var weatherInfo = string.Empty;
                    var casts = JArray.Parse(forecastData["casts"]!.ToString());
                    var lines = new List<(string Left, string Temp, string? Wind)>();

                    for (int i = 0; i < casts.Count; i++)
                    {
                        var cast = JObject.Parse(casts[i].ToString());

                        var week = cast["week"]!.ToString();
                        var dayWeather = cast["dayweather"]!.ToString();
                        var nightWeather = cast["nightweather"]!.ToString();
                        var dayTemp = cast["daytemp"]!.ToString();
                        var nightTemp = cast["nighttemp"]!.ToString();
                        var dayWind = cast["daywind"]!.ToString();
                        var nightWind = cast["nightwind"]!.ToString();
                        var dayPower = cast["daypower"]!.ToString();
                        var nightPower = cast["nightpower"]!.ToString();

                        // 今日 or 周几
                        var day = i == 0 ? "今天 " : $"周{"一二三四五六日"[int.Parse(week) - 1]} ";

                        // 温度格式
                        var temperature = $"{nightTemp}℃~{dayTemp}℃";

                        // 天气合并
                        var weather = dayWeather == nightWeather 
                            ? i == 0 ? WeatherEmoji.GetValueOrDefault(dayWeather, dayWeather) : dayWeather 
                            : i == 0 ? WeatherEmoji.GetValueOrDefault(dayWeather) + WeatherEmoji.GetValueOrDefault(nightWeather) : $"{dayWeather}转{nightWeather}";

                        // 风力（只用于今日）
                        string? wind = null;
                        if (i == 0)
                        {
                            dayPower = dayPower == nightPower ? dayPower : $"{dayPower}转{nightPower}";
                            dayWind = dayWind == nightWind ? dayWind : $"{dayWind}转{nightWind}";
                            wind = $"{dayWind}风{dayPower}级";
                        }

                        // 左侧内容
                        var left = $"{day}{weather}";

                        lines.Add((left, temperature, wind));
                    }

                    // -------- 对齐处理 --------

                    // 找最长 left 字段
                    int maxLeft = lines.Max(l => GetDisplayWidth(l.Left));

                    for (int i = 0; i < lines.Count; i++)
                    {
                        var l = lines[i];

                        var leftPadded = PadRightWide(l.Left, maxLeft + 2);

                        if (i == 0) // 今日
                        {
                            weatherInfo += $"{l.Left} {l.Temp} {l.Wind}\n";
                            weatherInfo += "----------------------\n";
                        }
                        else // 其他天
                        {
                            weatherInfo += $"{leftPadded}{l.Temp}\n";
                        }
                    }

                    res += $"\n✅ {cityName}·{province}\n----------------------\n{weatherInfo.Trim()}\n";
                }

                return $"{res}----------------------\n发布时间：{reportTime}".Trim("\n").ToString();
            }
            catch (Exception)
            {
                return "天气预报功能暂时不能使用";
            }
        }
    }
}
