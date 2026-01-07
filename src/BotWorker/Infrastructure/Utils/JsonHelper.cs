using System.Text.Json.Serialization;
using System.Text.Json;

namespace sz84.Infrastructure.Utils
{
    public static class JsonHelper
    {
        private static readonly JsonSerializerOptions Options = new()
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
            WriteIndented = false,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        };

        public static string Serialize<T>(T obj) => JsonSerializer.Serialize(obj, Options);
        public static T? Deserialize<T>(string json) => JsonSerializer.Deserialize<T>(json, Options);
    }

}
