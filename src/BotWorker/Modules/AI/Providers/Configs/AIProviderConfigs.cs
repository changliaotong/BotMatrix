namespace BotWorker.Modules.AI.Providers.Configs
{
    public static class DeepSeek
    {
        public static string Url => "https://api.deepseek.com";
        public static string Key => "sk-5506038d208349b2a7d05c8e76dea879";
        public static string ModelId => "deepseek-chat";
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

    public static class Doubao
    {
        private static readonly Random Random = new();

        public static string Url => "https://ark.cn-beijing.volces.com/api/v3/";
        public static string Key => "82358c14-543d-4e1d-8d0b-0a6ed38d9203";
        public static string Secret => "TldRNFptVTVObVF6WVdSak5EYzFZbUUxWXpNeU9EQTROVEk1TWpWaU1qUQ==";

        public static string ModelId => Models[Random.Next(Models.Length)];
        public static string ModelIdFunctionCall => ModelsFunctionCall[Random.Next(ModelsFunctionCall.Length)];

        public static readonly string[] Models =
        {
            "ep-20241204224159-ksfpb", //Doubao-lite-128k
            "ep-20250125002107-wlsq4", //Doubao-pro-128k-240628_0125
            "ep-20250125002533-t8sdq", //Doubao-1.5-pro-32k-250115_0125
            "ep-20250125002812-g26sh", //Doubao-1.5-vision-pro-32k-250115_0125
            "ep-20250125003016-zktzd", //Doubao-vision-lite-32k-241015
            "ep-20250125003155-wjk7k", //Doubao-vision-pro-32k-241028  
            "ep-20250125003308-nsttn", //Doubao-pro-256k-241115
            "ep-20250521112102-gm5q9", //Doubao-1.5-vision-pro-250328
            "doubao-1-5-lite-32k-250115",
            "doubao-seed-1-6-flash-250715",
            "doubao-seed-1-6-250615",
            "doubao-1-5-vision-pro-32k-250115",
            "doubao-seed-1-6-thinking-250715",
            "doubao-1-5-pro-256k-250115",
            "deepseek-r1-250528",
            "doubao-1.5-vision-lite-250315",
        };

        public static readonly string[] ModelsFunctionCall =
        {
            "ep-20241206163403-2dq77", //Doubao-pro-32k-functioncall-240815
            "doubao-seed-1-6-flash-250715",
            "doubao-seed-1-6-250615",
            "doubao-1-5-vision-lite-250315",
            "doubao-1-5-vision-pro-32k-250115",
            "doubao-seed-1-6-thinking-250715",
            "doubao-1-5-pro-256k-250115",
        };
    }

    public static class QWen
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

    public static class Ollama
    {
        public static string ModelId { get; } = "wangshenzhi/gemma2-9b-chinese-chat:latest";
        public static string OllamaUrl { get; } = "http://192.168.0.133:11434";
    }

    public static class AzureOpenAI
    {
        public static string DeploymentName => "gpt-4o-mini";
        public static string Endpoint => "https://east-us-derlin.openai.azure.com";
        public static string ApiKey => "sk-...";
    }

    public static class AzureDalle
    {
        public static string Endpoint => "https://australia-east-derlin.openai.azure.com/";
        public static string ApiKey => "190629909e64471f927ab52a1c3d6e76";
        public static string DeploymentName => "Dalle3";
    }

    public static class DoubaoTxt2Img
    {
        public static string Url => "https://visual.volcengineapi.com?Action=CVProcess&Version=2022-08-31";
        public static string Key => Doubao.Key;
        public static string Secret => Doubao.Secret;
    }
}
