using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderManager
    {
        private readonly Dictionary<string, IModelProvider> _providers = new(StringComparer.OrdinalIgnoreCase);
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
                Console.WriteLine("[AI] Starting to load AI providers from database...");
                _logger?.LogInformation("[AI] Starting to load AI providers from database...");

                // 确保表存在
                await LLMProvider.EnsureTableCreatedAsync();
                await LLMModel.EnsureTableCreatedAsync();

                var models = await LLMModel.GetAllActiveAsync();
                var providers = await LLMProvider.GetAllActiveAsync();

                if (providers == null || models == null)
                {
                    Console.WriteLine("[AI] No active providers or models found in database.");
                    _logger?.LogWarning("[AI] No active providers or models found in database.");
                    return;
                }

                Console.WriteLine($"[AI] Retrieved {models.Count} models and {providers.Count} providers from database.");
                _logger?.LogInformation("[AI] Retrieved {ModelCount} models and {ProviderCount} providers from database.", 
                    models.Count, providers.Count);

                foreach (var provider in providers)
                {
                    if (string.IsNullOrWhiteSpace(provider.ApiKey))
                    {
                        Console.WriteLine($"[AI] Provider {provider.Name} has empty API Key, skipping.");
                        _logger?.LogWarning("[AI] Provider {ProviderName} has empty API Key, skipping.", provider.Name);
                        continue;
                    }

                    IModelProvider? apiProvider = null;
                    var defaultModel = models?.FirstOrDefault(m => m.ProviderId == provider.Id)?.Name ?? "";

                    if (provider.ProviderType.Equals("azure", StringComparison.OrdinalIgnoreCase))
                    {
                        apiProvider = new OpenAIAzureApiHelper(provider.Name, provider.BaseUrl, provider.ApiKey);
                    }
                    else if (provider.ProviderType.Equals("openai", StringComparison.OrdinalIgnoreCase) || 
                             provider.ProviderType.Equals("doubao", StringComparison.OrdinalIgnoreCase) ||
                             provider.ProviderType.Equals("deepseek", StringComparison.OrdinalIgnoreCase))
                    {
                        // 许多提供者都兼容 OpenAI 协议
                        apiProvider = new GenericOpenAIProvider(provider.Name, provider.ApiKey, provider.BaseUrl, defaultModel);
                    }

                    if (apiProvider != null)
                    {
                        Console.WriteLine($"[AI] Registered AI Provider: {provider.Name} ({provider.ProviderType})");
                        _logger?.LogInformation("[AI] Registered AI Provider: {ProviderName} ({ProviderType})", provider.Name, provider.ProviderType);
                        RegisterProvider(apiProvider);
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[AI] Failed to load AI providers: {ex.Message}");
                _logger?.LogError(ex, "[AI] Failed to load AI providers from database");
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
