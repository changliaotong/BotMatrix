using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using BotWorker.Agents.Interfaces;
using BotWorker.Agents.Plugins;
using BotWorker.Agents.Providers.Configs;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Logging;

namespace BotWorker.Agents.Providers.Helpers
{
    public class OpenAIApiHelper(OpenAIConfig config) : IModelProvider
    {
        // openai
        public static async Task<string> CallOpenAIAsync(ChatHistory history, string modelId, string apiKey, string url)
        {
            try
            {
                var provider = KernelManager.GetProviderFromUrl(url);
                if (!KernelManager._httpClients.TryGetValue(provider, out var client))
                {
                    ErrorMessage($"未识别的模型服务地址: {url}");
                    return RetryMsg;
                }

                var kernel = Kernel.CreateBuilder()
                    .AddOpenAIChatCompletion(modelId, apiKey, httpClient: client)
                    .Build();

                var chatService = kernel.GetRequiredService<IChatCompletionService>();
                var result = await chatService.GetChatMessageContentAsync(history);

                return result.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                ErrorMessage($"[CallOpenAIAsync] {ex.GetType().Name}: {ex.Message}");
                return RetryMsg;
            }
        }

        public static async Task<string> CallOpenAIAsync(ChatHistory history, string modelId, string apiKey, string url, BotMessage context, IEnumerable<object> plugins)
        {
            try
            {
                var kernel = KernelManager.GetKernel(modelId, apiKey, url, (IEnumerable<string>)plugins);
                var chat = kernel.GetRequiredService<IChatCompletionService>();

                var settings = new PromptExecutionSettings
                {
                    FunctionChoiceBehavior = FunctionChoiceBehavior.Auto()
                };

                var result = await chat.GetChatMessageContentAsync(history, settings);
                return result.Content ?? string.Empty;
            }
            catch (Exception ex)
            {
                ErrorMessage($"OpenAIApiHelper.CallOpenAIWithFunctionAsync\n{ex.Message}");
                return RetryMsg;
            }
        }

        // openai stream
        public static async Task CallStreamOpenAIAsync(ChatHistory history, Func<string, bool, CancellationToken, Task> onUpdate,
            string modelId, string apiKey, string url, CancellationToken cts)
        {
            try
            {
                var builder = Kernel.CreateBuilder().AddOpenAIChatCompletion(modelId, apiKey, httpClient: new HttpClient { BaseAddress = new Uri(url) });
                var chat = builder.Build().GetRequiredService<IChatCompletionService>();
                await foreach (var update in chat.GetStreamingChatMessageContentsAsync(history, cancellationToken: cts))
                {
                    cts.ThrowIfCancellationRequested();
                    //Console.WriteLine(cts.Token.IsCancellationRequested);
                    await onUpdate(update.AsString(), true, cts);
                }
            }
            catch (OperationCanceledException)
            {
                await onUpdate("\n[DONE] Cancelled", false, cts);
            }
            catch (Exception ex)
            {
                Debug(ex.Message, "OpenAIApiHelper.CallStreamOpenAIAsync");
                await onUpdate("=".Times(30) + $"\n{RetryMsg}", false, cts);
            }
            await onUpdate(string.Empty, false, cts);
        }

        public static async Task CallStreamOpenAIAsync(
            ChatHistory history,
            Func<string, bool, CancellationToken, Task> onUpdate,
            string modelId,
            string apiKey,
            string url,
            IEnumerable<KernelPlugin> plugins,
            BotMessage context,
            CancellationToken cts)
        {
            try
            {
                var httpClient = new HttpClient(new LoggingHandler(new HttpClientHandler()))
                {
                    BaseAddress = new Uri(url)
                };

                var builder = Kernel.CreateBuilder();
                builder.AddOpenAIChatCompletion(modelId, apiKey, httpClient: httpClient);

                //var builder = Kernel.CreateBuilder().AddOpenAIChatCompletion(modelId, apiKey, httpClient: new HttpClient { BaseAddress = new Uri(url) });

                foreach (var plugin in plugins)
                {
                    builder.Plugins.AddFromObject(plugin);
                }

                var kernel = builder.Build();
                var chat = kernel.GetRequiredService<IChatCompletionService>();

                var settings = new PromptExecutionSettings
                {
                    FunctionChoiceBehavior = FunctionChoiceBehavior.Auto()
                };

                while (true)
                {
                    // 获取当前回复（可能是纯文本，也可能包含函数调用）
                    var result = await chat.GetChatMessageContentAsync(history, settings, kernel);

                    if (result.Content is not null)
                    {
                        // 内容存在，开始流式输出
                        await foreach (var chunk in chat.GetStreamingChatMessageContentsAsync(history, settings, kernel, cancellationToken: cts))
                        {
                            cts.ThrowIfCancellationRequested();
                            await onUpdate(chunk.AsString(), true, cts);
                        }

                        break;
                    }

                    // 检查是否有函数要调用
                    var functionCalls = FunctionCallContent.GetFunctionCalls(result);
                    if (!functionCalls.Any()) break;

                    foreach (var functionCall in functionCalls)
                    {
                        try
                        {
                            if (functionCall.FunctionName == "get_knowledge")
                            {
                                // 解析模型给的参数 JSON
                                if (functionCall == null || functionCall.Arguments == null)
                                    continue;

                                var question = functionCall.Arguments["question"]?.ToString() ?? string.Empty;

                                // 调用你的函数
                                if (context.KbService == null)
                                    continue;

                                var plugin = new KnowledgeBasePlugin(context.KbService, context.GroupId);                                
                                var kbResult = await plugin.GetKnowledgeAsync(question);

                                InfoMessage($"KB Result: {kbResult}");

                                // 把结果加进聊天记录
                                history.AddAssistantMessage(kbResult);
                            }
                            else
                            {
                                // 其他函数正常调用
                                var resultContent = await functionCall.InvokeAsync(kernel);
                                history.Add(resultContent.ToChatMessage());
                            }
                        }
                        catch (Exception ex)
                        {
                            history.Add(new FunctionResultContent(functionCall, ex).ToChatMessage());
                        }
                    }

                    // 添加模型回应到历史记录
                    history.Add(result);
                }

                await onUpdate(string.Empty, false, cts);
            }
            catch (OperationCanceledException)
            {
                await onUpdate("\n[DONE] Cancelled", false, cts);
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "OpenAIApiHelper.CallStreamOpenAIWithFunctionAsync");
                await onUpdate("=".Times(30) + $"\n{RetryMsg}", false, cts);
                await onUpdate(string.Empty, false, cts);
            }
        }

        private readonly OpenAIConfig _config = config;
        public string ProviderName => "Azure OpenAI";

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId)
        {
            return await CallOpenAIAsync(history, modelId, _config.Key, _config.Url);
        }

        public async Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins)
        {
            return await CallOpenAIAsync(history, modelId, _config.Key, _config.Url, context, plugins);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts)
        {
            await CallStreamOpenAIAsync(history, onUpdate, modelId, _config.Key, _config.Url, cts);
        }

        public async Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts)
        {
            await CallStreamOpenAIAsync(history, onUpdate, modelId, _config.Key, _config.Url, plugins, context, cts);
        }

    }
}
