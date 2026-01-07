using System.Net.Http.Headers;

namespace BotWorker.Common.Exts
{
    public static class HttpClientExtensions
    {
        // 发送Get请求并返回字符串结果（带默认超时）
        public static async Task<string> GetStringAsync(this HttpClient client, string url, int timeoutSeconds = 30)
        {
            using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(timeoutSeconds));
            return await client.GetStringAsync(url, cts.Token);
        }

        // Post Json 对象，返回反序列化结果
        public static async Task<T?> PostJsonAsync<T>(this HttpClient client, string url, object data)
        {
            var json = data.ToJson(); // 调用之前定义的 ToJson 扩展
            using var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
            using var response = await client.PostAsync(url, content);
            response.EnsureSuccessStatusCode();
            var respJson = await response.Content.ReadAsStringAsync();
            return respJson.FromJson<T>(); // 调用之前定义的 FromJson 扩展
        }

        // 添加 Bearer Token 头
        public static void AddBearerToken(this HttpClient client, string token)
        {
            client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", token);
        }
    }
}
