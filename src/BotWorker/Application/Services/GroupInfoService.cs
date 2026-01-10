using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Caching;

namespace BotWorker.Application.Services
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


