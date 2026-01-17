namespace BotWorker.Domain.Entities
{
    public class WhiteList
    {
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long WhiteId { get; set; }
        public DateTime InsertDate { get; set; }
    }
}
