using System.Net.Http;
using System.Threading.Tasks;
using BotWorker.Common.Exts;

namespace BotWorker.Services
{
    public interface IWeatherService
    {
        Task<string> GetWeatherAsync(string city);
    }

    public class WeatherService : IWeatherService
    {
        private readonly HttpClient _httpClient;

        public WeatherService(IHttpClientFactory httpClientFactory)
        {
            _httpClient = httpClientFactory.CreateClient();
        }

        public async Task<string> GetWeatherAsync(string city)
        {
            if (string.IsNullOrEmpty(city)) return "请输入城市名称";
            
            // 示例：这里可以调用外部天气 API
            return $"城市: {city}\n天气: 晴\n温度: 25℃\n风力: 3级";
        }
    }
}
