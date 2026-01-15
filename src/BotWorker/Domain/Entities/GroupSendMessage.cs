using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("SendMessage")]
    public class GroupSendMessage
    {
        private static IGroupSendMessageRepository Repo => GlobalConfig.ServiceProvider!.GetRequiredService<IGroupSendMessageRepository>();

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

        public static async Task<int> UserCountAsync(long groupId)
        {
            return await Repo.UserCountAsync(groupId);
        }

        public static int UserCount(long groupId) => UserCountAsync(groupId).GetAwaiter().GetResult();

        public static async Task<int> AppendAsync(BotMessage bm)
        {
            if (bm.User.IsLog) BotLog.Log($"{bm.GroupName}({bm.GroupId}) {bm.Name}({bm.UserId}) {bm.EventType}：\n{bm.Message}", "处理后", bm);
            if (bm.IsBlackSystem && bm.EventType.In("EventPrivateMessage", "EventGroupMessage", "TempMessageEvent")) return 0;

            var entity = new GroupSendMessage
            {
                MsgGuid = bm.MsgGuid,
                BotUin = bm.SelfId,
                GroupId = bm.RealGroupId,
                GroupName = bm.RealGroupName,
                UserId = bm.UserId,
                UserName = bm.Name,
                MsgId = bm.MsgId,
                Question = bm.Message.IsNull() ? bm.EventType : bm.Message,
                Message = bm.IsSend && bm.IsRealProxy && !bm.IsAI && bm.AnswerId == 0 ? $"@{bm.Card.ReplaceInvalid().RemoveUserIds().ReplaceSensitive(Regexs.OfficalRejectWords)}:{bm.Answer}" : bm.Answer,
                AnswerAi = bm.AnswerAI,
                AnswerId = bm.AnswerId,
                IsAI = bm.IsAI,
                AgentId = bm.AgentId,
                AgentName = bm.AgentName,
                IsSend = bm.IsSend,
                IsRealProxy = bm.IsRealProxy,
                Reason = bm.Reason,
                IsCmd = bm.IsCmd,
                InputTokens = bm.InputTokens,
                OutputTokens = bm.OutputTokens,
                TokensMinus = bm.TokensMinus,
                IsVoiceReply = bm.IsVoiceReply,
                VoiceName = bm.VoiceName,
                CostTime = bm.CostTime,
                IsRecall = bm.IsRecall,
                ReCallAfterMs = bm.RecallAfterMs,
                InsertDate = DateTime.Now
            };

            return await Repo.AppendAsync(entity);
        }

        public static int Append(BotMessage bm) => AppendAsync(bm).GetAwaiter().GetResult();
    }
}
