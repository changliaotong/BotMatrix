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

        public static async Task<long> AppendAsync(BotMessage bm)
        {
            await Agent.UsedTimesIncrementAsync(bm.AgentId);
            
            var log = new AgentLog
            {
                UserId = bm.UserId,
                AgentId = bm.AgentId,
                ModelName = bm.ModelId.ToString(), // Or resolve name if possible
                InputTokens = bm.InputTokens,
                OutputTokens = bm.OutputTokens,
                DurationMs = (int)((bm.CurrentStopwatch?.Elapsed.TotalSeconds ?? 0) * 1000),
                Status = string.IsNullOrEmpty(bm.AnswerAI) ? "failed" : "success",
                Guid = bm.MsgId, // Assuming MsgId can be used as Guid if not provided
                GroupId = bm.RealGroupId,
                GroupName = bm.GroupName,
                UserName = bm.Name,
                MsgId = bm.MsgId,
                Question = bm.Message,
                Answer = bm.AnswerAI,
                Messages = BotWorker.Infrastructure.Extensions.Text.JsonExtensions.ToJsonString(bm.History),
                Credit = (decimal)bm.TokensMinus,
                CreatedAt = DateTime.UtcNow
            };

            using var scope = LLMApp.ServiceProvider!.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentLogRepository>();
            return await repo.AddAsync(log);
        }

        public static long Append(BotMessage bm)
        {
            // Sync version - usually better to use Async but keeping for compatibility
            return AppendAsync(bm).GetAwaiter().GetResult();
        }
    }
}
