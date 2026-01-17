using System;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Interfaces;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games.Gift
{
    [Table("group_member")]
    public class GroupGift
    {
        [ExplicitKey]
        public long GroupId { get; set; }
        [ExplicitKey]
        public long UserId { get; set; }
        public long FansValue { get; set; }

        public const string GiftFormat = "格式：赠送 + QQ + 礼物名 + 数量(默认1)\n例如：赠送 {客服QQ} 小心心 10";
    }
}
