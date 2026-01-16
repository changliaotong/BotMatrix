using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    #region 配对系统数据模型

    /// <summary>
    /// 用户社交资料
    /// </summary>
    [Table("UserPairingProfiles")]
    public class UserPairingProfile
    {
        private static IUserPairingProfileRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserPairingProfileRepository>() 
            ?? throw new InvalidOperationException("IUserPairingProfileRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Nickname { get; set; } = string.Empty;
        public string Gender { get; set; } = "未知"; // 男, 女, 隐藏
        public string Zodiac { get; set; } = "未知"; // 星座
        public string Intro { get; set; } = "这个人很懒，什么都没留下。";
        public bool IsLooking { get; set; } = true; // 是否正在寻找配对
        public DateTime LastActive { get; set; } = DateTime.Now;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public static async Task<UserPairingProfile?> GetByUserIdAsync(string userId)
        {
            return await Repository.GetByUserIdAsync(userId);
        }

        public static async Task<List<UserPairingProfile>> GetActiveSeekersAsync(int limit = 10)
        {
            return await Repository.GetActiveSeekersAsync(limit);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }
    }

    /// <summary>
    /// 配对记录 (CP记录)
    /// </summary>
    [Table("PairingRecords")]
    public class PairingRecord
    {
        private static IPairingRecordRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IPairingRecordRepository>() 
            ?? throw new InvalidOperationException("IPairingRecordRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string User1Id { get; set; } = string.Empty;
        public string User2Id { get; set; } = string.Empty;
        public string Status { get; set; } = "pairing"; // pairing (匹配中), coupled (已成对), broken (已解绑)
        public DateTime PairDate { get; set; } = DateTime.Now;

        public static async Task<PairingRecord?> GetCurrentPairAsync(string userId)
        {
            return await Repository.GetCurrentPairAsync(userId);
        }

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }
    }

    #endregion
}
