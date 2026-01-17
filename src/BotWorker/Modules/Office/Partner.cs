using System.Text.Json.Serialization;

namespace BotWorker.Modules.Office
{
    public class Partner
    {
        public int Id { get; set; }
        public long UserId { get; set; }
        public long RefUserId { get; set; }
        public bool IsValid { get; set; } = true;
        public DateTime InsertDate { get; set; } = DateTime.Now;
    }
}
