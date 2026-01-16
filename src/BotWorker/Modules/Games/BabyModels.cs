using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    [Table("Babies")]
    public class Baby
    {
        private static IBabyRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBabyRepository>() 
            ?? throw new InvalidOperationException("IBabyRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public DateTime Birthday { get; set; } = DateTime.Now;
        public int GrowthValue { get; set; } = 0;
        public int DaysOld { get; set; } = 0;
        public int Level { get; set; } = 1;
        public int Points { get; set; } = 0;
        public string Status { get; set; } = "active"; // active, abandoned
        public DateTime CreatedAt { get; set; } = DateTime.Now;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;
        public DateTime LastDailyUpdate { get; set; } = DateTime.MinValue;

        public static async Task<Baby?> GetByUserIdAsync(string userId)
        {
            return await Repository.GetByUserIdAsync(userId);
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

    [Table("BabyEvents")]
    public class BabyEvent
    {
        private static IBabyEventRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBabyEventRepository>() 
            ?? throw new InvalidOperationException("IBabyEventRepository not registered");

        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid BabyId { get; set; }
        public string EventType { get; set; } = string.Empty; // birthday, learn, work, interact
        public string Content { get; set; } = string.Empty;
        public DateTime CreatedAt { get; set; } = DateTime.Now;

        public async Task<bool> InsertAsync(IDbTransaction? trans = null)
        {
            return await Repository.InsertAsync(this, trans);
        }
    }

    [Table("BabyConfig")]
    public class BabyConfig
    {
        private static IBabyConfigRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBabyConfigRepository>() 
            ?? throw new InvalidOperationException("IBabyConfigRepository not registered");

        [ExplicitKey]
        public int Id { get; set; } = 1;
        public bool IsEnabled { get; set; } = true;
        public int GrowthRate { get; set; } = 1000;
        public DateTime UpdatedAt { get; set; } = DateTime.Now;

        public static async Task<BabyConfig> GetAsync()
        {
            return await Repository.GetAsync();
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
}
