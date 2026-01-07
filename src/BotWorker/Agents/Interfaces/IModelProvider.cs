using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;
using sz84.Bots.BotMessages;

namespace sz84.Agents.Interfaces
{
    public interface IModelProvider
    {
        string ProviderName { get; }
                
        Task<string> ExecuteAsync(ChatHistory history, string modelId);
        Task<string> ExecuteAsync(ChatHistory history, string modelId, BotMessage context, IEnumerable<KernelPlugin> plugins);
        Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, CancellationToken cts);
        Task StreamExecuteAsync(ChatHistory history, string modelId, Func<string, bool, CancellationToken, Task> onUpdate, IEnumerable<KernelPlugin> plugins, BotMessage context, CancellationToken cts);
    }
}
