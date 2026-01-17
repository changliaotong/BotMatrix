using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("vip")]
    public class GroupVip
    {
        [ExplicitKey]
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public decimal FirstPay { get; set; }
        public DateTime StartDate { get; set; }
        public DateTime EndDate { get; set; }
        public string VipInfo { get; set; } = string.Empty;
        public long UserId { get; set; }
        public decimal IncomeDay { get; set; }
        public int IsYearVip { get; set; }
        public int InsertBy { get; set; }
        public int? IsGoon { get; set; }

        public static async Task<int> BuyRobotAsync(BotMessage bm, long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy)
        {
            return await bm.GroupVipRepository.BuyRobotAsync(botUin, groupId, groupName, qqBuyer, buyerName, month, payMoney, payMethod, trade, memo, insertBy);
        }

        public static async Task<int> ChangeGroupAsync(BotMessage bm, long groupId, long newGroupId, long qq)
        {
            return await bm.GroupVipRepository.ChangeGroupAsync(groupId, newGroupId, qq);
        }

        public static async Task<int> RestDaysAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.RestDaysAsync(groupId);
        }

        public static async Task<int> RestMonthsAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.RestMonthsAsync(groupId);
        }

        public static async Task<bool> IsYearVIPAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.IsYearVIPAsync(groupId);
        }

        public static async Task<bool> IsVipAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.IsVipAsync(groupId);
        }

        public static async Task<bool> IsForeverAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.IsForeverAsync(groupId);
        }

        public static async Task<bool> IsVipOnceAsync(BotMessage bm, long groupId)
        {
            return await bm.GroupVipRepository.IsVipOnceAsync(groupId);
        }

        public static async Task<bool> IsClientVipAsync(BotMessage bm, long qq)
        {
            return await bm.GroupVipRepository.IsClientVipAsync(qq);
        }
    }
}
