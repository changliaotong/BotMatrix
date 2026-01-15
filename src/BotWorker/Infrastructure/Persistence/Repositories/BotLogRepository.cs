using System;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotLogRepository : BaseRepository<BotLog>, IBotLogRepository
    {
        public BotLogRepository(string? connectionString = null) 
            : base("bot_logs", connectionString ?? GlobalConfig.LogConnection)
        {
        }

        public async Task<int> LogAsync(string info, string memo, BotMessage bm)
        {
            var properties = bm.GetType().GetProperties();
            var boMessage = string.Join(", ", properties.Select(p => $"{p.Name}: {p.GetValue(bm)}"));

            var botLog = new BotLog
            {
                MsgGuid = bm.MsgGuid,
                Info = info,
                Memo = memo,
                Json = boMessage,
                BotType = bm.Platform.ToString(),
                IsSignalR = bm.SelfInfo?.IsSignalR ?? false,
                EventType = bm.EventType,
                EventMessage = bm.EventMessage,
                BotQQ = bm.SelfId,
                BotName = bm.SelfName,
                GroupId = bm.GroupId,
                GroupName = bm.GroupName,
                GroupOpenid = bm.GroupOpenid,
                QQ = bm.UserId,
                Name = bm.Name,
                UserOpenid = bm.UserOpenId,
                MsgId = bm.MsgId,
                Message = bm.Message,
                Operater = bm.Operater,
                OperaterName = bm.OperaterName,
                InvitorQQ = bm.InvitorQQ,
                InvitorName = bm.InvitorName,
                Period = bm.Period,
                BotPerm = bm.SelfPerm,
                Perm = bm.UserPerm,
                IsAtMe = bm.IsAtMe,
                IsGroup = bm.IsGroup,
                GroupOwner = bm.Group?.RobotOwner ?? 0,
                IsCmd = bm.IsCmd,
                IsRefresh = bm.IsRefresh,
                RealGroupId = bm.GroupId,
                RealMessage = bm.Message,
                CmdName = bm.CmdName,
                CmdPara = bm.CmdPara,
                IsConfirm = bm.IsConfirm,
                AgentId = bm.AgentId,
                Context = bm.HistoryMessageCount,
                AgentName = bm.AgentName,
                InputTokens = bm.InputTokens,
                OutputTokens = bm.OutputTokens,
                TokensTimes = bm.TokensTimes,
                TokensTimesOutput = bm.TokensTimesOutput,
                TokensMinus = bm.TokensMinus,
                ModelId = bm.ModelId,
                IsDup = bm.IsDup,
                IsMusic = bm.IsMusic,
                AnswerId = bm.AnswerId,
                Answer = bm.Answer,
                IsAI = bm.IsAI,
                IsSend = bm.IsSend,
                IsRecall = bm.IsRecall,
                TargetUin = bm.TargetUin,
                Accept = bm.Accept,
                Reason = bm.Reason,
                IsSet = bm.IsSet
            };

            return (int)await InsertAsync(botLog);
        }
    }
}
