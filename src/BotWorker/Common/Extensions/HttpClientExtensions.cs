using System.Net.Http.Headers;

namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class HttpClientExtensions
    {
        // ����Get���󲢷����ַ����������Ĭ�ϳ�ʱ��
        public static async Task<string> GetStringAsync(this HttpClient client, string url, int timeoutSeconds = 30)
        {
            using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(timeoutSeconds));
            return await client.GetStringAsync(url, cts.Token);
        }

        // Post Json ���󣬷��ط����л����
        public static async Task<T?> PostJsonAsync<T>(this HttpClient client, string url, object data)
        {
            var json = data.ToJson(); // ����֮ǰ����� ToJson ��չ
            using var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
            using var response = await client.PostAsync(url, content);
            response.EnsureSuccessStatusCode();
            var respJson = await response.Content.ReadAsStringAsync();
            return respJson.FromJson<T>(); // ����֮ǰ����� FromJson ��չ
        }

        // ��� Bearer Token ͷ
        public static void AddBearerToken(this HttpClient client, string token)
        {
            client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", token);
        }
    }
}


