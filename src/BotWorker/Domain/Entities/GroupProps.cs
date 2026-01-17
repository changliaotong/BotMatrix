using System;
using System.Data;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("props")]
    public class GroupProps
    {
        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public long PropId { get; set; }
        public int IsUsed { get; set; }
        public DateTime? UsedDate { get; set; }
        public long? UsedUserId { get; set; }

        public const string PropClosed = "道具系统已关闭";
    }
}
