namespace BotWorker.Domain.Entities
{
    public partial class ChatMessage
    {
        public string? ApiKey { get; set; } = "TOKEN:AFCDE195E9EE00DCFCB5E0ED44D129EB";
        public long RobotQQ { get; set; } = 1098299491;
        public long AgentId { get; set; } = 86;
        public string AgentName { get; set; } = string.Empty;
        public int Context { get; set; } = 2;
        public long GroupId { get; set; }
        public string? GroupName { get; set; } = string.Empty;
        public long QQ { get; set; }
        public string? Name { get; set; }
        public string? MsgId { get; set; } = string.Empty;
        public string? Message { get; set; }
        public bool IsRefresh { get; set; } = false;
        public bool IsAtme { get; set; } = false;
        public int RobotPerm { get; set; } = 2;
        public int ClientPerm { get; set; } = 2;
    }
}
