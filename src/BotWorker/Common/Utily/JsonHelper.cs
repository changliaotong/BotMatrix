using Newtonsoft.Json;

namespace BotWorker.Common.Utily
{
    public static class JsonHelper
    {
        public static readonly JsonSerializerSettings Settings = new()
        {
            TypeNameHandling = TypeNameHandling.Auto,
            Formatting = Formatting.Indented
        };

        public static string Serialize<T>(T obj) =>
            JsonConvert.SerializeObject(obj, Settings);

        public static T Deserialize<T>(string json) =>
            JsonConvert.DeserializeObject<T>(json, Settings)!;
    }


}
