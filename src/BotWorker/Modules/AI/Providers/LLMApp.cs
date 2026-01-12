using BotWorker.Modules.AI.Services;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Providers
{
    public class LLMApp
    {
        public readonly ModelProviderManager _manager;
        public static IServiceProvider? ServiceProvider { get; set; }

        public LLMApp(ILogger<ModelProviderManager>? logger = null)
        {
            _manager = new ModelProviderManager(logger);
            // 异步初始化将由外部调用或在第一次使用时触发
        }

        public async Task InitializeAsync()
        {
            await _manager.LoadFromDatabaseAsync();
        }
    }
}
