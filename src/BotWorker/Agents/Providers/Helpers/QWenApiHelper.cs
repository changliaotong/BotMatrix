using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using sz84.Agents.Interfaces;
using sz84.Agents.Providers.Configs;
using sz84.Bots.BotMessages;

namespace sz84.Agents.Providers.Helpers
{
    public class QWenApiHelper(QWenConfig config) : IModelProvider
    {
        public string ProviderName => "QWen";
        private readonly QWenConfig _config = config;

        public static async Task<string> QWenAsync(ChatHistory history, string modelId, string apiKey, string url)
        {
            return await OpenAIApiHelper.CallOpenAIAsync(history, QWen.ModelId, apiKey, url);
        }

        //stream
        public static async Task StreamQWenAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, string apiKey, string url, CancellationToken cts)
        {
            await OpenAIApiHelper.CallStreamOpenAIAsync(history, onUpdate, QWen.ModelId, apiKey, url, cts);
        }

        //funcall
        public static async Task<string> QWenAsync(ChatHistory history, string modelId, string apiKey, string url, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await OpenAIApiHelper.CallOpenAIAsync(history, QWen.ModelIdFunctionCall, apiKey, url, context, plugins);
        }

        //stream funcall
        public static async Task StreamQWenAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, string apiKey, string url, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await OpenAIApiHelper.CallStreamOpenAIAsync(history, onUpdate, QWen.ModelIdFunctionCall, apiKey, url, plugins, context, cts);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            return await QWenAsync(history, modelId, _config.Key, _config.Url);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            await StreamQWenAsync(history, modelId, onUpdate, _config.Key, _config.Url, cts);
        }
        public async Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await QWenAsync(history, modelId, _config.Key, _config.Url, context, plugins);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await StreamQWenAsync(history, modelId, onUpdate, _config.Key, _config.Url, plugins, context, cts);
        }
    }
}
