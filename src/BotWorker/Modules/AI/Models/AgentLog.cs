using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations.Schema;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Modules.AI.Models
{
    [Table("ai_usage_logs")]
    public class AgentLog
    {
        public long Id { get; set; }
        public long UserId { get; set; }
        public long AgentId { get; set; }
        public string? ModelName { get; set; }
        public int InputTokens { get; set; }
        public int OutputTokens { get; set; }
        public int DurationMs { get; set; }
        public string Status { get; set; } = "success";
        public string? ErrorMessage { get; set; }
        public string? Guid { get; set; }
        public string? GroupId { get; set; }
        public string? GroupName { get; set; }
        public string? UserName { get; set; }
        public string? MsgId { get; set; }
        public string? Question { get; set; }
        public string? Answer { get; set; }
        public string? Messages { get; set; }
        public decimal Credit { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    }
}
