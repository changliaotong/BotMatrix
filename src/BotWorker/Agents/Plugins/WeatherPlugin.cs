using Microsoft.SemanticKernel;
using System.ComponentModel;
using BotWorker.Infrastructure.Tools;

namespace BotWorker.Agents.Plugins
{
    internal class WeatherPlugin
    {
        [KernelFunction("Weather")]
        [Description("天气预报，有城市名返回该城市天气，否则返回默认城市天气预报")]
        public static async Task<string> GetWeatherAsync(string cityName = "深圳")
        {
            return await Weather.GetWeatherAsync(cityName);
        }
    }
}
