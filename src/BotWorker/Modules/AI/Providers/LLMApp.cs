using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Services;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Providers
{
    public class LLMApp
    {
        public readonly ModelProviderManager _manager;
        public static IServiceProvider? ServiceProvider { get; set; }

        public LLMApp(ModelProviderManager manager)
        {
            _manager = manager;
        }

        public async Task InitializeAsync()
        {
            await _manager.LoadFromDatabaseAsync();
        }
    }
}
