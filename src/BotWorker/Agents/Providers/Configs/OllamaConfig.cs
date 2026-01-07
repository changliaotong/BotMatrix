namespace sz84.Agents.Providers.Configs
{
    public class OllamaConfig(string modelId, string ollamaUrl)
    {
        public string ModelId { get; set; } = modelId;
        public string OllamaUrl { get; set; } = ollamaUrl;
    }

    public static class Ollama
    {
        public static string ModelId { get; } = "wangshenzhi/gemma2-9b-chinese-chat:latest";
        public static string OllamaUrl { get; } = "http://192.168.0.133:11434";
    }

}
