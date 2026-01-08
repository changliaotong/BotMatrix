using System;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.EntityFrameworkCore;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Infrastructure.Persistence;

namespace BotWorker.Application.Services
{
    public class LimiterService(AppDbContext db) : ILimiter
    {
        private readonly AppDbContext _db = db;

        public async Task<bool> HasUsedAsync(long? groupId, long userId, string actionKey)
        {
            var record = await _db.DailyLimitLogs
                .AsNoTracking()
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            return record != null && record.UsedAt.Date == DateTime.Today;
        }

        public async Task MarkUsedAsync(long? groupId, long userId, string actionKey)
        {
            var record = await _db.DailyLimitLogs
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            if (record == null)
            {
                await _db.DailyLimitLogs.AddAsync(new LimiterLog
                {
                    GroupId = groupId,
                    UserId = userId,
                    ActionKey = actionKey,
                    UsedAt = DateTime.Now
                });
            }
            else
            {
                record.UsedAt = DateTime.Now;
            }

            await _db.SaveChangesAsync();
        }

        public async Task<DateTime?> GetLastUsedAsync(long? groupId, long userId, string actionKey)
        {
            return await _db.DailyLimitLogs
                .AsNoTracking()
                .Where(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey)
                .Select(x => (DateTime?)x.UsedAt)
                .FirstOrDefaultAsync();
        }

        public async Task<bool> TryUseAsync(long? groupId, long userId, string actionKey)
        {
            if (await HasUsedAsync(groupId, userId, actionKey))
                return false;

            await MarkUsedAsync(groupId, userId, actionKey);
            return true;
        }
    }
}


