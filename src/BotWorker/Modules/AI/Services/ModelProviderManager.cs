using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderManager
    {
        private readonly Dictionary<string, IModelProvider> _providers = new(StringComparer.OrdinalIgnoreCase);
        private readonly Dictionary<string, (IModelProvider Provider, string ModelId, LLMModelType Type)> _modelMapping = new(StringComparer.OrdinalIgnoreCase);
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
                await UserAIConfig.EnsureTableCreatedAsync();

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

                _modelMapping.Clear();

                foreach (var provider in providers)
                {
                    Console.WriteLine($"[AI] Processing Provider: {provider.Name}, Status: {provider.Status}, Type: {provider.ProviderType}, HasKey: {!string.IsNullOrWhiteSpace(provider.ApiKey)}");
                    if (string.IsNullOrWhiteSpace(provider.ApiKey))
                    {
                        Console.WriteLine($"[AI] Provider {provider.Name} has empty API Key, skipping.");
                        _logger?.LogWarning("[AI] Provider {ProviderName} has empty API Key, skipping.", provider.Name);
                        continue;
                    }

                    IModelProvider? apiProvider = null;
                    var providerModels = models.Where(m => m.ProviderId == provider.Id).ToList();
                    var defaultModel = providerModels.FirstOrDefault()?.Name ?? "";

                    if (provider.ProviderType.Equals("azure", StringComparison.OrdinalIgnoreCase))
                    {
                        apiProvider = new OpenAIAzureApiHelper(provider.Name, provider.BaseUrl, provider.ApiKey);
                    }
                    else if (provider.ProviderType.Equals("openai", StringComparison.OrdinalIgnoreCase) || 
                             provider.ProviderType.Equals("doubao", StringComparison.OrdinalIgnoreCase) ||
                             provider.ProviderType.Equals("deepseek", StringComparison.OrdinalIgnoreCase))
                    {
                        apiProvider = new GenericOpenAIProvider(provider.Name, provider.ApiKey, provider.BaseUrl, defaultModel);
                    }

                    if (apiProvider != null)
                    {
                        Console.WriteLine($"[AI] Registered AI Provider: {provider.Name} ({provider.ProviderType})");
                        _logger?.LogInformation("[AI] Registered AI Provider: {ProviderName} ({ProviderType})", provider.Name, provider.ProviderType);
                        RegisterProvider(apiProvider);

                        // 注册具体模型映射
                        foreach (var model in providerModels)
                        {
                            _modelMapping[model.Name] = (apiProvider, model.Name, (LLMModelType)model.ModelType);
                            Console.WriteLine($"[AI] Registered Model: {model.Name} (Type: {(LLMModelType)model.ModelType}) for Provider: {provider.Name}");
                        }
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
                // 如果找不到 Provider，尝试按模型名称找
                if (_modelMapping.TryGetValue(providerName, out var mapping))
                {
                    return mapping.Provider;
                }
                return null;
            }

            return provider;
        }

        public (IModelProvider? Provider, string? ModelId) GetProviderAndModel(string? modelOrProviderName, LLMModelType preferredType = LLMModelType.Chat)
        {
            if (string.IsNullOrEmpty(modelOrProviderName) || modelOrProviderName.Equals("Random", StringComparison.OrdinalIgnoreCase))
            {
                // 随机选择一个符合类型的模型
                var suitableModels = _modelMapping.Values.Where(m => m.Type == preferredType).ToList();
                if (suitableModels.Count > 0)
                {
                    var picked = suitableModels[_random.Next(suitableModels.Count)];
                    return (picked.Provider, picked.ModelId);
                }
                return (GetRandomProvider(), null);
            }

            // 1. 尝试作为模型名称查找
            if (_modelMapping.TryGetValue(modelOrProviderName, out var mapping))
            {
                return (mapping.Provider, mapping.ModelId);
            }

            // 2. 尝试作为 Provider 名称查找
            if (_providers.TryGetValue(modelOrProviderName, out var provider))
            {
                // 找到该 Provider 下符合类型的第一个模型
                var providerModel = _modelMapping.Values.FirstOrDefault(m => m.Provider == provider && m.Type == preferredType);
                return (provider, providerModel.ModelId);
            }

            return (null, null);
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
