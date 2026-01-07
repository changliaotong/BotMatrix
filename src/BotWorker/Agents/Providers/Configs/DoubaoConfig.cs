namespace BotWorker.Agents.Providers.Configs
{
    public class DoubaoConfig(string url, string key, string modelId)
    {
        public string Url { get; set; } = url;
        public string Key { get; set; } = key;
        public string ModelId { get; set; } = modelId;
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
                    "doubao-1-5-lite-32k-250115", //doubao-1-5-lite-32k-250115
                    "doubao-seed-1-6-flash-250715", //doubao-seed-1-6-flash-250715
                    "doubao-seed-1-6-250615", //doubao-seed-1-6-250615
                    "doubao-1-5-vision-pro-32k-250115", //doubao-1-5-vision-pro-32k-250115
                    "doubao-seed-1-6-thinking-250715", //doubao-seed-1-6-thinking-250715
                    "doubao-1-5-pro-256k-250115", //doubao-1-5-pro-256k-250115
                    "deepseek-r1-250528", //deepseek-r1-250528
                    "doubao-1-5-pro-256k-250115", //doubao-1-5-pro-256k-250115
                    "doubao-1.5-vision-lite-250315", //doubao-1.5-vision-lite-250315
            };

        public static readonly string[] ModelsFunctionCall =
        {
                    "ep-20241206163403-2dq77", //Doubao-pro-32k-functioncall-240815
                    "doubao-seed-1-6-flash-250715", //doubao-seed-1-6-flash-250715
                    "doubao-seed-1-6-250615", //doubao-seed-1-6-250615
                    "doubao-1.5-vision-lite-250315", //doubao-1.5-vision-lite-250315
                    "doubao-1-5-vision-pro-32k-250115", //doubao-1-5-vision-pro-32k-250115
                    "doubao-seed-1-6-thinking-250715", //doubao-seed-1-6-thinking-250715
                    "doubao-1-5-pro-256k-250115", //doubao-1-5-pro-256k-250115
                    //"deepseek-r1-250528", //deepseek-r1-250528
            };
    }
}
