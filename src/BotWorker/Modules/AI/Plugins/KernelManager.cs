using System.Net;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.TextToImage;
using BotWorker.Modules.AI.Providers.Configs;
using BotWorker.Modules.AI.Providers.Helpers;

namespace BotWorker.Modules.AI.Plugins
{
    public static class KernelManager
    {
        private static readonly Dictionary<string, Kernel> _kernels = [];
        private static readonly Dictionary<string, object> _availablePlugins = [];

        public enum ModelProvider
        {
            OpenAI,
            Doubao,
            Qwen,
            Ollama,
            Unknown
        }

        public static ModelProvider GetProviderFromUrl(string url)
        {
            if (string.IsNullOrEmpty(url)) return ModelProvider.Unknown;
            
            if (url.Contains("volces.com", StringComparison.OrdinalIgnoreCase))
                return ModelProvider.Doubao;
            else if (url.Contains("dashscope.aliyuncs.com", StringComparison.OrdinalIgnoreCase))
                return ModelProvider.Qwen;
            else if (url.Contains("openai.com", StringComparison.OrdinalIgnoreCase))
                return ModelProvider.OpenAI;
            else if (url.Contains("localhost") || url.Contains("127.0.0.1"))
                return ModelProvider.Ollama;
                
            return ModelProvider.Unknown;
        }

        private static readonly Dictionary<string, HttpClient> _urlHttpClients = new();

        public static HttpClient GetHttpClient(string url)
        {
            if (_urlHttpClients.TryGetValue(url, out var client))
                return client;

            lock (_urlHttpClients)
            {
                if (_urlHttpClients.TryGetValue(url, out client))
                    return client;

                client = CreateClient(url);
                _urlHttpClients[url] = client;
                return client;
            }
        }


        private static HttpClient CreateClient(string baseUrl)
        {
            var handler = new SocketsHttpHandler
            {
                MaxConnectionsPerServer = 100,
                PooledConnectionLifetime = TimeSpan.FromMinutes(10),
                AutomaticDecompression = DecompressionMethods.GZip | DecompressionMethods.Deflate,
                EnableMultipleHttp2Connections = true,
                UseCookies = false // ✅ 禁用 Cookie，防止非 ASCII 字符导致的 Header 错误
            };

            return new HttpClient(handler)
            {
                BaseAddress = new Uri(baseUrl),
                Timeout = Timeout.InfiniteTimeSpan,
                DefaultRequestVersion = HttpVersion.Version20,
                DefaultVersionPolicy = HttpVersionPolicy.RequestVersionOrHigher
            };
        }

        /// <summary>
        /// 注册可用插件（所有插件）
        /// </summary>
        public static void RegisterPlugin(string name, object plugin)
        {
            _availablePlugins[name] = plugin;
        }

        /// <summary>
        /// 获取指定模型和插件组合的 Kernel（自动缓存）
        /// </summary>
        public static Kernel GetKernel(string modelId, string apiKey, string baseUrl, IEnumerable<string> pluginNames)
        {
            var pluginKey = string.Join(",", pluginNames.OrderBy(x => x));
            var cacheKey = $"{modelId}@{baseUrl}#{pluginKey}";

            if (_kernels.TryGetValue(cacheKey, out var cached))
                return cached;

            var client = GetHttpClient(baseUrl);

            var builder = Kernel.CreateBuilder()
                .AddOpenAIChatCompletion(modelId, apiKey, httpClient: client) // ✅ 使用共享连接
                .AddOpenAIEmbeddingGenerator(modelId, apiKey, httpClient: client); // ✅ 同时也支持向量生成

            // 使用 SK 标准的 AddOpenAITextToImage 调用生图（基于 OpenAI 兼容性）
            builder.AddOpenAITextToImage(modelId, apiKey, httpClient: client);

            foreach (var name in pluginNames)
            {
                if (_availablePlugins.TryGetValue(name, out var plugin))
                {
                    var kernelPlugin = plugin as KernelPlugin;
                    if (kernelPlugin != null)
                    {
                        builder.Plugins.Add(kernelPlugin);
                    }
                    else
                    {
                        throw new InvalidOperationException($"插件 {name} 未注册或类型不正确");
                    }
                }
                else
                {
                    throw new InvalidOperationException($"插件 {name} 未注册");
                }
            }

            var kernel = builder.Build();
            _kernels[cacheKey] = kernel;
            return kernel;
        }
    }


}
