using Microsoft.CodeAnalysis;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Serilog;
using sz84.Core.Data;
using sz84.Core.Interfaces;


namespace sz84.Bots.Models.Limiter
{
    public class LimiterService(AppDbContext db) : ILimiter, IBotModule
    {
        private readonly AppDbContext _db = db;

        public IModuleMetadata Metadata => new LimiterMetadata();

        public async Task<bool> HasUsedAsync(long? groupId, long userId, string actionKey)
        {
            var record = await _db.LimiterLog
                .AsNoTracking()
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            return record != null && record.UsedAt.Date == DateTime.Today;
        }

        public async Task MarkUsedAsync(long? groupId, long userId, string actionKey)
        {
            var record = await _db.LimiterLog
                .FirstOrDefaultAsync(x => x.GroupId == groupId && x.UserId == userId && x.ActionKey == actionKey);

            if (record == null)
            {
                await _db.LimiterLog.AddAsync(new LimiterLog
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
            return await _db.LimiterLog
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

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<LimiterService>();
            InfoMessage($"✅ [{nameof(LimiterService)}] 注册成功");
        }

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            modelBuilder.Entity<LimiterLog>();
        }
    }
}
