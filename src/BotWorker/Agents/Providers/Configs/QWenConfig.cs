namespace BotWorker.Agents.Providers.Configs
{
    public class QWenConfig(string url, string key, string modelId)
    {
        //
        public string Url { get; set; } = url;
        public string Key { get; set; } = key;
        public string ModelId { get; set; } = modelId;
    }

    public class QWen
    {
        private static readonly Random Random = new();

        public static string Url => "https://dashscope.aliyuncs.com/compatible-mode/v1";
        public static string Key => "sk-c34939e80b5046abbe4b471b09b9fe43";

        public static string ModelId => Models[Random.Next(Models.Length)];
        public static string ModelIdFunctionCall => ModelsFunctionCall[Random.Next(ModelsFunctionCall.Length)];

        public static readonly string[] Models =
        {
                    "qwen-max",
                    "qwen-turbo",
                    "qwen-plus-latest",
                    "qwen-turbo-latest",
            };

        public static readonly string[] ModelsFunctionCall =
        {
                    "qwen-max",
                    "qwen-turbo",
            };
    }
}
