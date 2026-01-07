namespace BotWorker.Agents.Entries
{
    public class UserMessage
    {
        public long MsgId { get; set; }
        public long RobotQQ { get; set; }
        public long GroupId { get; set; }
        public string? GroupName { get; set; }
        public Guid Guid => Agent.GetGuid(GroupId == 10084 ? 86 : GroupId - 9900000000);
        public long QQ { get; set; }
        public string? ClientName { get; set; }
        public string? SendTime { get; set; }
        public bool IsAI { get; set; } = false;
        public string? Message { get; set; }
        public bool IsCurr { get; set; }
    }


}
