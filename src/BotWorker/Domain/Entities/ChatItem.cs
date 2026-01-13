namespace BotWorker.Domain.Entities
{
    public class ChatItem
    {
        public int Id { get; set; }
        public string AvatarUrl { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string LastMessage { get; set; } = string.Empty;
        public string Time { get; set; } = string.Empty;    
    }

}
