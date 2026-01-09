using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IModelProvider
    {
        string ProviderName { get; }
                
        Task<string> ExecuteAsync(ChatHistory history, ModelExecutionOptions options);
        IAsyncEnumerable<string> StreamExecuteAsync(ChatHistory history, ModelExecutionOptions options);
    }

    public class ModelExecutionOptions
    {
        public string? ModelId { get; set; }
        public IEnumerable<KernelPlugin>? Plugins { get; set; }
        public CancellationToken CancellationToken { get; set; } = default;
        
        // 允许传递额外参数
        public Dictionary<string, object> ExtraParameters { get; set; } = new();
    }
}
