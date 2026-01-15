using System.Net.Http.Json;
using System.Text.Json;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using Microsoft.SemanticKernel.Embeddings;
using Microsoft.SemanticKernel.TextToImage;
using Microsoft.Extensions.AI;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Plugins;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.AI.Providers.Helpers
{
    public abstract class OpenAIBaseProvider : IModelProvider
    {
        protected readonly string _apiKey;
        protected readonly string _url;
        protected readonly string _defaultModelId;
        protected readonly string _providerName;
        protected readonly ILogger? _logger;

        public virtual string ProviderName => _providerName;

        protected OpenAIBaseProvider(string providerName, string apiKey, string url, string defaultModelId, ILogger? logger = null)
        {
            _providerName = providerName;
            _apiKey = SanitizeHeaderValue(apiKey?.Trim() ?? string.Empty);
            _url = url?.Trim() ?? string.Empty;
            _defaultModelId = SanitizeHeaderValue(defaultModelId?.Trim() ?? string.Empty);
            _logger = logger;

            if (apiKey != null && apiKey != _apiKey)
            {
                _logger?.LogWarning("[{ProviderName}] API Key contained non-ASCII characters and was sanitized.", _providerName);
            }
            if (defaultModelId != null && defaultModelId != _defaultModelId)
            {
                _logger?.LogWarning("[{ProviderName}] Default Model ID contained non-ASCII characters and was sanitized.", _providerName);
            }
        }

        private static string SanitizeHeaderValue(string value)
        {
            if (string.IsNullOrEmpty(value)) return value;
            // 仅保留 ASCII 字符 (32-126)
            return new string(value.Where(c => c >= 32 && c <= 126).ToArray());
        }

        public virtual async Task<string> ExecuteAsync(ChatHistory history, ModelExecutionOptions options)
        {
            try
            {
                var kernel = BuildKernel(options);
                var chatService = kernel.GetRequiredService<IChatCompletionService>();
                
                var settings = new PromptExecutionSettings();
                if (kernel.Plugins.Count > 0)
                {
                    settings.FunctionChoiceBehavior = FunctionChoiceBehavior.Auto();
                }

                var result = await chatService.GetChatMessageContentAsync(history, settings, kernel, options.CancellationToken);
                return result.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] ExecuteAsync error", ProviderName);
                return $"❌ 错误：{ex.Message}";
            }
        }

        public virtual async IAsyncEnumerable<string> StreamExecuteAsync(ChatHistory history, ModelExecutionOptions options)
        {
            var kernel = BuildKernel(options);
            var chatService = kernel.GetRequiredService<IChatCompletionService>();

            var settings = new PromptExecutionSettings();
            if (kernel.Plugins.Count > 0)
            {
                settings.FunctionChoiceBehavior = FunctionChoiceBehavior.Auto();
            }

            IAsyncEnumerable<StreamingChatMessageContent> stream;
            try
            {
                stream = chatService.GetStreamingChatMessageContentsAsync(history, settings, kernel, options.CancellationToken);
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] StreamExecuteAsync start error", ProviderName);
                throw; // Re-throw instead of yield return in catch
            }

            await foreach (var chunk in stream.WithCancellation(options.CancellationToken))
            {
                if (chunk.Content != null)
                {
                    yield return chunk.Content;
                }
            }
        }

        public virtual async Task<string> GenerateImageAsync(string prompt, ModelExecutionOptions options)
        {
            try
            {
                var kernel = BuildKernel(options);
                var imageService = kernel.GetRequiredService<ITextToImageService>();

                // 处理尺寸
                int width = 1024;
                int height = 1024;
                if (options.ExtraParameters.TryGetValue("size", out var size) && size != null)
                {
                    var sizeStr = size.ToString() ?? "1024x1024";
                    var parts = sizeStr.Split('x');
                    if (parts.Length == 2 && int.TryParse(parts[0], out int w) && int.TryParse(parts[1], out int h))
                    {
                        width = w;
                        height = h;
                    }
                }

                return await imageService.GenerateImageAsync(prompt, width, height, kernel: kernel, cancellationToken: options.CancellationToken);
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] GenerateImageAsync error", ProviderName);
                return $"❌ 错误：{ex.Message}";
            }
        }

        public virtual async Task<float[]> GenerateEmbeddingAsync(string text, ModelExecutionOptions options)
        {
            try
            {
                var kernel = BuildKernel(options);
                var embeddingGenerator = kernel.GetRequiredService<IEmbeddingGenerator<string, Embedding<float>>>();
                var result = await embeddingGenerator.GenerateAsync(new List<string> { text }, cancellationToken: options.CancellationToken);
                return result.Count > 0 ? result[0].Vector.ToArray() : Array.Empty<float>();
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] GenerateEmbeddingAsync error", ProviderName);
                return Array.Empty<float>();
            }
        }

        public virtual Kernel BuildKernel(ModelExecutionOptions options)
        {
            var modelId = (options.ModelId ?? _defaultModelId)?.Trim();
            var chatModelId = (options.ChatModelId ?? modelId)?.Trim();
            var embeddingModelId = (options.EmbeddingModelId ?? (options.ModelId == null ? null : modelId))?.Trim();
            var imageModelId = (options.ImageModelId ?? (options.ModelId == null ? null : modelId))?.Trim();
            
            var baseUrl = (options.BaseUrl ?? _url)?.Trim();
            var chatBaseUrl = (options.ChatBaseUrl ?? baseUrl)?.Trim();
            var embeddingBaseUrl = (options.EmbeddingBaseUrl ?? baseUrl)?.Trim();
            var imageBaseUrl = (options.ImageBaseUrl ?? baseUrl)?.Trim();

            var apiKey = (options.ApiKey ?? _apiKey)?.Trim();
            var chatApiKey = (options.ChatApiKey ?? apiKey)?.Trim();
            var embeddingApiKey = (options.EmbeddingApiKey ?? apiKey)?.Trim();
            var imageApiKey = (options.ImageApiKey ?? apiKey)?.Trim();

            var builder = Kernel.CreateBuilder();

            if (!string.IsNullOrEmpty(chatBaseUrl) && !string.IsNullOrEmpty(chatModelId))
            {
                _logger?.LogDebug("[{ProviderName}] Adding Chat Completion: Model={ModelId}, Url={Url}", ProviderName, chatModelId, chatBaseUrl);
                builder.AddOpenAIChatCompletion(chatModelId, chatApiKey ?? string.Empty, httpClient: KernelManager.GetHttpClient(chatBaseUrl));
            }

            if (!string.IsNullOrEmpty(embeddingBaseUrl) && !string.IsNullOrEmpty(embeddingModelId))
            {
                _logger?.LogDebug("[{ProviderName}] Adding Embedding Generator: Model={ModelId}, Url={Url}", ProviderName, embeddingModelId, embeddingBaseUrl);
                builder.AddOpenAIEmbeddingGenerator(embeddingModelId, embeddingApiKey ?? string.Empty, httpClient: KernelManager.GetHttpClient(embeddingBaseUrl));
            }

            if (!string.IsNullOrEmpty(imageBaseUrl) && !string.IsNullOrEmpty(imageModelId))
            {
                _logger?.LogDebug("[{ProviderName}] Adding Text-to-Image: Model={ModelId}, Url={Url}", ProviderName, imageModelId, imageBaseUrl);
                builder.AddOpenAITextToImage(imageModelId, imageApiKey ?? string.Empty, httpClient: KernelManager.GetHttpClient(imageBaseUrl));
            }

            if (options.Plugins != null)
            {
                foreach (var plugin in options.Plugins)
                {
                    builder.Plugins.Add(plugin);
                }
            }

            if (options.Filters != null)
            {
                foreach (var filter in options.Filters)
                {
                    builder.Services.AddSingleton(filter);
                }
            }

            return builder.Build();
        }
    }
}
