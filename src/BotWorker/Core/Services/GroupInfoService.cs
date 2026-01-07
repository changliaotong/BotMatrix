using BotWorker.Bots.Entries;
using BotWorker.Infrastructure.Caching;

namespace BotWorker.Core.Services
{
    public class GroupInfoService(ICacheRepository<GroupInfo> cache)
    {
        private readonly ICacheRepository<GroupInfo> _cache = cache;

        public async Task RemoveGroupInfoCacheAsync(string groupId)
        {
            await _cache.RemoveAsync(groupId);
        }
    }

}
