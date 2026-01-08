using System.Net;
using System.Text;
using BotWorker.Common;
using BotWorker.Core.Logging;

namespace BotWorker.Infrastructure.Utils
{
    public static class HttpHelper
    {
        private static readonly HttpClient client = new();

        public static async Task<string> GetUrlDataAsync(this string? url)
        {
            try
            {
                return await client.GetStringAsync(url ?? "");
            }
            catch (Exception e)
            {
                Logger.Error($"\nException Caught!\nMessage :{e.Message}");
                return "";
            }
        }

        /// <summary>
        /// 使用 Squid 代理访问 URL（支持用户名+密码）
        /// </summary>
        /// <param name="url">目标 URL</param>
        /// <param name="proxyIP">Squid 服务器 IP</param>
        /// <param name="proxyPort">Squid 端口</param>
        /// <param name="proxyUser">代理用户名</param>
        /// <param name="proxyPass">代理密码</param>
        /// <param name="method">GET 或 POST</param>
        /// <param name="postData">POST 数据（可为空）</param>
        /// <returns>返回响应字符串</returns>
        public static async Task<string> SendRequestAsync(
            string url,
            string proxyIP,
            int proxyPort,
            string proxyUser,
            string proxyPass,
            string method = "GET",
            string postData = "")
        {
            var proxy = new WebProxy(proxyIP, proxyPort)
            {
                Credentials = new NetworkCredential(proxyUser, proxyPass)
            };

            var handler = new HttpClientHandler
            {
                Proxy = proxy,
                PreAuthenticate = true,
                UseDefaultCredentials = false
            };

            using var client = new HttpClient(handler);

            HttpResponseMessage response;

            if (method.ToUpper() == "POST" && postData != null)
            {
                var content = new StringContent(postData, Encoding.UTF8, "application/x-www-form-urlencoded");
                response = await client.PostAsync(url, content);
            }
            else
            {
                response = await client.GetAsync(url);
            }

            response.EnsureSuccessStatusCode();
            return await response.Content.ReadAsStringAsync();
        }
    }
}
