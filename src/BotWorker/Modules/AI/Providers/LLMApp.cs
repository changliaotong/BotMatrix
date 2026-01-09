using BotWorker.Modules.AI.Providers.Configs;
using BotWorker.Modules.AI.Providers.Helpers;
using BotWorker.Modules.AI.Services;

namespace BotWorker.Modules.AI.Providers
{
    public class LLMApp
    {
        public readonly ModelProviderManager _manager;

        public LLMApp()
        {
            _manager = new ModelProviderManager();
            InitializeProviders();
        }

        private void InitializeProviders()
        {
            _manager.RegisterProvider(new OpenAIAzureApiHelper(AzureOpenAI.DeploymentName, AzureOpenAI.Endpoint, AzureOpenAI.ApiKey));

            _manager.RegisterProvider(new GenericOpenAIProvider("Ollama", "ollama", Ollama.OllamaUrl.EndsWith("/v1") ? Ollama.OllamaUrl : Ollama.OllamaUrl.TrimEnd('/') + "/v1", Ollama.ModelId));

            _manager.RegisterProvider(new GenericOpenAIProvider("Doubao", Doubao.Key, Doubao.Url, Doubao.ModelId));

            _manager.RegisterProvider(new GenericOpenAIProvider("QWen", QWen.Key, QWen.Url, QWen.ModelId));

            _manager.RegisterProvider(new GenericOpenAIProvider("DeepSeek", DeepSeek.Key, DeepSeek.Url, DeepSeek.ModelId));
        }
    }
}
