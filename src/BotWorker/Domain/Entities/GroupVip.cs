using System;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("VIP")]
    public class GroupVip
    {
        private static IGroupVipRepository Repository => 
            GlobalConfig.ServiceProvider?.GetRequiredService<IGroupVipRepository>() 
            ?? throw new InvalidOperationException("IGroupVipRepository not registered");

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

        public static async Task<int> BuyRobotAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy)
        {
            return await Repository.BuyRobotAsync(botUin, groupId, groupName, qqBuyer, buyerName, month, payMoney, payMethod, trade, memo, insertBy);
        }

        public static async Task<int> ChangeGroupAsync(long groupId, long newGroupId, long qq)
        {
            return await Repository.ChangeGroupAsync(groupId, newGroupId, qq);
        }

        public static async Task<int> RestDaysAsync(long groupId)
        {
            return await Repository.RestDaysAsync(groupId);
        }

        public static async Task<int> RestMonthsAsync(long groupId)
        {
            return await Repository.RestMonthsAsync(groupId);
        }

        public static async Task<bool> IsYearVIPAsync(long groupId)
        {
            return await Repository.IsYearVIPAsync(groupId);
        }

        public static async Task<bool> IsVipAsync(long groupId)
        {
            return await Repository.IsVipAsync(groupId);
        }

        public static async Task<bool> IsForeverAsync(long groupId)
        {
            return await Repository.IsForeverAsync(groupId);
        }

        public static async Task<bool> IsVipOnceAsync(long groupId)
        {
            return await Repository.IsVipOnceAsync(groupId);
        }

        public static async Task<bool> IsClientVipAsync(long qq)
        {
            return await Repository.IsClientVipAsync(qq);
        }
    }
}
