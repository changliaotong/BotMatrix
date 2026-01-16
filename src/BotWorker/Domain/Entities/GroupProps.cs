using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("props")]
    public class GroupProps
    {
        private static IGroupPropsRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupPropsRepository>() 
            ?? throw new InvalidOperationException("IGroupPropsRepository not registered");

        private static IPropRepository PropRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IPropRepository>() 
            ?? throw new InvalidOperationException("IPropRepository not registered");

        private static IUserRepository UserRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserRepository>() 
            ?? throw new InvalidOperationException("IUserRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public long PropId { get; set; }
        public int IsUsed { get; set; }
        public DateTime? UsedDate { get; set; }
        public long? UsedUserId { get; set; }

        public const string PropClosed = "道具系统已关闭";

        public static async Task<long> GetIdAsync(long groupId, long qq, long propId)
        {
            return await Repository.GetIdAsync(groupId, qq, propId);
        }

        public static async Task<bool> HavePropAsync(long groupId, long userId, long propId)
        {
            return await Repository.HavePropAsync(groupId, userId, propId);
        }

        public static async Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp)
        {
            return await Repository.UsePropAsync(groupId, userId, propId, qqProp);
        }

        public static async Task<string> GetMyPropListAsync(long groupId, long userId)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;
            return await Repository.GetMyPropListAsync(groupId, userId);
        }

        public static async Task<bool> IsClosedAsync(long groupId)
        {
            return !await GroupRepository.GetBoolAsync("IsProp", groupId);
        }

        public static string GetBuyRes(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
            => GetBuyResAsync(botUin, groupId, groupName, qq, name, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> GetBuyResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;

            if (string.IsNullOrEmpty(cmdPara) || cmdPara == "道具")
                return await Prop.GetPropListAsync();
            else
            {
                long prop_id = await Prop.GetIdAsync(cmdPara);
                if (prop_id != 0)
                {
                    long credit_value = await UserRepository.GetCreditAsync(botUin, groupId, qq);
                    int prop_price = await PropRepository.GetPropPriceAsync(prop_id);
                    if (credit_value < prop_price)
                        return $"您的积分{credit_value}不足{prop_price}";
                    
                    using var wrapper = await BotWorker.Infrastructure.Persistence.TransactionWrapper.BeginTransactionAsync();
                    try
                    {
                        // 1. 通用加积分函数 (含日志记录)
                        var res = await UserRepository.AddCreditAsync(botUin, groupId, groupName, qq, name, -prop_price, $"购买道具:{prop_id}", wrapper.Transaction);
                        if (!res.Success) throw new Exception("更新积分失败");

                        // 2. 插入道具购买记录
                        await Repository.InsertAsync(groupId, qq, prop_id, wrapper.Transaction);

                        await wrapper.CommitAsync();

                        await UserRepository.SyncCreditCacheAsync(botUin, groupId, qq, res.CreditValue);

                        return $"购买道具成功\n积分：-{prop_price}，累计：{res.CreditValue}";
                    }
                    catch (Exception ex)
                    {
                        await wrapper.RollbackAsync();
                        Logger.Error($"[GetBuyRes Error] {ex.Message}");
                        return "操作失败，请重试";
                    }
                }
                else
                    return "没有此道具";
            }
        }
    }

    [Table("prop")]
    public class Prop
    {
        private static IPropRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IPropRepository>() 
            ?? throw new InvalidOperationException("IPropRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        [Key]
        public long Id { get; set; }
        public string PropName { get; set; } = "";
        public int PropPrice { get; set; }
        public int IsValid { get; set; }

        public static async Task<long> GetIdAsync(string propName)
        {
            return await Repository.GetIdAsync(propName);
        }

        public static long GetId(string propName)
            => GetIdAsync(propName).GetAwaiter().GetResult();

        public static async Task<string> GetPropListAsync()
        {
            return await Repository.GetPropListAsync();
        }

        public static string GetPropList()
            => GetPropListAsync().GetAwaiter().GetResult();

        public static async Task<string> GetPropResAsync(long groupId)
        {
            bool is_prop = await GroupRepository.GetBoolAsync("IsProp", groupId);
            return is_prop 
                ? "道具系统\n可用道具：\n禁言卡\n飞机票\n免踢卡\n购买道具请发送【购买 + 道具名称】"
                : GroupProps.PropClosed;
        }

        public static string GetPropRes(long groupId)
            => GetPropResAsync(groupId).GetAwaiter().GetResult();
    }
}
