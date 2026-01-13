using Newtonsoft.Json;

namespace BotWorker.Models
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

        public static MusicShareMessage FromMirai(Mirai.Net.Data.Messages.Concretes.MusicShareMessage m)
        {
            return new MusicShareMessage
            {
                Kind = m.Kind,
                Title = m.Title,
                Summary = m.Summary,
                JumpUrl = m.JumpUrl,
                PictureUrl = m.PictureUrl,
                MusicUrl = m.MusicUrl,
                Brief = m.Brief
            };
        }
    }

    public record AppMessage
    {
        [JsonProperty("type")] public string Type { get; set; } = "app";
        [JsonProperty("content")] public string Content { get; set; } = string.Empty;
    }
}
