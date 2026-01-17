using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("send_message")]
    public class GroupSendMessage
    {
        [Key]
        public int Id { get; set; }
        
        public long MsgId { get; set; }
        public long? GroupId { get; set; }
        public long? ClientQQ { get; set; }
        public string? Question { get; set; }
        public string? AnswerAi { get; set; }
        public string? Message { get; set; }
        public DateTime? InsertDate { get; set; }

        // Additional fields from Cov list in Append
        public string? MsgGuid { get; set; }
        public long? BotUin { get; set; }
        public string? GroupName { get; set; }
        public long? UserId { get; set; }
        public string? UserName { get; set; }
        public long? AnswerId { get; set; }
        public bool? IsAI { get; set; }
        public string? AgentId { get; set; }
        public string? AgentName { get; set; }
        public bool? IsSend { get; set; }
        public bool? IsRealProxy { get; set; }
        public string? Reason { get; set; }
        public bool? IsCmd { get; set; }
        public int? InputTokens { get; set; }
        public int? OutputTokens { get; set; }
        public int? TokensMinus { get; set; }
        public bool? IsVoiceReply { get; set; }
        public string? VoiceName { get; set; }
        public int? CostTime { get; set; }
        public bool? IsRecall { get; set; }
        public int? ReCallAfterMs { get; set; }
    }
}
