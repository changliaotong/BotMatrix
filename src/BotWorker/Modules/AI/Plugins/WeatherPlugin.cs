using Microsoft.SemanticKernel;
using System.ComponentModel;
using BotWorker.Infrastructure.Tools;

namespace BotWorker.Modules.AI.Plugins
{
    internal class WeatherPlugin
    {
        private readonly IToolService _toolService;

        public WeatherPlugin(IToolService toolService)
        {
            _toolService = toolService;
        }

        [KernelFunction("Weather")]
        [Description("天气预报，有城市名返回该城市天气，否则返回默认城市天气预报")]
        public async Task<string> GetWeatherAsync(string cityName = "深圳")
        {
            var res = await _toolService.GetWeatherAsync(new[] { cityName });
            return res.TryGetValue(cityName, out var weather) ? weather : "获取天气失败";
        }
    }
}
