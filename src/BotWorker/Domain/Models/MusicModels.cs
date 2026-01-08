using Newtonsoft.Json;

namespace BotWorker.Domain.Models
{
    public class MsgMusic
    {
        [JsonProperty("type")] public string Type { get; set; } = "music";
        [JsonProperty("data")] public MsgData Data { get; set; } = new();

        public class MsgData
        {
            [JsonProperty("type")] public string Type { get; set; } = string.Empty;
            [JsonProperty("id")] public string Id { get; set; } = string.Empty;
            [JsonProperty("url")] public string Url { get; set; } = string.Empty;
            [JsonProperty("audio")] public string Audio { get; set; } = string.Empty;
            [JsonProperty("title")] public string Title { get; set; } = string.Empty;
            [JsonProperty("content")] public string Content { get; set; } = string.Empty;
            [JsonProperty("image")] public string Image { get; set; } = string.Empty;
        }
    }
}


