using Newtonsoft.Json;

namespace BotWorker.Domain.Models
{
    public record MusicShareMessage
    {
        [JsonProperty("kind")] public string Kind { get; set; } = string.Empty;
        [JsonProperty("title")] public string Title { get; set; } = string.Empty;
        [JsonProperty("summary")] public string Summary { get; set; } = string.Empty;
        [JsonProperty("jumpUrl")] public string JumpUrl { get; set; } = string.Empty;
        [JsonProperty("pictureUrl")] public string PictureUrl { get; set; } = string.Empty;
        [JsonProperty("musicUrl")] public string MusicUrl { get; set; } = string.Empty;
        [JsonProperty("brief")] public string Brief { get; set; } = string.Empty;
    }

    public record AppMessage
    {
        [JsonProperty("type")] public string Type { get; set; } = "app";
        [JsonProperty("content")] public string Content { get; set; } = string.Empty;
    }
}


