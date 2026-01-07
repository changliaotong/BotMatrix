using System.Text.Encodings.Web;
using System.Text.Json;

namespace sz84.Infrastructure.Logging
{
    public class LoggingHandler(HttpMessageHandler innerHandler) : DelegatingHandler(innerHandler)
    {
        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            // 检查 request.Content 是否为 null，避免空引用问题
            if (request.Content != null)
            {
                string content = await request.Content.ReadAsStringAsync(cancellationToken);

                Console.OutputEncoding = System.Text.Encoding.UTF8; // 确保控制台支持中文

                Console.WriteLine("🔍 OpenAI Request Body:");

                try
                {
                    // 解析为 JsonDocument
                    using var jsonDoc = JsonDocument.Parse(content);

                    // 格式化序列化为带中文的漂亮字符串
                    var prettyJson = JsonSerializer.Serialize(jsonDoc.RootElement, new JsonSerializerOptions
                    {
                        WriteIndented = true,
                        Encoder = JavaScriptEncoder.UnsafeRelaxedJsonEscaping // 让序列化时不转义中文
                    });

                    Console.WriteLine(prettyJson);
                }
                catch (JsonException)
                {
                    // 解析失败就直接打印原始字符串
                    Console.WriteLine(content);
                }
            }
            else
            {
                Console.WriteLine("Request content is null.");
            }

            return await base.SendAsync(request, cancellationToken);
        }
    }
}
