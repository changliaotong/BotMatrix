using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderFactory : IModelProviderFactory
    {
        private readonly ILoggerFactory _loggerFactory;

        public ModelProviderFactory(ILoggerFactory loggerFactory)
        {
            _loggerFactory = loggerFactory;
        }

        public IModelProvider? CreateProvider(LLMProvider provider, string defaultModel)
        {
            var apiKey = provider.GetDecryptedApiKey();
            if (string.IsNullOrWhiteSpace(apiKey))
            {
                return null;
            }

            var type = provider.Type.ToLower();
            var logger = _loggerFactory.CreateLogger(provider.Name);

            return type switch
            {
                "azure" => new OpenAIAzureApiHelper(provider.Name, provider.Endpoint ?? "", apiKey),
                "openai" or "doubao" or "deepseek" or "qwen" or "ollama" => new GenericOpenAIProvider(provider.Name, apiKey, provider.Endpoint ?? "", defaultModel, logger),
                _ => null
            };
        }
    }
}
