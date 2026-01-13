using BotWorker.Modules.AI.Providers.Configs;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public class GenericOpenAIProvider : OpenAIBaseProvider
    {
        public GenericOpenAIProvider(OpenAIProviderConfig config, ILogger? logger = null)
            : base(config.ProviderName, config.Key, config.Url, config.ModelId, logger)
        {
        }

        public GenericOpenAIProvider(string providerName, string apiKey, string url, string defaultModelId, ILogger? logger = null)
            : base(providerName, apiKey, url, defaultModelId, logger)
        {
        }
    }
}
