using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class GuildEvent : MetaData<GuildEvent>
    {
        public override string TableName => "GuildEvent";

        public override string KeyField => "Id";

        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long BotUin { get; set; }
        public string BotName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string EventType { get; set; } = string.Empty;
        public string EventName { get; set; } = string.Empty;
        public string EventInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static async Task<Dictionary<string, object>> AppendAsync(GuildEvent @event, params string[] fields)
        {
            return await InsertReturnFieldsAsync(new
            {
                @event.GroupId,
                @event.GroupName,
                @event.BotUin,
                @event.BotName,
                @event.UserId,
                @event.UserName,
                @event.EventType,
                @event.EventName,
                @event.EventInfo,               
  
            }, fields);
        }
    }
}
