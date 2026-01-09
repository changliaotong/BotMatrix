using System.Collections.Generic;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Plugins;
using Microsoft.Extensions.Logging;

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
            _apiKey = apiKey;
            _url = url;
            _defaultModelId = defaultModelId;
            _logger = logger;
        }

        public virtual async Task<string> ExecuteAsync(ChatHistory history, ModelExecutionOptions options)
        {
            try
            {
                var kernel = BuildKernel(options);
                var chatService = kernel.GetRequiredService<IChatCompletionService>();
                
                var settings = new PromptExecutionSettings
                {
                    FunctionChoiceBehavior = FunctionChoiceBehavior.Auto()
                };

                var result = await chatService.GetChatMessageContentAsync(history, settings, kernel, options.CancellationToken);
                return result.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] ExecuteAsync error", ProviderName);
                return $"Error: {ex.Message}";
            }
        }

        public virtual async IAsyncEnumerable<string> StreamExecuteAsync(ChatHistory history, ModelExecutionOptions options)
        {
            var kernel = BuildKernel(options);
            var chatService = kernel.GetRequiredService<IChatCompletionService>();

            var settings = new PromptExecutionSettings
            {
                FunctionChoiceBehavior = FunctionChoiceBehavior.Auto()
            };

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

        protected virtual Kernel BuildKernel(ModelExecutionOptions options)
        {
            var modelId = options.ModelId ?? _defaultModelId;
            
            // 使用 KernelManager 获取共享的 HttpClient
            var httpClient = KernelManager.GetHttpClient(_url);

            var builder = Kernel.CreateBuilder()
                .AddOpenAIChatCompletion(modelId, _apiKey, httpClient: httpClient);

            if (options.Plugins != null)
            {
                foreach (var plugin in options.Plugins)
                {
                    builder.Plugins.Add(plugin);
                }
            }

            return builder.Build();
        }
    }
}
