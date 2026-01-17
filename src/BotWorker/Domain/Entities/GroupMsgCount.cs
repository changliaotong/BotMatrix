using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("msgcount")]
    public class GroupMsgCount
    {
        [Key]
        public long Id { get; set; }
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = "";
        public long UserId { get; set; }
        public string UserName { get; set; } = "";
        public DateTime CDate { get; set; }
        public DateTime MsgDate { get; set; }
        public int CMsg { get; set; }
    }
}
