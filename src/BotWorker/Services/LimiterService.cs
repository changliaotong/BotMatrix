using Microsoft.EntityFrameworkCore;
using BotWorker.Core.OneBot;

namespace BotWorker.Services
{
    public interface ILimiter
    {
        Task<bool> HasUsedAsync(string? groupId, string userId, string actionKey);
        Task MarkUsedAsync(string? groupId, string userId, string actionKey);
        Task<DateTime?> GetLastUsedAsync(string? groupId, string userId, string actionKey);
        Task<bool> TryUseAsync(string? groupId, string userId, string actionKey);
    }

    public class LimiterService(BotDbContext db) : ILimiter
    {
        private readonly BotDbContext _db = db;

        public async Task<bool> HasUsedAsync(string? groupId, string userId, string actionKey)
        {
            var record = await _db.LimiterLogs
                .AsNoTracking()
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            return record != null && record.UsedAt.Date == DateTime.Today;
        }

        public async Task MarkUsedAsync(string? groupId, string userId, string actionKey)
        {
            var record = await _db.LimiterLogs
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            if (record == null)
            {
                await _db.LimiterLogs.AddAsync(new LimiterLog
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

        public async Task<DateTime?> GetLastUsedAsync(string? groupId, string userId, string actionKey)
        {
            return await _db.LimiterLogs
                .AsNoTracking()
                .Where(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey)
                .Select(x => (DateTime?)x.UsedAt)
                .FirstOrDefaultAsync();
        }

        public async Task<bool> TryUseAsync(string? groupId, string userId, string actionKey)
        {
            if (await HasUsedAsync(groupId, userId, actionKey))
                return false;

            await MarkUsedAsync(groupId, userId, actionKey);
            return true;
        }
    }

    public class LimiterLog
    {
        public int Id { get; set; }
        public string? GroupId { get; set; }
        public string UserId { get; set; } = default!;
        public string ActionKey { get; set; } = default!;
        public DateTime UsedAt { get; set; }
    }
}
