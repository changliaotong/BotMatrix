namespace BotWorker.Modules.AI.Models
{
    public class ChatHistoryItem
    {
        public string Question { get; set; } = string.Empty;
        public string Answer { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
    }
}
