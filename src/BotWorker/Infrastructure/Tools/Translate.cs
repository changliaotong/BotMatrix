using System.Text;
using Newtonsoft.Json;

namespace BotWorker.Infrastructure.Tools
{
    public class Translate : MetaData<Translate>
    {
        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();

        public static async Task<string> GetAzureResAsync(string text) 
        {
            string subscriptionKey = AzureTranslateSubscriptionKey;
            string endpoint = AzureTranslateEndpoint; 
            string location = AzureTranslateLocation;

            // 发起自动检测请求
            string detectRequestBody = JsonConvert.SerializeObject(new[] { new { Text = text } });
            string detectRequestUrl = $"{endpoint}/detect?api-version=3.0";

            using HttpClient client = new();
            client.DefaultRequestHeaders.Add("Ocp-Apim-Subscription-Key", subscriptionKey);
            client.DefaultRequestHeaders.Add("Ocp-Apim-Subscription-Region", location);

            var detectRequest = new HttpRequestMessage(HttpMethod.Post, detectRequestUrl)
            {
                Content = new StringContent(detectRequestBody, Encoding.UTF8, "application/json")
            };

            var detectResponse = await client.SendAsync(detectRequest);
            var detectResponseBody = await detectResponse.Content.ReadAsStringAsync();

            if (detectResponse.IsSuccessStatusCode)
            {
                // 解析语言检测响应
                var detectionResult = JsonConvert.DeserializeObject<DetectionResponse[]>(detectResponseBody);
                string detectedLanguage = detectionResult![0].Language ?? "";

                // 根据检测到的语言进行翻译
                string targetLanguage;
                if (detectedLanguage == "zh-Hans")
                {
                    // 中文检测为英文翻译
                    targetLanguage = "en";
                }
                else
                {
                    // 其他语言翻译为中文
                    targetLanguage = "zh-Hans";
                }

                // 发起翻译请求
                string translateRequestBody = JsonConvert.SerializeObject(new[] { new { Text = text } });
                string translateRequestUrl = $"{endpoint}/translate?api-version=3.0&to={targetLanguage}";

                var translateContent = new StringContent(translateRequestBody, Encoding.UTF8, "application/json");

                var translateResponse = await client.PostAsync(translateRequestUrl, translateContent);
                var translateResponseBody = await translateResponse.Content.ReadAsStringAsync();

                if (translateResponse.IsSuccessStatusCode)
                {
                    // 解析翻译响应
                    var translationResult = JsonConvert.DeserializeObject<TranslationResponse[]>(translateResponseBody);
                    string translatedText = translationResult![0].Translations![0].Text ?? "";
                    return translatedText;
                }
                else
                {
                    Console.WriteLine("翻译请求失败，错误代码：");
                    Console.WriteLine(translateResponse.StatusCode);
                }
            }
            else
            {
                Console.WriteLine("语言检测请求失败，错误代码：");
                Console.WriteLine(detectResponse.StatusCode);
            }
            return "";
        }    

        // 定义用于反序列化语言检测响应的类
        class DetectionResponse
        {
            public string? Language { get; set; }
            public float Score { get; set; }
        }

        // 定义用于反序列化翻译响应的类
        class TranslationResponse
        {
            public Translation[]? Translations { get; set; }
        }

        class Translation
        {
            public string? Text { get; set; }
            public string? To { get; set; }
        }
    }
}
