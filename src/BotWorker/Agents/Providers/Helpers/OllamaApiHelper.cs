using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using Codeblaze.SemanticKernel.Connectors.Ollama;
using BotWorker.Agents.Interfaces;
using BotWorker.Agents.Providers.Configs;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Common.Extensions;

namespace BotWorker.Agents.Providers.Helpers
{
    public class OllamaApiHelper(OllamaConfig config) : IModelProvider
    {
        // ollama
        public static async Task<string> CallOllamaAsync(ChatHistory history, string modelId, string OllamaUrl)
        {
            try
            {
                var builder = Kernel.CreateBuilder();
                builder.Services.AddLogging(services => services.AddConsole().SetMinimumLevel(Microsoft.Extensions.Logging.LogLevel.Trace));
#pragma warning disable SKEXP0070
                builder.AddOllamaChatCompletion(modelId, new Uri(OllamaUrl));
#pragma warning restore SKEXP0070
                var response = await builder.Build()
                    .GetRequiredService<IChatCompletionService>()
                    .GetChatMessageContentAsync(history);
                return response.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "OpenAIApiHelper.CallOllamaAsync");
                return RetryMsg;
            }
        }

        // ollama stream
        public static async Task CallOllamaStreamAsync(ChatHistory history, Func<string, bool, CancellationToken, Task> onUpdate, string modelId, string OllamaUrl, CancellationToken cts)
        {
            try
            {
                var builder = Kernel.CreateBuilder();
                builder.Services.AddLogging(services => services.AddConsole().SetMinimumLevel(Microsoft.Extensions.Logging.LogLevel.Trace));
#pragma warning disable SKEXP0070
                builder.AddOllamaChatCompletion(modelId, new Uri(OllamaUrl));
#pragma warning restore SKEXP0070
                var chatService = builder.Build().GetRequiredService<IChatCompletionService>();
                var response = chatService.GetStreamingChatMessageContentsAsync(chatHistory: history);
                await foreach (var chunk in response)
                {
                    await onUpdate(chunk.ToString(), true, cts);
                    if (cts.IsCancellationRequested)
                    {
                        break;
                    }
                }
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "OpenAIApiHelper.CallOllamaStreamAsync");
                await onUpdate("=".Times(30) + $"\n{RetryMsg}", false, cts);
            }

            // 完成生成后，调用更新函数，去掉 '_'
            await onUpdate(string.Empty, false, cts);
        }

        private readonly OllamaConfig _config = config;
        private readonly List<ModelConfig> _models =
                [
                    new ModelConfig("wangshenzhi/gemma2-9b-chinese-chat:latest", "Ollama Gemma2 Chinese", "中文对话模型", "large", "chat", 4096),
                ];

        public string ProviderName => "Ollama";
        public IEnumerable<ModelConfig> AvailableModels => _models;

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            var model = _models.FirstOrDefault(m => m.ModelId == modelId) ?? throw new ArgumentException($"Model {modelId} not found.");
            return await CallOllamaAsync(history, _config.OllamaUrl, _config.ModelId);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            var model = _models.FirstOrDefault(m => m.ModelId == modelId) ?? throw new ArgumentException($"Model {modelId} not found.");
            return await CallOllamaAsync(history, _config.OllamaUrl, _config.ModelId);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            var model = _models.FirstOrDefault(m => m.ModelId == modelId) ?? throw new ArgumentException($"Model {modelId} not found.");
            await CallOllamaStreamAsync(history, onUpdate, _config.OllamaUrl, _config.ModelId, cts);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            var model = _models.FirstOrDefault(m => m.ModelId == modelId) ?? throw new ArgumentException($"Model {modelId} not found.");
            await CallOllamaStreamAsync(history, onUpdate, _config.OllamaUrl, _config.ModelId, cts);
        }
    }
}
