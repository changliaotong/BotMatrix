using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderManager
    {
        private readonly Dictionary<string, IModelProvider> _providers = [];
        private readonly ILogger<ModelProviderManager>? _logger;
        private static readonly Random _random = new();

        public ModelProviderManager(ILogger<ModelProviderManager>? logger = null)
        {
            _logger = logger;
        }

        public void RegisterProvider(IModelProvider provider)
        {
            _providers[provider.ProviderName] = provider;
        }

        public async Task LoadFromDatabaseAsync()
        {
            try
            {
                var models = await LLMModel.GetAllActiveAsync();
                var providers = await LLMProvider.GetAllAsync();

                if (providers == null || models == null) return;

                foreach (var model in models)
                {
                    var provider = providers.FirstOrDefault(p => p.Id == model.ProviderId);
                    if (provider == null) continue;

                    IModelProvider? apiProvider = null;

                    if (provider.ProviderType.Equals("azure", StringComparison.OrdinalIgnoreCase))
                    {
                        apiProvider = new OpenAIAzureApiHelper(model.Name, provider.BaseUrl, provider.ApiKey);
                    }
                    else if (provider.ProviderType.Equals("openai", StringComparison.OrdinalIgnoreCase))
                    {
                        // 使用通用 OpenAI 提供者
                        apiProvider = new GenericOpenAIProvider(model.Name, provider.ApiKey, provider.BaseUrl, model.Name);
                    }

                    if (apiProvider != null)
                    {
                        RegisterProvider(apiProvider);
                    }
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "Failed to load AI providers from database");
            }
        }

        public IModelProvider? GetProvider(string providerName)
        {
            if (providerName.Equals("Random", StringComparison.OrdinalIgnoreCase))
            {
                return GetRandomProvider();
            }

            if (!_providers.TryGetValue(providerName, out var provider))
            {
                return null;
            }

            return provider;
        }

        public IModelProvider? GetRandomProvider()
        {
            if (_providers.Count == 0) return null;
            var keys = _providers.Keys.ToList();
            var randomKey = keys[_random.Next(keys.Count)];
            return _providers[randomKey];
        }

        public IEnumerable<string> GetAvailableModels()
        {
            return _providers.Keys;
        }
    }
}
