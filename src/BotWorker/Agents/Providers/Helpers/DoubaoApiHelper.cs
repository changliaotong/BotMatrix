using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Agents.Interfaces;
using BotWorker.Agents.Providers.Configs;
using BotWorker.Domain.Models.Messages.BotMessages;

namespace BotWorker.Agents.Providers.Helpers
{
    public class DoubaoApiHelper(DoubaoConfig config) : IModelProvider
    {
        private readonly DoubaoConfig _config = config;

        public string ProviderName => "Doubao";


        public static async Task<string> DoubaoAsync(ChatHistory history, string modelId, string apiKey, string url)
        {
            return await OpenAIApiHelper.CallOpenAIAsync(history, modelId, apiKey, url);
        }

        //stream
        public static async Task StreamDoubaoAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, string apiKey, string url, CancellationToken cts)
        {
            await OpenAIApiHelper.CallStreamOpenAIAsync(history, onUpdate, modelId, apiKey, url, cts);
        }

        //funcall
        public static async Task<string> DoubaoAsync(ChatHistory history, string modelId, string apiKey, string url, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await OpenAIApiHelper.CallOpenAIAsync(history, modelId, apiKey, url, context, plugins);
        }

        //stream funcall
        public static async Task StreamDoubaoAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, string apiKey, string url, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await OpenAIApiHelper.CallStreamOpenAIAsync(history, onUpdate, modelId, apiKey, url, plugins, context, cts);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            return await DoubaoAsync(history, modelId, _config.Key, _config.Url);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            await StreamDoubaoAsync(history, modelId, onUpdate, _config.Key, _config.Url, cts);
        }
        public async Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await DoubaoAsync(history, modelId, _config.Key, _config.Url, context, plugins);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await StreamDoubaoAsync(history, modelId, onUpdate, _config.Key, _config.Url, plugins, context, cts);
        }
    }
}
