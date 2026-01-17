using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Core.Configurations;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Extensions;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Infrastructure.Tools;
using BotWorker.Infrastructure.Utils;
using Microsoft.CodeAnalysis.CSharp.Scripting;
using Microsoft.CodeAnalysis.Scripting;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class ToolService : IToolService
    {
        private static readonly HttpClient _httpClient = new();
        private readonly IWeatherRepository _weatherRepository;
        private readonly IIDCRepository _idcRepository;

        public ToolService(IWeatherRepository weatherRepository, IIDCRepository idcRepository)
        {
            _weatherRepository = weatherRepository;
            _idcRepository = idcRepository;
        }

        public async Task<Dictionary<string, string>> GetWeatherAsync(IEnumerable<string> cities)
        {
            Dictionary<string, string> res = new();
            foreach (var cityName in cities)
            {
                string? weatherInfo = await _weatherRepository.GetRecentWeatherAsync(cityName, 5);
                if (string.IsNullOrEmpty(weatherInfo))
                {
                    weatherInfo = await GetWeatherFromApiAsync(cityName);
                    if (!string.IsNullOrEmpty(weatherInfo) && 
                        weatherInfo != "没有此位置的天气资料" && 
                        weatherInfo != "天气预报功能暂时不能使用")
                    {
                        await _weatherRepository.InsertAsync(new Weather 
                        { 
                            CityName = cityName, 
                            WeatherInfo = weatherInfo,
                            InsertDate = DateTime.Now
                        });
                    }
                }
                res.TryAdd(cityName, weatherInfo ?? "获取天气失败");
            }
            return res;
        }

        private static readonly Dictionary<string, string> WeatherEmoji = new()
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

        private static int GetDisplayWidth(string text)
        {
            int width = 0;
            foreach (char c in text)
            {
                if (c >= 0x4E00 && c <= 0x9FA5) width += 3;
                else width += 1;
            }
            return width;
        }

        private static string PadRightWide(string text, int totalWidth)
        {
            int w = GetDisplayWidth(text);
            int spaces = totalWidth - w;
            return text + new string(' ', Math.Max(0, spaces));
        }

        private async Task<string> GetWeatherFromApiAsync(string cityName)
        {
            var encodedCityName = WebUtility.UrlEncode(cityName);
            var url = $"https://restapi.amap.com/v3/weather/weatherInfo?key=5fd93c8870028ba274e66ab20d8c4a7d&city={encodedCityName}&extensions=all&output=json";
            
            try
            {
                var response = await url.GetUrlDataAsync();
                if (string.IsNullOrEmpty(response)) return "天气预报功能暂时不能使用";

                var weatherData = JObject.Parse(response);
                var forecasts = JArray.Parse(weatherData["forecasts"]!.ToString());

                if (forecasts.Count == 0) return "没有此位置的天气资料";

                var reportTime = string.Empty;
                var res = string.Empty;

                foreach (var forecast in forecasts)
                {
                    var forecastData = JObject.Parse(forecast.ToString());
                    var city = forecastData["city"]!.ToString();
                    var province = forecastData["province"]!.ToString();
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

                        var day = i == 0 ? "今天 " : $"周{"一二三四五六日"[int.Parse(week) - 1]} ";
                        var temperature = $"{nightTemp}℃~{dayTemp}℃";
                        var weather = dayWeather == nightWeather 
                            ? i == 0 ? WeatherEmoji.GetValueOrDefault(dayWeather, dayWeather) : dayWeather 
                            : i == 0 ? WeatherEmoji.GetValueOrDefault(dayWeather) + WeatherEmoji.GetValueOrDefault(nightWeather) : $"{dayWeather}转{nightWeather}";

                        string? wind = null;
                        if (i == 0)
                        {
                            dayPower = dayPower == nightPower ? dayPower : $"{dayPower}转{nightPower}";
                            dayWind = dayWind == nightWind ? dayWind : $"{dayWind}转{nightWind}";
                            wind = $"{dayWind}风{dayPower}级";
                        }

                        lines.Add(($"{day}{weather}", temperature, wind));
                    }

                    int maxLeft = lines.Max(l => GetDisplayWidth(l.Left));
                    for (int i = 0; i < lines.Count; i++)
                    {
                        var l = lines[i];
                        var leftPadded = PadRightWide(l.Left, maxLeft + 2);
                        if (i == 0)
                        {
                            weatherInfo += $"{l.Left} {l.Temp} {l.Wind}\n";
                            weatherInfo += "----------------------\n";
                        }
                        else
                        {
                            weatherInfo += $"{leftPadded}{l.Temp}\n";
                        }
                    }
                    res += $"\n✅ {city}·{province}\n----------------------\n{weatherInfo.Trim()}\n";
                }

                return $"{res}----------------------\n发布时间：{reportTime}".Trim('\n');
            }
            catch
            {
                return "天气预报功能暂时不能使用";
            }
        }

        public async Task<string> GetCountDownAsync()
        {
            var now = DateTime.Today;
            var dates = new[]
            {
                new DateTime(2025, 10, 1),
                new DateTime(2025, 10, 6),
                new DateTime(2026, 1, 1),
                new DateTime(2026, 2, 17),
                new DateTime(2026, 6, 7)
            };

            var diffs = dates.Select(d => (d - now).Days).Cast<object>().ToArray();
            
            string template = "🕒 2025倒计时：\n🇨🇳 国庆节{0}天✨(25/10/01)\n🌕 中秋节{1}天🥮（25/10/06）\n\n🕒 2026倒计时：\n✨ 元旦{2}天🎉（26/01/01）\n🏮 春节{3}天🧨（26/02/17）\n📚 高考{4}天✏️（26/06/07）";
            
            return await Task.FromResult(string.Format(template, diffs));
        }

        public string GetMonthRes(DateTime dt, bool isYinli = false, int spaceCount = 3, int spaceCount2 = 1)
        {
            // Replicating logic from Calendar.cs
            DateTime FirstDay = dt.AddDays(-dt.Day + 1);
            DateTime LastDay = FirstDay.AddMonths(1).AddDays(-1);
            int dayOfWeek = (int)FirstDay.DayOfWeek;
            dayOfWeek = dayOfWeek == 0 ? 7 : dayOfWeek;

            string res = $"\n\n{" ".Times((int)(isYinli ? 8 + Ext.Max(spaceCount * 2, spaceCount2 * 3) : 4 + spaceCount2 * 3))}{dt.Year}年{dt.Month}月\n\n{(isYinli ? " " : "  ")}";

            foreach (var dow in Yinli.dayOfWeeks2)
                res += isYinli ? $" {dow}{" ".Times(spaceCount2 + 1)}" : $"{dow}{" ".Times(spaceCount - 2)}";

            string res1 = "\n" + " ".Times((dayOfWeek - 1) * (isYinli ? spaceCount + 2 : spaceCount) + 2);
            string res2 = " ".Times((dayOfWeek - 1) * (spaceCount2 + 4));
            int j = 0;
            for (int i = 0; i < LastDay.Day; i++)
            {
                DateTime today = FirstDay.AddDays(i);
                res1 += $"{(today.Day < 10 ? $"0{today.Day}" : $"{today.Day}")}{" ".Times(isYinli ? spaceCount : spaceCount - 2)}";
                if (isYinli)
                {
                    if (isYinli && (dt > Yinli.dateMax || dt < Yinli.dateMin))
                        return $"农历仅支持{Yinli.dateMin}至{Yinli.dateMax}";
                    try
                    {
                        Yinli yldt = new(today);
                        res2 += (yldt.Day == 1 ? $"{yldt.MonthName}{(yldt.MonthName?.Length > 1 ? "" : "月")}" : yldt.DayName) + " ".Times(spaceCount2);
                    }
                    catch (Exception ex)
                    {
                        SQLConn.DbDebug(ex.Message, "Calendar 日历");
                        return $"农历仅支持{Yinli.dateMin}至{Yinli.dateMax}";
                    }
                }
                if (today.DayOfWeek == DayOfWeek.Sunday || today.Month == LastDay.Month && today.Day == LastDay.Day)
                {
                    res += $"  {res1}\n";
                    if (isYinli)
                        res += $" {res2}\n";
                    res1 = "";
                    res2 = "";
                    j++;
                }
            }
            return res + "\n".Times(6 - j);
        }

        public async Task<string> GetTranslateAsync(string text)
        {
            string subscriptionKey = AppConfig.AzureTranslateSubscriptionKey;
            string endpoint = AppConfig.AzureTranslateEndpoint;
            string location = AppConfig.AzureTranslateLocation;

            if (string.IsNullOrEmpty(subscriptionKey)) return "翻译服务未配置";

            try
            {
                string detectRequestBody = JsonConvert.SerializeObject(new[] { new { Text = text } });
                string detectRequestUrl = $"{endpoint}/detect?api-version=3.0";

                using var request = new HttpRequestMessage(HttpMethod.Post, detectRequestUrl);
                request.Content = new StringContent(detectRequestBody, Encoding.UTF8, "application/json");
                request.Headers.Add("Ocp-Apim-Subscription-Key", subscriptionKey);
                request.Headers.Add("Ocp-Apim-Subscription-Region", location);

                var response = await _httpClient.SendAsync(request);
                if (!response.IsSuccessStatusCode) return "语言检测失败";

                var body = await response.Content.ReadAsStringAsync();
                var detection = JsonConvert.DeserializeObject<DetectionResponse[]>(body);
                string detectedLanguage = detection![0].Language ?? "en";

                string targetLanguage = detectedLanguage == "zh-Hans" ? "en" : "zh-Hans";

                string translateRequestUrl = $"{endpoint}/translate?api-version=3.0&to={targetLanguage}";
                using var translateRequest = new HttpRequestMessage(HttpMethod.Post, translateRequestUrl);
                translateRequest.Content = new StringContent(detectRequestBody, Encoding.UTF8, "application/json");
                translateRequest.Headers.Add("Ocp-Apim-Subscription-Key", subscriptionKey);
                translateRequest.Headers.Add("Ocp-Apim-Subscription-Region", location);

                var translateResponse = await _httpClient.SendAsync(translateRequest);
                if (!translateResponse.IsSuccessStatusCode) return "翻译失败";

                var translateBody = await translateResponse.Content.ReadAsStringAsync();
                var translation = JsonConvert.DeserializeObject<TranslationResponse[]>(translateBody);
                return translation![0].Translations![0].Text ?? "翻译结果为空";
            }
            catch (Exception ex)
            {
                Logger.Error($"Translation error: {ex.Message}");
                return "翻译服务异常";
            }
        }

        public async Task<string> CalculateAsync(string expression)
        {
            try
            {
                // Clean expression
                expression = expression.Replace("＋", "+").Replace("－", "-").Replace("×", "*").Replace("／", "/").Replace("[", "(").Replace("]", ")").Replace("（", "(").Replace("）", ")").Replace("÷", "/");
                expression = expression.Replace(";", "").Replace("ｘ", "*").Replace("＊", "*");
                expression = expression.Replace("=", "").Replace("＝", "").Replace("?", "").Replace("？", "");
                if (expression.Contains('/')) expression = expression.Replace("/", "*1.0/");

                var result = await CSharpScript.EvaluateAsync<double>(expression);
                return result.ToString();
            }
            catch
            {
                return "不正确的表达式";
            }
        }

        public async Task<string> GetCidResAsync(BotWorker.Domain.Models.BotMessages.BotMessage msg, bool isMinus = true)
        {
            var id = msg.Message;
            if (id.Length != 18)
                return $"命令格式：身份证 + 18位号码\n例如：\n身份证 {await GenerateRandomIDAsync(id)}";
            
            string ymd = id[6..14];
            string result;

            if (ymd == "********")
            {
                if (!CheckIDCard18(id.Replace("********", "20111111"), false))
                    return "身份证号不正确";
                result = GuessId(id);
            }
            else
            {
                if (!CheckIDCard(id))
                    return "身份证号不正确";

                result = $"身份证号：{id}\n" +
                         $"地区：{await GetAreaNameAsync(id[..6])}\n" +
                         $"生日：{id[6..10]}年{id[10..12]}月{id[12..14]}日\n" +
                         $"性别：{(int.Parse(id[14..17]) % 2 == 0 ? "女" : "男")} 年龄：{DateTime.Now.Year - int.Parse(id[6..10])}";
            }

            if (isMinus)            
                result += msg.MinusCreditRes(10, "查身份证扣分");

            return result;
        }

        public async Task<string> GetAreaNameAsync(string areaCode)
        {
            return await _idcRepository.GetAreaNameAsync(areaCode) ?? "未知";
        }

        private async Task<string> GenerateRandomIDAsync(string dq = "")
        {
            string areaCode = await _idcRepository.GetRandomBmAsync(dq) ?? "110101";

            Random rnd = new Random();
            int year = rnd.Next(1920, DateTime.Now.Year);
            int month = rnd.Next(1, 13);
            int day = rnd.Next(1, DateTime.DaysInMonth(year, month) + 1);
            int order = rnd.Next(1, 1000);

            string id = $"{areaCode}{year}{month:00}{day:00}{order:03}";

            int[] factors = { 7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2 };
            int sum = 0;
            for (int i = 0; i < 17; i++)
            {
                sum += factors[i] * int.Parse(id[i].ToString());
            }
            int mod = sum % 11;
            string[] checkCodes = { "1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2" };

            return $"{id}{checkCodes[mod]}";
        }

        private bool CheckIDCard(string id)
        {
            return id.Length switch
            {
                18 => CheckIDCard18(id),
                15 => CheckIDCard15(id),
                _ => false
            };
        }

        private bool CheckIDCard18(string id, bool isCheckValid = true)
        {
            if (long.TryParse(id.Remove(17), out long n) == false || n < Math.Pow(10, 16) || long.TryParse(id.Replace('x', '0').Replace('X', '0'), out n) == false)
                return false;

            if (!System.Globalization.DateTime.TryParseExact(id.Substring(6, 8), "yyyyMMdd", System.Globalization.CultureInfo.InvariantCulture, System.Globalization.DateTimeStyles.None, out _))
                return false;

            if (isCheckValid)
            {
                int[] factors = { 7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2 };
                int sum = factors.Select((factor, index) => factor * int.Parse(id[index].ToString())).Sum();
                int mod = sum % 11;
                string[] checkCode = { "1", "0", "x", "9", "8", "7", "6", "5", "4", "3", "2" };
                if (!string.Equals(checkCode[mod], id.Substring(17, 1), StringComparison.OrdinalIgnoreCase))
                    return false;
            }

            return true;
        }

        private bool CheckIDCard15(string id)
        {
            if (long.TryParse(id, out long n) == false || n < Math.Pow(10, 14))
                return false;
            return true;
        }

        private string GuessId(string id)
        {
            string res = string.Empty;
            for (int year = DateTime.Now.Year; year >= 1900; year--)
            {
                for (int month = 12; month >= 1; month--)
                {
                    int daysInMonth = DateTime.DaysInMonth(year, month);
                    for (int day = daysInMonth; day >= 1; day--)
                    {
                        string newid = id.Replace("********", $"{year}{month:00}{day:00}");
                        if (CheckIDCard18(newid))
                        {
                            res += $"{newid}\n";
                        }
                    }
                }
            }
            return res;
        }

        private class DetectionResponse
        {
            [JsonProperty("language")]
            public string? Language { get; set; }
            [JsonProperty("score")]
            public float Score { get; set; }
        }

        private class TranslationResponse
        {
            [JsonProperty("translations")]
            public Translation[]? Translations { get; set; }
        }

        private class Translation
        {
            [JsonProperty("text")]
            public string? Text { get; set; }
            [JsonProperty("to")]
            public string? To { get; set; }
        }
    }
}
