namespace BotWorker.Modules.AI.Models
{
    public class UserMessage
    {
        public long MsgId { get; set; }
        public long RobotQQ { get; set; }
        public long GroupId { get; set; }
        public string? GroupName { get; set; }
        public Guid Guid { get; set; }
        public long QQ { get; set; }
        public string? ClientName { get; set; }
        public string? SendTime { get; set; }
        public bool IsAI { get; set; } = false;
        public string? Message { get; set; }
        public bool IsCurr { get; set; }
    }


}
