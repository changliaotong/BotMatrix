namespace BotWorker.Modules.AI.Providers.Configs
{
    public class OpenAIConfig(string url, string key, string modelId)
    {
        public string Url { get; set; } = url;
        public string Key { get; set; } = key;
        public string ModelId { get; set; } = modelId;
    }
}
