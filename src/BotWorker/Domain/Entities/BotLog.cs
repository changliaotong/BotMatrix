using System;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("bot_logs")]
    public class BotLog
    {
        [Key]
        public long Id { get; set; }
        public string MsgGuid { get; set; } = string.Empty;
        public string Info { get; set; } = string.Empty;
        public string Memo { get; set; } = string.Empty;
        public string Json { get; set; } = string.Empty;
        public string BotType { get; set; } = string.Empty;
        public bool IsSignalR { get; set; }
        public string EventType { get; set; } = string.Empty;
        public string EventMessage { get; set; } = string.Empty;
        public long BotQQ { get; set; }
        public string BotName { get; set; } = string.Empty;
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public string GroupOpenid { get; set; } = string.Empty;
        public long QQ { get; set; }
        public string Name { get; set; } = string.Empty;
        public string UserOpenid { get; set; } = string.Empty;
        public string MsgId { get; set; } = string.Empty;
        public string Message { get; set; } = string.Empty;
        public long Operater { get; set; }
        public string OperaterName { get; set; } = string.Empty;
        public long InvitorQQ { get; set; }
        public string InvitorName { get; set; } = string.Empty;
        public string Period { get; set; } = string.Empty;
        public int BotPerm { get; set; }
        public int Perm { get; set; }
        public bool IsAtMe { get; set; }
        public bool IsGroup { get; set; }
        public long GroupOwner { get; set; }
        public bool IsCmd { get; set; }
        public bool IsRefresh { get; set; }
        public long RealGroupId { get; set; }
        public string RealMessage { get; set; } = string.Empty;
        public string CmdName { get; set; } = string.Empty;
        public string CmdPara { get; set; } = string.Empty;
        public bool IsConfirm { get; set; }
        public string AgentId { get; set; } = string.Empty;
        public int Context { get; set; }
        public string AgentName { get; set; } = string.Empty;
        public int InputTokens { get; set; }
        public int OutputTokens { get; set; }
        public int TokensTimes { get; set; }
        public int TokensTimesOutput { get; set; }
        public int TokensMinus { get; set; }
        public string ModelId { get; set; } = string.Empty;
        public bool IsDup { get; set; }
        public bool IsMusic { get; set; }
        public string AnswerId { get; set; } = string.Empty;
        public string Answer { get; set; } = string.Empty;
        public bool IsAI { get; set; }
        public bool IsSend { get; set; }
        public bool IsRecall { get; set; }
        public long TargetUin { get; set; }
        public bool Accept { get; set; }
        public string Reason { get; set; } = string.Empty;
        public bool IsSet { get; set; }
    }
}
