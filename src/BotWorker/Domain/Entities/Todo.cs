using System;
using BotWorker.Infrastructure.Utils.Schema.Attributes;

namespace BotWorker.Domain.Entities
{
    [Table("todo")]
    public class Todo
    {
        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public int TodoNo { get; set; }
        public string TodoTitle { get; set; } = "";
        public string Description { get; set; } = ""; // 详细描述
        public string Priority { get; set; } = "Medium"; // Low, Medium, High
        public int Progress { get; set; } = 0; // 进度 0-100
        public string Status { get; set; } = "Pending"; // Pending, InProgress, Completed
        public string Category { get; set; } = "Todo"; // Todo, Dev, Test
        public DateTime? DueDate { get; set; } // 截止日期
        public DateTime InsertDate { get; set; } = DateTime.Now;
    }
}
