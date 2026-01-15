using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models;
using BotWorker.Modules.AI.Providers.Helpers;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class ModelProviderManager
    {
        public class ModelRegistration
        {
            public IModelProvider Provider { get; set; } = null!;
            public LLMModel Metadata { get; set; } = null!;
            public string ActualModelId => !string.IsNullOrWhiteSpace(Metadata.ApiModelId) ? Metadata.ApiModelId : Metadata.Name;
            public string? ActualBaseUrl => Metadata.BaseUrl;
            public string? ActualApiKey => Metadata.ApiKey;
            public LLMModelType Type => MapStringToType(Metadata.Type);

            private static LLMModelType MapStringToType(string type)
            {
                return type.ToLower() switch
                {
                    "chat" => LLMModelType.Chat,
                    "image" => LLMModelType.Image,
                    "embedding" => LLMModelType.Embedding,
                    "audio" => LLMModelType.Audio,
                    _ => LLMModelType.Chat
                };
            }
        }

        private readonly Dictionary<string, IModelProvider> _providers = new(StringComparer.OrdinalIgnoreCase);
        private readonly Dictionary<string, ModelRegistration> _modelMapping = new(StringComparer.OrdinalIgnoreCase);
        private readonly ILogger<ModelProviderManager>? _logger;
        private static readonly Random _random = new();

        private readonly ILLMRepository _llmRepository;
        private readonly IModelProviderFactory _providerFactory;

        public ModelProviderManager(ILLMRepository llmRepository, IModelProviderFactory providerFactory, ILogger<ModelProviderManager>? logger = null)
        {
            _llmRepository = llmRepository;
            _providerFactory = providerFactory;
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

                var models = (await _llmRepository.GetActiveModelsAsync()).ToList();
                var providers = (await _llmRepository.GetActiveProvidersAsync()).ToList();

                if (providers == null || models == null || !providers.Any())
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
                    Console.WriteLine($"[AI] Processing Provider: {provider.Name}, IsActive: {provider.IsActive}, Type: {provider.Type}");
                    
                    var providerModels = models.Where(m => m.ProviderId == provider.Id).ToList();
                    var defaultModel = providerModels.FirstOrDefault()?.Name ?? "";

                    var apiProvider = _providerFactory.CreateProvider(provider, defaultModel);

                    if (apiProvider != null)
                    {
                        Console.WriteLine($"[AI] Registered AI Provider: {provider.Name} ({provider.Type})");
                        _logger?.LogInformation("[AI] Registered AI Provider: {ProviderName} ({ProviderType})", provider.Name, provider.Type);
                        RegisterProvider(apiProvider);

                        // 注册具体模型映射
                        foreach (var model in providerModels)
                        {
                            _modelMapping[model.Name] = new ModelRegistration 
                            { 
                                Provider = apiProvider, 
                                Metadata = model 
                            };
                            Console.WriteLine($"[AI] Registered Model: {model.Name} (Internal ID: {_modelMapping[model.Name].ActualModelId}, Type: {_modelMapping[model.Name].Type}) for Provider: {provider.Name}");
                        }
                    }
                    else
                    {
                        Console.WriteLine($"[AI] Failed to create provider {provider.Name} (Type: {provider.Type}). Check API Key or Type.");
                        _logger?.LogWarning("[AI] Failed to create provider {ProviderName} (Type: {ProviderType}). Check API Key or Type.", provider.Name, provider.Type);
                    }
                }
            }
            catch (Exception ex)
            { 
                Console.WriteLine($"[AI] Failed to load AI providers: {ex.Message}");
                _logger?.LogError(ex, "[AI] Failed to load AI providers from database");
            }
        }

        private static LLMModelType MapStringToType(string type)
        {
            return type.ToLower() switch
            {
                "chat" => LLMModelType.Chat,
                "image" => LLMModelType.Image,
                "embedding" => LLMModelType.Embedding,
                "audio" => LLMModelType.Audio,
                _ => LLMModelType.Chat
            };
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

        public (IModelProvider? Provider, string? ModelId, string? BaseUrl, string? ApiKey) GetProviderAndModel(string? modelOrProviderName, LLMModelType preferredType = LLMModelType.Chat)
        {
            if (string.IsNullOrEmpty(modelOrProviderName) || modelOrProviderName.Equals("Random", StringComparison.OrdinalIgnoreCase))
            {
                // 随机选择一个符合类型的模型
                var suitableModels = _modelMapping.Values.Where(m => m.Type == preferredType).ToList();
                if (suitableModels.Count > 0)
                {
                    var picked = suitableModels[_random.Next(suitableModels.Count)];
                    return (picked.Provider, picked.ActualModelId, picked.ActualBaseUrl, picked.ActualApiKey);
                }
                return (null, null, null, null);
            }

            // 支持 "Random:OpenAI" 这种语法，表示在特定提供商中随机选择
            if (modelOrProviderName.StartsWith("Random:", StringComparison.OrdinalIgnoreCase))
            {
                var targetProvider = modelOrProviderName.Substring(7);
                var providerModels = _modelMapping.Values
                    .Where(m => m.Provider.ProviderName.Equals(targetProvider, StringComparison.OrdinalIgnoreCase) && m.Type == preferredType)
                    .ToList();
                
                if (providerModels.Count > 0)
                {
                    var picked = providerModels[_random.Next(providerModels.Count)];
                    return (picked.Provider, picked.ActualModelId, picked.ActualBaseUrl, picked.ActualApiKey);
                }
            }

            // 1. 尝试作为模型名称查找
            if (_modelMapping.TryGetValue(modelOrProviderName, out var mapping))
            {
                return (mapping.Provider, mapping.ActualModelId, mapping.ActualBaseUrl, mapping.ActualApiKey);
            }

            // 2. 尝试作为 Provider 名称查找
            if (_providers.TryGetValue(modelOrProviderName, out var provider))
            {
                // 找到该 Provider 下符合类型的第一个模型
                var providerModel = _modelMapping.Values.FirstOrDefault(m => m.Provider == provider && m.Type == preferredType);
                return (provider, providerModel?.ActualModelId, providerModel?.ActualBaseUrl, providerModel?.ActualApiKey);
            }

            return (null, null, null, null);
        }

        /// <summary>
        /// 根据更复杂的策略选择模型
        /// </summary>
        /// <param name="strategy">策略字符串，例如: "cheapest", "fastest", "random", "gpt-4o"</param>
        /// <param name="preferredType">模型类型</param>
        public (IModelProvider? Provider, string? ModelId, string? BaseUrl, string? ApiKey) SelectByStrategy(string strategy, LLMModelType preferredType = LLMModelType.Chat)
        {
            if (string.IsNullOrWhiteSpace(strategy)) return GetProviderAndModel(null, preferredType);

            return strategy.ToLower() switch
            {
                "random" => GetProviderAndModel("Random", preferredType),
                "cheapest" => GetCheapestModel(preferredType),
                _ => GetProviderAndModel(strategy, preferredType)
            };
        }

        private (IModelProvider? Provider, string? ModelId, string? BaseUrl, string? ApiKey) GetCheapestModel(LLMModelType preferredType)
        {
            var cheapest = _modelMapping.Values
                .Where(m => m.Type == preferredType)
                .OrderBy(m => m.Metadata.InputPricePer1kTokens + m.Metadata.OutputPricePer1kTokens)
                .FirstOrDefault();

            if (cheapest != null)
            {
                return (cheapest.Provider, cheapest.ActualModelId, cheapest.ActualBaseUrl, cheapest.ActualApiKey);
            }

            return GetProviderAndModel("Random", preferredType);
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
            return _modelMapping.Keys;
        }

        public IEnumerable<string> GetAvailableProviders()
        {
            return _providers.Keys;
        }
    }
}
