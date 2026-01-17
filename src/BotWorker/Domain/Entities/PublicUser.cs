using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("public_user")]
    public class PublicUser
    {
        [Key]
        public long Id { get; set; }
        public string BotKey { get; set; } = string.Empty;
        public string UserKey { get; set; } = string.Empty;
        public long UserId { get; set; }
        public bool IsBind { get; set; }
        public DateTime? BindDate { get; set; }
        public string BindToken { get; set; } = string.Empty;
        public long BindCredit { get; set; }
        public long RecUserId { get; set; }
        public DateTime InsertDate { get; set; }
    }
}
