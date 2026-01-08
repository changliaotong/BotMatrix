using System.Text.Json;

namespace BotWorker.Common.Extensions
{
    public static class JsonExtensions
    {
        // JSON字符串格式化输出
        public static string JsonPrettyPrint(this string json)
        {
            try
            {
                using var doc = JsonDocument.Parse(json);
                return JsonSerializer.Serialize(doc.RootElement, new JsonSerializerOptions { WriteIndented = true });
            }
            catch
            {
                return json;
            }
        }
    }
}


