using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("cmd")]
    public partial class BotCmd
    {
        [Key]
        public long Id { get; set; }
        public string CmdName { get; set; } = string.Empty;
        public string CmdText { get; set; } = string.Empty;
        public int IsClose { get; set; }
    }
}
