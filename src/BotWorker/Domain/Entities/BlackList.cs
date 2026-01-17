using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class BlackList
    {
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public long BlackId { get; set; }
        public string BlackInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public const string regexBlack = @"^(?<cmdName>(取消|解除|删除)?(黑名单|拉黑|加黑|删黑))(?<cmdPara>([ ]*(\[?@:?)?[1-9]+\d*(\]?))+)$";
    }
}
