using sz84.Agents.Providers.Configs;
using sz84.Agents.Providers.Helpers;
using sz84.Agents.Services;

namespace sz84.Agents.Providers
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
            _manager.RegisterProvider(new OpenAIAzureApiHelper(new OpenAIAzureConfig(AzureOpenAI.DeploymentName, AzureOpenAI.Endpoint, AzureOpenAI.ApiKey)));

            _manager.RegisterProvider(new OllamaApiHelper(new OllamaConfig(Ollama.ModelId, Ollama.OllamaUrl)));

            _manager.RegisterProvider(new DoubaoApiHelper(new DoubaoConfig(Doubao.Url, Doubao.Key, Doubao.ModelId)));

            _manager.RegisterProvider(new QWenApiHelper(new QWenConfig(QWen.Url, QWen.Key, QWen.ModelId)));

            _manager.RegisterProvider(new DeepSeekApiHelper(new DeepSeekConfig(DeepSeek.Url, DeepSeek.Key, DeepSeek.ModelId)));
        }
    }
}
