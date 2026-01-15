using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Games;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class FishingUserRepository : BaseRepository<FishingUser>, IFishingUserRepository
    {
        protected override string KeyField => "user_id";

        public FishingUserRepository(string? connectionString = null)
            : base("fishing_user", connectionString ?? GlobalConfig.KnowledgeBaseConnection)
        {
        }

        public async Task UpdateStateAsync(long userId, int state, int waitMinutes)
        {
            var user = await GetByIdAsync(userId);
            if (user != null)
            {
                user.State = state;
                user.WaitMinutes = waitMinutes;
                user.LastActionTime = DateTime.Now;
                await UpdateEntityAsync(user);
            }
        }

        public async Task UpdateStateAsync(long userId, int state)
        {
            var user = await GetByIdAsync(userId);
            if (user != null)
            {
                user.State = state;
                await UpdateEntityAsync(user);
            }
        }

        public async Task AddExpAndResetStateAsync(long userId, int exp)
        {
            var user = await GetByIdAsync(userId);
            if (user != null)
            {
                user.State = 0;
                user.Exp += exp;
                await UpdateEntityAsync(user);
            }
        }

        public async Task UpgradeRodAsync(long userId, long cost)
        {
            var user = await GetByIdAsync(userId);
            if (user != null)
            {
                user.Gold -= cost;
                user.RodLevel++;
                await UpdateEntityAsync(user);
            }
        }

        public async Task SellFishAsync(long userId, long totalGold)
        {
            using var wrapper = await BeginTransactionAsync();
            try
            {
                var user = await GetByIdAsync(userId);
                if (user != null)
                {
                    user.Gold += totalGold;
                    await UpdateEntityAsync(user, wrapper.Transaction);
                }
                
                // For bulk delete, raw SQL is still most efficient
                await wrapper.Transaction.Connection.ExecuteAsync(
                    "DELETE FROM fishing_bag WHERE user_id = @userId", 
                    new { userId }, 
                    wrapper.Transaction);
                
                await wrapper.CommitAsync();
            }
            catch
            {
                await wrapper.RollbackAsync();
                throw;
            }
        }
    }

    public class FishingBagRepository : BaseRepository<FishingBag>, IFishingBagRepository
    {
        public FishingBagRepository(string? connectionString = null)
            : base("fishing_bag", connectionString ?? GlobalConfig.KnowledgeBaseConnection)
        {
        }

        public async Task<IEnumerable<FishingBag>> GetByUserIdAsync(long userId, int limit)
        {
            return await GetListAsync($"WHERE user_id = @userId ORDER BY catch_time DESC LIMIT {limit}", new { userId });
        }

        public async Task<IEnumerable<FishingBag>> GetAllByUserIdAsync(long userId)
        {
            return await GetListAsync("WHERE user_id = @userId", new { userId });
        }
    }
}
