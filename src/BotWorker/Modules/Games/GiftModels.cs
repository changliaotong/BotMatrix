using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 礼物配置模型
    /// </summary>
    [Table("GiftStoreItem")]
    public class GiftStoreItem
    {
        private static IGiftStoreItemRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGiftStoreItemRepository>() 
            ?? throw new InvalidOperationException("IGiftStoreItemRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string GiftName { get; set; } = string.Empty;
        public long GiftCredit { get; set; }
        public string GiftUrl { get; set; } = string.Empty;
        public string GiftImage { get; set; } = string.Empty;
        public int GiftType { get; set; } // 1: 普通, 2: 高级
        public bool IsValid { get; set; }

        public static async Task<List<GiftStoreItem>> GetValidGiftsAsync()
        {
            return await Repository.GetValidGiftsAsync();
        }

        public static async Task<GiftStoreItem?> GetByNameAsync(string name)
        {
            return await Repository.GetByNameAsync(name);
        }

        public async Task<bool> InsertAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }
    }

    /// <summary>
    /// 用户背包模型
    /// </summary>
    [Table("GiftBackpack")]
    public class GiftBackpack
    {
        private static IGiftBackpackRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGiftBackpackRepository>() 
            ?? throw new InvalidOperationException("IGiftBackpackRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public long GiftId { get; set; } // 修改为 long 以匹配 GiftStoreItem.Id
        public int ItemCount { get; set; } // 重命名以避免与 MetaData.Count() 冲突

        public static async Task<List<GiftBackpack>> GetUserBackpackAsync(string userId)
        {
            return await Repository.GetUserBackpackAsync(userId);
        }

        public static async Task<GiftBackpack?> GetItemAsync(string userId, long giftId)
        {
            return await Repository.GetItemAsync(userId, giftId);
        }

        public async Task<bool> InsertAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }
    }

    /// <summary>
    /// 礼物赠送记录
    /// </summary>
    [Table("GiftLog")]
    public class GiftRecord
    {
        private static IGiftLogRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGiftLogRepository>() 
            ?? throw new InvalidOperationException("IGiftLogRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public string UserId { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
        public string GiftUserId { get; set; } = string.Empty;
        public string GiftUserName { get; set; } = string.Empty;
        public long GiftId { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public int GiftCount { get; set; }
        public long GiftCredit { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;

        public async Task<bool> InsertAsync(System.Data.IDbTransaction? trans = null)
        {
            // Note: Since GiftLog table might have different fields in different models, 
            // we should be careful. But for now we just follow the pattern.
            // We might need a separate repository if fields differ significantly.
            return await Repository.InsertAsync(new Gift.GiftLog {
                BotUin = this.BotUin,
                GroupId = this.GroupId,
                GroupName = this.GroupName,
                UserId = long.Parse(this.UserId),
                UserName = this.UserName,
                GiftUserId = long.Parse(this.GiftUserId),
                GiftUserName = this.GiftUserName,
                GiftId = this.GiftId,
                GiftName = this.GiftName,
                GiftCount = this.GiftCount,
                GiftCredit = this.GiftCredit,
                CreatedAt = this.InsertDate
            }, trans);
        }
    }
}
