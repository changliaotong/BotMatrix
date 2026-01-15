using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.ChatCompletion;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IModelProvider
    {
        string ProviderName { get; }
                
        Task<string> ExecuteAsync(ChatHistory history, ModelExecutionOptions options);
        IAsyncEnumerable<string> StreamExecuteAsync(ChatHistory history, ModelExecutionOptions options);

        Task<string> GenerateImageAsync(string prompt, ModelExecutionOptions options);

        Task<float[]> GenerateEmbeddingAsync(string text, ModelExecutionOptions options);

        Kernel BuildKernel(ModelExecutionOptions options);
    }

    public class ModelExecutionOptions
    {
        public string? ModelId { get; set; }
        public string? ChatModelId { get; set; }
        public string? EmbeddingModelId { get; set; }
        public string? ImageModelId { get; set; }
        
        public string? BaseUrl { get; set; }
        public string? ChatBaseUrl { get; set; }
        public string? EmbeddingBaseUrl { get; set; }
        public string? ImageBaseUrl { get; set; }

        public string? ApiKey { get; set; }
        public string? ChatApiKey { get; set; }
        public string? EmbeddingApiKey { get; set; }
        public string? ImageApiKey { get; set; }

        public IEnumerable<KernelPlugin>? Plugins { get; set; }
        public IEnumerable<IFunctionInvocationFilter>? Filters { get; set; }
        public CancellationToken CancellationToken { get; set; } = default;
        
        // 允许传递额外参数
        public Dictionary<string, object> ExtraParameters { get; set; } = new();
    }
}
