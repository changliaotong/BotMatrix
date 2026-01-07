using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using sz84.Agents.Interfaces;
using sz84.Agents.Providers.Configs;
using sz84.Bots.BotMessages;
using BotWorker.Common.Exts;

namespace sz84.Agents.Providers.Helpers
{
    public class DeepSeekApiHelper(DeepSeekConfig config) : IModelProvider
    {
        private readonly DeepSeekConfig _config = config;
        public string ProviderName => "DeepSeek";

        public static async Task<string> DeepSeekAsync(ChatHistory history, string modelId, string apiKey, string url)
        {
            if (modelId.IsNull())
                modelId = DeepSeek.ModelId;
            return await OpenAIApiHelper.CallOpenAIAsync(history, DeepSeek.ModelId, apiKey, url);
        }

        //stream
        public static async Task StreamDeepSeekAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, string apiKey, string url, CancellationToken cts)
        {
            if (modelId.IsNull())
                modelId = DeepSeek.ModelId;
            await OpenAIApiHelper.CallStreamOpenAIAsync(history, onUpdate, DeepSeek.ModelId, apiKey, url, cts);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            return await DeepSeekAsync(history, modelId, _config.Key, _config.Url);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await DeepSeekAsync(history, modelId, _config.Key, _config.Url);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            await StreamDeepSeekAsync(history, modelId, onUpdate, _config.Key, _config.Url, cts);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await StreamDeepSeekAsync(history, modelId, onUpdate, _config.Key, _config.Url, cts);
        }
    }
}
