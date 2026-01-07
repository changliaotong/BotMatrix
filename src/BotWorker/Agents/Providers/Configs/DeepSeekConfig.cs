namespace BotWorker.Agents.Providers.Configs
{
    public static class DeepSeek
    {
        public static string Url => "https://api.deepseek.com";
        public static string Key => "sk-5506038d208349b2a7d05c8e76dea879";
        public static string ModelId => "deepseek-chat";
    }

    public class DeepSeekConfig(string url, string key, string modelId)
    {
        public string Url { get; set; } = url;
        public string Key { get; set; } = key;
        public string ModelId { get; set; } = modelId;
    }

    public static class DoubaoDeepSeek
    {
        private static readonly Random Random = new();

        public static string Url => "https://ark.cn-beijing.volces.com/api/v3/";
        public static string Key => "82358c14-543d-4e1d-8d0b-0a6ed38d9203";
        public static string Secret => "TldRNFptVTVObVF6WVdSak5EYzFZbUUxWXpNeU9EQTROVEk1TWpWaU1qUQ==";

        public static string ModelId => Models[Random.Next(Models.Length)];

        public static readonly string[] Models =
        {
                    "ep-20250223012927-sfmn9", //DeepSeek-R1-250120                   
            };
    }
}
