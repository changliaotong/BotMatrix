namespace BotWorker.Modules.AI.Providers.Configs
{
    public class OpenAIProviderConfig
    {
        public string ProviderName { get; set; } = string.Empty;
        public string Url { get; set; } = string.Empty;
        public string Key { get; set; } = string.Empty;
        public string ModelId { get; set; } = string.Empty;

        public OpenAIProviderConfig() { }

        public OpenAIProviderConfig(string providerName, string url, string key, string modelId)
        {
            ProviderName = providerName;
            Url = url;
            Key = key;
            ModelId = modelId;
        }
    }
}
