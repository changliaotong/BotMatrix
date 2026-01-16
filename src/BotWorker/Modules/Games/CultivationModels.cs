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
    [Table("CultivationProfiles")]
    public class CultivationProfile
    {
        private static ICultivationProfileRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<ICultivationProfileRepository>() 
            ?? throw new InvalidOperationException("ICultivationProfileRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;

        public int Level { get; set; } = 1;

        public long Exp { get; set; } = 0;

        public long MaxExp { get; set; } = 100;

        public int CultivationSpeed { get; set; } = 10;

        public DateTime LastCultivateTime { get; set; } = DateTime.MinValue;

        public static async Task<CultivationProfile?> GetByUserIdAsync(string userId)
        {
            return await Repository.GetByUserIdAsync(userId);
        }

        public static async Task<List<CultivationProfile>> GetTopCultivatorsAsync(int limit = 10)
        {
            return await Repository.GetTopCultivatorsAsync(limit);
        }

        public async Task<bool> InsertAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }

        public async Task<bool> UpdateAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.UpdateAsync(this, trans);
        }

        public string GetStageName()
        {
            return Level switch
            {
                < 10 => "炼气期",
                < 20 => "筑基期",
                < 30 => "金丹期",
                < 40 => "元婴期",
                < 50 => "化神期",
                < 60 => "炼虚期",
                < 70 => "合体期",
                < 80 => "大乘期",
                < 90 => "渡劫期",
                _ => "飞升成仙"
            };
        }

        public string GetRankDescription()
        {
            int subLevel = (Level - 1) % 10 + 1;
            return $"{GetStageName()} {subLevel} 层";
        }
    }

    [Table("CultivationRecords")]
    public class CultivationRecord
    {
        private static ICultivationRecordRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<ICultivationRecordRepository>() 
            ?? throw new InvalidOperationException("ICultivationRecordRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;

        public string ActionType { get; set; } = string.Empty; // 修炼, 突破, 走火入魔

        public string Detail { get; set; } = string.Empty;

        public DateTime CreateTime { get; set; } = DateTime.Now;

        public async Task<bool> InsertAsync(System.Data.IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }
}
