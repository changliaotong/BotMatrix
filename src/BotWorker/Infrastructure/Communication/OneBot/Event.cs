using System;
using System.Text.Json.Serialization;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Infrastructure.Communication.OneBot
{
    public abstract class EventBase
    {
        [JsonPropertyName("time")]
        [JsonConverter(typeof(FlexibleLongConverter))]
        public long Time { get; set; }

        [JsonPropertyName("self_id")]
        [JsonConverter(typeof(FlexibleLongConverter))]
        public long SelfId { get; set; }

        [JsonPropertyName("post_type")]
        public string PostType { get; set; } = string.Empty;

        public string Platform { get; set; } = string.Empty;

        public abstract string UserId { get; }
        public abstract string? GroupId { get; }
        public abstract string RawMessage { get; set; }
    }

    public class OneBotEvent : EventBase
    {
        [JsonPropertyName("message_type")]
        public string? MessageType { get; set; }

        [JsonPropertyName("sub_type")]
        public string? SubType { get; set; }

        [JsonPropertyName("message_id")]
        [JsonConverter(typeof(FlexibleLongConverter))]
        public long MessageId { get; set; }

        [JsonPropertyName("user_id")]
        [JsonConverter(typeof(FlexibleLongConverter))]
        public long UserIdLong { get; set; }

        [JsonPropertyName("group_id")]
        [JsonConverter(typeof(FlexibleLongConverter))]
        public long GroupIdLong { get; set; }

        public override string UserId => UserIdLong.ToString();
        public override string? GroupId => GroupIdLong == 0 ? null : GroupIdLong.ToString();

        [JsonPropertyName("raw_message")]
        public override string RawMessage { get; set; } = string.Empty;

        [JsonPropertyName("self_role")]
        public string? SelfRole { get; set; }

        [JsonPropertyName("sender")]
        public Sender? Sender { get; set; }
    }

    public class Sender
    {
        [JsonPropertyName("user_id")]
        public long UserId { get; set; }

        [JsonPropertyName("nickname")]
        public string? Nickname { get; set; }

        [JsonPropertyName("card")]
        public string? Card { get; set; }

        [JsonPropertyName("role")]
        public string? Role { get; set; }
    }

    public class BotMessageEvent : EventBase
    {
        private readonly BotMessage _msg;
        public BotMessage BotMessage => _msg;

        public BotMessageEvent(BotMessage msg)
        {
            _msg = msg;
            SelfId = msg.SelfId;
            Platform = msg.Platform;
            Time = DateTimeOffset.Now.ToUnixTimeSeconds();
            
            if (msg.EventType == "GroupMessageEvent" || msg.EventType == "FriendMessageEvent" || string.IsNullOrEmpty(msg.EventType))
            {
                PostType = "message";
            }
            else
            {
                PostType = msg.EventType;
            }
            RawMessage = msg.Message;
        }

        public override string UserId => _msg.UserId.ToString();
        public override string? GroupId => _msg.GroupId == 0 ? null : _msg.GroupId.ToString();
        public override string RawMessage 
        { 
            get => _msg.CurrentMessage; 
            set => _msg.CurrentMessage = value; 
        }
    }
}


