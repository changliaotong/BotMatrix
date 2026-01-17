using System;
using System.Collections.Generic;
using BotWorker.Common;

namespace BotWorker.Domain.Entities
{
    [Table("question")]
    public class QuestionInfo
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string Question { get; set; } = string.Empty;
        public int CUsed { get; set; }
        public int Audit2 { get; set; }
        public DateTime? Audit2Date { get; set; }
        public long Audit2By { get; set; }
        public bool IsSystem { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;
    }
}
