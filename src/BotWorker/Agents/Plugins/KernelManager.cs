using System.Net;
using Microsoft.SemanticKernel;
using sz84.Agents.Providers.Configs;

namespace sz84.Agents.Plugins
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
            if (url == Doubao.Url)
                return ModelProvider.Doubao;
            else if (url == QWen.Url)
                return ModelProvider.Qwen;
            return ModelProvider.Unknown;
        }

        public static readonly Dictionary<ModelProvider, HttpClient> _httpClients = new()
        {
            [ModelProvider.Qwen] = CreateClient(QWen.Url),
            [ModelProvider.Doubao] = CreateClient(Doubao.Url)
        };


        private static HttpClient CreateClient(string baseUrl)
        {
            var handler = new SocketsHttpHandler
            {
                MaxConnectionsPerServer = 100,
                PooledConnectionLifetime = TimeSpan.FromMinutes(10),
                AutomaticDecompression = DecompressionMethods.GZip | DecompressionMethods.Deflate,
                EnableMultipleHttp2Connections = true
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

            var provider = GetProviderFromUrl(baseUrl);
            if (!_httpClients.TryGetValue(provider, out var client))
                throw new InvalidOperationException($"未识别模型 URL: {baseUrl}");

            var builder = Kernel.CreateBuilder()
                .AddOpenAIChatCompletion(modelId, apiKey, httpClient: client); // ✅ 使用共享连接

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
