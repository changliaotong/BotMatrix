using System.Collections.Generic;
using Newtonsoft.Json;

namespace BotWorker.Domain.Models
{
    public class MessageSegment
    {
        [JsonProperty("type")]
        public string Type { get; set; } = string.Empty;
        [JsonProperty("data")]
        public object Data { get; set; } = new object();
    }

    public class TextSegmentData
    {
        [JsonProperty("text")]
        public string Text { get; set; } = string.Empty;
    }

    public class ImageSegmentData
    {
        [JsonProperty("file")]
        public string File { get; set; } = string.Empty;
        [JsonProperty("url")]
        public string? Url { get; set; }
    }

    public class OneBotMessage
    {
        [JsonProperty("id")]
        public string Id { get; set; } = string.Empty;
        [JsonProperty("time")]
        public long Time { get; set; }
        [JsonProperty("platform")]
        public string Platform { get; set; } = string.Empty;
        [JsonProperty("self_id")]
        public string SelfId { get; set; } = string.Empty;
        [JsonProperty("post_type")]
        public string PostType { get; set; } = string.Empty;
        [JsonProperty("message_type")]
        public string MessageType { get; set; } = string.Empty;
        [JsonProperty("sub_type")]
        public string SubType { get; set; } = string.Empty;
        [JsonProperty("user_id")]
        public string UserId { get; set; } = string.Empty;
        [JsonProperty("group_id")]
        public string GroupId { get; set; } = string.Empty;
        [JsonProperty("group_name")]
        public string GroupName { get; set; } = string.Empty;
        [JsonProperty("message")]
        public List<MessageSegment> Message { get; set; } = new List<MessageSegment>();
        [JsonProperty("raw_message")]
        public string RawMessage { get; set; } = string.Empty;
        [JsonProperty("sender_name")]
        public string SenderName { get; set; } = string.Empty;
        [JsonProperty("sender_card")]
        public string SenderCard { get; set; } = string.Empty;
        [JsonProperty("user_avatar")]
        public string UserAvatar { get; set; } = string.Empty;
        [JsonProperty("echo")]
        public string Echo { get; set; } = string.Empty;
        [JsonProperty("retcode")]
        public int Retcode { get; set; }
        [JsonProperty("msg")]
        public string Msg { get; set; } = string.Empty;
        [JsonProperty("meta_type")]
        public string MetaType { get; set; } = string.Empty;
        [JsonProperty("extras")]
        public Dictionary<string, object> Extras { get; set; } = new Dictionary<string, object>();
    }

    public class OneBotAction
    {
        [JsonProperty("action")]
        public string Action { get; set; } = string.Empty;
        [JsonProperty("params")]
        public Dictionary<string, object> Params { get; set; } = new Dictionary<string, object>();
        [JsonProperty("echo")]
        public string Echo { get; set; } = string.Empty;
        [JsonProperty("self_id")]
        public string? SelfId { get; set; }
        [JsonProperty("platform")]
        public string? Platform { get; set; }
    }
}


