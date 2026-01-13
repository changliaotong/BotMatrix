using System.Net.Http.Json;
using System.Text.Json;
using System.Collections.Generic;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using Microsoft.SemanticKernel.Embeddings;
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
                var modelId = options.ModelId ?? _defaultModelId;
                var httpClient = KernelManager.GetHttpClient(_url);

                // 豆包 (Ark) 的生图接口通常是 api/v3/images/generations
                // 如果 _url 包含 ark.cn-beijing.volces.com，则需要特殊处理路径
                var endpoint = _url.TrimEnd('/');
                if (endpoint.Contains("ark.cn-beijing.volces.com"))
                {
                    // 如果 baseUrl 只是到 v3，则补齐
                    if (!endpoint.EndsWith("/v3")) endpoint += "/v3";
                    endpoint += "/images/generations";
                }
                else
                {
                    // 通用 OpenAI 路径
                    endpoint += "/images/generations";
                }

                var isArk = endpoint.Contains("ark.cn-beijing.volces.com");
                
                var requestData = new Dictionary<string, object>
                {
                    { "model", modelId },
                    { "prompt", prompt },
                    { "response_format", "url" }
                };

                // 处理尺寸
                if (options.ExtraParameters.TryGetValue("size", out var size) && size != null)
                {
                    requestData["size"] = size.ToString() ?? "1024x1024";
                }
                else if (isArk)
                {
                    requestData["size"] = "2K"; // 豆包默认使用 2K
                }
                else
                {
                    requestData["size"] = "1024x1024";
                }

                // 处理豆包特有参数
                if (isArk)
                {
                    requestData["sequential_image_generation"] = options.ExtraParameters.TryGetValue("sequential_image_generation", out var sig) ? sig : "disabled";
                    requestData["watermark"] = options.ExtraParameters.TryGetValue("watermark", out var wm) ? wm : true;
                }

                // 合并其他额外参数
                foreach (var param in options.ExtraParameters)
                {
                    if (!requestData.ContainsKey(param.Key))
                    {
                        requestData[param.Key] = param.Value;
                    }
                }

                using var request = new HttpRequestMessage(HttpMethod.Post, endpoint);
                request.Headers.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", _apiKey);
                request.Content = JsonContent.Create(requestData);

                var response = await httpClient.SendAsync(request, options.CancellationToken);
                var content = await response.Content.ReadAsStringAsync();

                if (!response.IsSuccessStatusCode)
                {
                    _logger?.LogError("[{ProviderName}] GenerateImageAsync error: {StatusCode} - {Content}", ProviderName, response.StatusCode, content);
                    return $"❌ 错误：{response.StatusCode} - {content}";
                }

                var json = JsonDocument.Parse(content);
                var url = json.RootElement.GetProperty("data")[0].GetProperty("url").GetString();
                
                return url ?? "❌ 错误：解析返回内容失败。";
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
                var embeddingService = kernel.GetRequiredService<ITextEmbeddingGenerationService>();
                var result = await embeddingService.GenerateEmbeddingAsync(text, kernel, options.CancellationToken);
                return result.ToArray();
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[{ProviderName}] GenerateEmbeddingAsync error", ProviderName);
                return Array.Empty<float>();
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
