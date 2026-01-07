using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using sz84.Agents.Interfaces;
using sz84.Agents.Providers.Configs;
using sz84.Bots.BotMessages;
using BotWorker.Common.Exts;

namespace sz84.Agents.Providers.Helpers
{
    public class OpenAIAzureApiHelper(OpenAIAzureConfig config) : IModelProvider
    {
        //azure openai
        public static async Task<string> CallAzureOpenAIAsync(ChatHistory history, string apiKey, string endpoint, string deploymentName)
        {
            try
            {
                var result = await Kernel.CreateBuilder()
                    .AddAzureOpenAIChatCompletion(deploymentName, endpoint, apiKey).Build()
                    .GetRequiredService<IChatCompletionService>()
                    .GetChatMessageContentAsync(history);
                return result.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                Debug(ex.Message, "CallAzureOpenAIAsync");
                return RetryMsg;
            }
        }

        // Azure OpenAI stream
        public static async Task CallAzureOpenAIStreamAsync(ChatHistory history, Func<string, bool, CancellationToken, Task> onUpdate,
            string apiKey, string endpoint, string deploymentName, CancellationToken cts)
        {
            try
            {
                var chat = Kernel.CreateBuilder()
                    .AddAzureOpenAIChatCompletion(deploymentName, endpoint, apiKey).Build()
                    .GetRequiredService<IChatCompletionService>();
                await foreach (var update in chat.GetStreamingChatMessageContentsAsync(history, cancellationToken: cts))
                {
                    await onUpdate(update.AsString(), true, cts);
                }
                await onUpdate(string.Empty, false, cts);
            }
            catch (Exception ex)
            {
                Debug(ex.Message, "AzureOpenAIApiHelper.CallAzureOpenAIStreamAsync");
                await onUpdate("=".Times(30) + $"\n{RetryMsg}", false, cts);
            }
        }

        private readonly OpenAIAzureConfig _config = config;

        public string ProviderName => "Azure OpenAI";

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            return await CallAzureOpenAIAsync(history, _config.ApiKey, _config.Endpoint, _config.DeploymentName);
        }

        public Task<string> ExecuteAsync(ChatHistory history, string modelId, IEnumerable<object> plugins)
        {
            throw new NotImplementedException();
        }

        public Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<object> plugins)
        {
            throw new NotImplementedException();
        }

        public Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            throw new NotImplementedException();
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            await CallAzureOpenAIStreamAsync(history, onUpdate, _config.ApiKey, _config.Endpoint, _config.DeploymentName, cts);
        }

        public Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<object> plugins, CancellationToken cts)
        {
            throw new NotImplementedException();
        }

        public Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<object> plugins, BotMessage context, CancellationToken cts)
        {
            throw new NotImplementedException();
        }

        public Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            throw new NotImplementedException();
        }
    }
}
