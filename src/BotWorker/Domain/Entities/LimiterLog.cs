namespace BotWorker.Domain.Entities
{
    public class LimiterLog
    {
        public int Id { get; set; }
        public long? GroupId { get; set; }  // NULL 表示私聊签到
        public long UserId { get; set; }
        public string ActionKey { get; set; } = default!;
        public DateTime UsedAt { get; set; }
    }
}
